package monitor

import (
	"fmt"
	"sync"
	"time"
	"w3chttpd"
)

type AlertStatus int

const (
	StatusRecovered AlertStatus = iota
	StatusExceed    AlertStatus = iota
)

type Alert struct {
	Timestamp time.Time
	Total     int
	Status    AlertStatus
}

func (a *Alert) String() string {

	switch a.Status {

	case StatusExceed:
		return fmt.Sprintf(
			"High traffic generated an alert - hits = %d, triggered at %s",
			a.Total, a.Timestamp.Format("02/01/2006:15:04:05"))

	case StatusRecovered:
		return fmt.Sprintf("Traffic recovered at %s",
			a.Timestamp.Format("02/01/2006:15:04:05"))

	default:
		panic("Wrong alert status!")
	}
}

type window struct {
	queue *entryQueue

	// Left element closest to the window
	// elt1 elt2 elt3 edge [window] ...
	edge    *w3chttpd.Entry
	edgePos int

	lastProcessed    *w3chttpd.Entry
	lastProcessedPos int
	size             int
	trafficWindow    time.Duration
	status           AlertStatus
	threshold        int
	alerts           []*Alert
}

func (w *window) init(tw time.Duration, th int, poolSize int) {

	w.queue = &entryQueue{
		&sync.RWMutex{},
		make([]*w3chttpd.Entry, 0),
		&entryPool{},
	}

	w.queue.epool.init(poolSize)

	w.edge = nil
	w.edgePos = -1
	w.lastProcessed = nil
	w.lastProcessedPos = -1
	w.size = 0
	w.trafficWindow = tw
	w.status = StatusRecovered
	w.threshold = th
	w.alerts = nil
}

func (w *window) updateEdge(timestamp time.Time) {

	if len(w.queue.entries) == 0 {

		w.edge = nil
		w.edgePos = -1
		w.size = 0
		return
	}

	for _, we := range w.queue.entries[w.edgePos+1:] {

		if timestamp.Sub(we.Timestamp) >= w.trafficWindow {

			w.size -= we.Size
			w.edge = we
			w.edgePos++

		} else {
			break
		}
	}
}

func (w *window) getNextEntryToProcess() *w3chttpd.Entry {

	if w.lastProcessedPos+1 < len(w.queue.entries) {
		return w.queue.entries[w.lastProcessedPos+1]
	}
	return nil
}

func (w *window) processStatusExceedForEntry(e *w3chttpd.Entry) {

	w.updateEdge(e.Timestamp)

	if w.edge == nil {
		return
	}

	delta := e.Timestamp.Sub(w.edge.Timestamp)

	if delta >= w.trafficWindow && w.size < w.threshold {

		w.status = StatusRecovered
		w.alerts = append(w.alerts, &Alert{
			Timestamp: e.Timestamp,
			Total:     w.size,
			Status:    w.status,
		})
	}
}

func (w *window) processStatusRecoveredForEntry(e *w3chttpd.Entry) {

	w.updateEdge(e.Timestamp)

	if w.size > w.threshold {

		w.status = StatusExceed
		w.alerts = append(w.alerts, &Alert{
			Timestamp: e.Timestamp,
			Total:     w.size,
			Status:    w.status,
		})
	}
}

func (w *window) getNewAlerts(end time.Time, deleted int) []*Alert {

	w.lastProcessedPos -= deleted
	w.edgePos -= deleted
	w.alerts = []*Alert{}

	e := w.getNextEntryToProcess()

	for e != nil && !e.Timestamp.After(end) {

		if w.status == StatusExceed {

			previousSec := e.Timestamp.Add(-time.Second)
			w.processStatusExceedForTime(previousSec)
		}

		w.size += e.Size
		w.lastProcessedPos++

		next := w.getNextEntryToProcess()
		for next != nil && next.Timestamp == e.Timestamp {

			w.size += next.Size
			w.lastProcessedPos++
			next = w.getNextEntryToProcess()
		}

		w.lastProcessed = w.queue.entries[w.lastProcessedPos]

		switch w.status {

		case StatusRecovered:
			w.processStatusRecoveredForEntry(e)

		case StatusExceed:
			w.processStatusExceedForEntry(e)
		}

		e = w.getNextEntryToProcess()
	}

	if w.status == StatusExceed {
		w.processStatusExceedForTime(end.Add(-time.Second))
	}

	return w.alerts
}

func (w *window) processStatusExceedForTime(t time.Time) {

	pos := w.edgePos + 1

	if t.Sub(w.queue.entries[pos].Timestamp) < w.trafficWindow {
		return
	}

	// At least one element in the queue:
	// Last element processed still in the queue OR
	// edge + at least last processed element still in the queue

	for pos <= w.lastProcessedPos &&
		t.Sub(w.queue.entries[pos].Timestamp) >= w.trafficWindow {

		w.size -= w.queue.entries[pos].Size
		w.edgePos = pos
		w.edge = w.queue.entries[pos]
		pos++

		if w.size <= w.threshold {

			timestamp := w.queue.entries[pos-1].Timestamp

			for pos <= w.lastProcessedPos &&
				w.queue.entries[pos].Timestamp == timestamp {

				w.size -= w.queue.entries[pos].Size
				w.edgePos = pos
				w.edge = w.queue.entries[pos]
				pos++
			}
			break
		}
	}

	if w.size > w.threshold {
		return
	}

	// edge in the queue
	timestamp := w.edge.Timestamp.Add(w.trafficWindow)

	w.status = StatusRecovered

	w.alerts = append(w.alerts, &Alert{
		Timestamp: timestamp,
		Total:     w.size,
		Status:    w.status,
	})
}
