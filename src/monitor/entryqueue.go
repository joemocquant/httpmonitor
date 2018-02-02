package monitor

import (
	"sync"
	"time"
	"w3chttpd"
)

type entryQueue struct {
	*sync.RWMutex
	entries []*w3chttpd.Entry
	epool   *entryPool
}

func (eq *entryQueue) Len() int {

	eq.RLock()
	defer eq.RUnlock()

	return len(eq.entries)
}

func (eq *entryQueue) Less(i, j int) bool {

	eq.RLock()
	defer eq.RUnlock()

	return eq.entries[i].Timestamp.Before(eq.entries[j].Timestamp)
}

func (eq *entryQueue) Swap(i, j int) {

	eq.Lock()
	defer eq.Unlock()

	eq.entries[i], eq.entries[j] = eq.entries[j], eq.entries[i]
}

func (eq *entryQueue) add(entry *w3chttpd.Entry) {

	if entry == nil {
		return
	}

	eq.Lock()
	defer eq.Unlock()

	eq.entries = append(eq.entries, entry)
}

func (eq *entryQueue) getEntriesInWindow(
	start, end time.Time) []*w3chttpd.Entry {

	eq.RLock()
	defer eq.RUnlock()

	i := 0
	j := len(eq.entries) - 1

	for i < len(eq.entries) && start.After(eq.entries[i].Timestamp) {
		i++
	}

	if i == len(eq.entries) {
		return nil
	}

	for j >= i && eq.entries[j].Timestamp.After(end) {
		j--
	}

	if j < i {
		return nil
	}

	return eq.entries[i : j+1]
}

func (eq *entryQueue) next(entry *w3chttpd.Entry) *w3chttpd.Entry {

	eq.RLock()
	defer eq.RUnlock()

	if entry == nil {
		return nil
	}

	for i, e := range eq.entries {

		if e == entry {
			if i+1 < len(eq.entries) {
				return eq.entries[i+1]
			}
			return nil
		}
	}

	return nil
}

func (eq *entryQueue) entriesWithStartEntry(
	entry *w3chttpd.Entry) []*w3chttpd.Entry {

	eq.RLock()
	defer eq.RUnlock()

	if entry == nil {
		return nil
	}

	for i, e := range eq.entries {

		if e == entry {
			return eq.entries[i:]
		}

	}
	return nil
}

func (eq *entryQueue) removeOutdatedEntries(edge *w3chttpd.Entry,
	startMetrics, startTrafficWindow time.Time) int {

	var min time.Time

	if edge == nil {
		return 0
	}

	if startMetrics.Before(startTrafficWindow) {
		min = startMetrics
	} else {
		min = startTrafficWindow
	}

	eq.Lock()
	defer eq.Unlock()

	start := 0
	for _, e := range eq.entries {

		if e != edge && e.Timestamp.Before(min) {
			start++
		} else {
			break
		}
	}

	for i := 0; i < start; i++ {
		eq.epool.recycle(eq.entries[i])
		eq.entries[i] = nil
	}

	eq.entries = eq.entries[start:]

	return start
}
