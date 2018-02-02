package monitor

import (
	"sync"
	"testing"
	"time"
	"w3chttpd"
)

func TestAdd(t *testing.T) {

	eq := &entryQueue{
		&sync.RWMutex{},
		make([]*w3chttpd.Entry, 0),
		&entryPool{},
	}

	eq.add(nil)

	if len(eq.entries) != 0 {
		t.Errorf("Length of queue differs. Want %d, got %d", 0, len(eq.entries))
	}

	toAdd := &w3chttpd.Entry{}
	eq.add(toAdd)

	if len(eq.entries) != 1 {
		t.Errorf("Length of queue differs. Want %d, got %d", 1, len(eq.entries))
	}
}

func TestGetEntriesInWindow(t *testing.T) {

	eq := &entryQueue{
		&sync.RWMutex{},
		[]*w3chttpd.Entry{
			&w3chttpd.Entry{Timestamp: time.Unix(2, 0)},
			&w3chttpd.Entry{Timestamp: time.Unix(3, 0)},
			&w3chttpd.Entry{Timestamp: time.Unix(4, 0)},
			&w3chttpd.Entry{Timestamp: time.Unix(4, 0)},
			&w3chttpd.Entry{Timestamp: time.Unix(5, 0)},
			&w3chttpd.Entry{Timestamp: time.Unix(5, 0)},
		},
		&entryPool{},
	}

	entries := eq.getEntriesInWindow(time.Unix(3, 0), time.Unix(4, 0))
	if len(entries) != 3 {
		t.Errorf("Length of queue differs. Want %d, got %d", 3, len(entries))
	}

	want := time.Unix(3, 0)
	if entries[0].Timestamp != want {
		t.Errorf("Queue differs. Want first timestamp %v, got %v",
			want, entries[0].Timestamp)
	}

	entries = eq.getEntriesInWindow(time.Unix(2, 0), time.Unix(6, 0))
	if len(entries) != 6 {
		t.Errorf("Length of queue differs. Want %d, got %d", 6, len(entries))
	}

	want = time.Unix(2, 0)
	if entries[0].Timestamp != want {
		t.Errorf("Queue differs. Want first timestamp %v, got %v",
			want, entries[0].Timestamp)
	}

	want = time.Unix(5, 0)
	if entries[len(entries)-1].Timestamp != want {
		t.Errorf("Queue differs. Want last timestamp %v, got %v",
			want, entries[len(entries)-1].Timestamp)
	}

	entries = eq.getEntriesInWindow(time.Unix(0, 0), time.Unix(1, 0))
	if len(entries) != 0 {
		t.Errorf("Length of queue differs. Want %d, got %d", 0, len(entries))
	}

	entries = eq.getEntriesInWindow(time.Unix(3, 1), time.Unix(3, 2))
	if len(entries) != 0 {
		t.Errorf("Length of queue differs. Want %d, got %d", 0, len(entries))
	}

	entries = eq.getEntriesInWindow(time.Unix(6, 0), time.Unix(7, 0))
	if len(entries) != 0 {
		t.Errorf("Length of queue differs. Want %d, got %d", 0, len(entries))
	}
}

func TestEntriesWithStartEntry(t *testing.T) {

	entry := &w3chttpd.Entry{Timestamp: time.Unix(3, 0)}

	eq := &entryQueue{
		&sync.RWMutex{},
		[]*w3chttpd.Entry{
			&w3chttpd.Entry{Timestamp: time.Unix(2, 0)},
			entry,
			&w3chttpd.Entry{Timestamp: time.Unix(4, 0)},
			&w3chttpd.Entry{Timestamp: time.Unix(5, 0)},
			&w3chttpd.Entry{Timestamp: time.Unix(5, 0)},
		},
		&entryPool{},
	}

	entries := eq.entriesWithStartEntry(entry)

	if len(entries) != 4 {
		t.Errorf("Length of queue differs. Want %d, got %d", 4, len(entries))
	}

	if entries[0] != entry {
		t.Errorf("First element of queue differs. Want %v, got %v", entry, entries[0])
	}

}

func TestRemoveOutdatedEntries(t *testing.T) {

	edge := &w3chttpd.Entry{Timestamp: time.Unix(3, 0)}
	next := &w3chttpd.Entry{Timestamp: time.Unix(4, 0)}

	startMetrics := time.Unix(4, 0)
	startTrafficWindow := time.Unix(5, 0)

	eq := &entryQueue{
		&sync.RWMutex{},
		[]*w3chttpd.Entry{
			&w3chttpd.Entry{Timestamp: time.Unix(2, 0)},
			edge,
			next,
			&w3chttpd.Entry{Timestamp: time.Unix(4, 0)},
			&w3chttpd.Entry{Timestamp: time.Unix(5, 0)},
			&w3chttpd.Entry{Timestamp: time.Unix(6, 0)},
		},
		&entryPool{},
	}

	count := eq.removeOutdatedEntries(edge, startMetrics,
		startTrafficWindow)

	if count != 1 {
		t.Errorf("Number of entries deleted differs. Want %d, got %d",
			1, count)
	}

	if len(eq.entries) != 5 {
		t.Errorf("Length of queue differs. Want %d, got %d",
			5, len(eq.entries))
	}

	if eq.entries[0] != edge {
		t.Errorf("First element of queue differs. Want %v, got %v",
			next, eq.entries[0])
	}

	startMetrics = time.Unix(4, 0)
	startTrafficWindow = time.Unix(2, 2)

	eq = &entryQueue{
		&sync.RWMutex{},
		[]*w3chttpd.Entry{
			&w3chttpd.Entry{Timestamp: time.Unix(2, 0)},
			&w3chttpd.Entry{Timestamp: time.Unix(2, 1)},
			edge,
			next,
			&w3chttpd.Entry{Timestamp: time.Unix(5, 0)},
			&w3chttpd.Entry{Timestamp: time.Unix(6, 0)},
		},
		&entryPool{},
	}

	count = eq.removeOutdatedEntries(edge, startMetrics,
		startTrafficWindow)

	if count != 2 {
		t.Errorf("Number of entries deleted differs. Want %d, got %d",
			2, count)
	}

	if len(eq.entries) != 4 {
		t.Errorf("Length of queue differs. Want %d, got %d",
			4, len(eq.entries))
	}

	if eq.entries[0] != edge {
		t.Errorf("First element of queue differs. Want %v, got %v",
			next, eq.entries[0])
	}

	startMetrics = time.Unix(14, 0)
	startTrafficWindow = time.Unix(14, 0)

	count = eq.removeOutdatedEntries(nil, startMetrics,
		startTrafficWindow)

	if count != 0 {
		t.Errorf("Number of entries deleted differs. Want %d, got %d",
			0, count)
	}

	if len(eq.entries) != 4 {
		t.Errorf("Length of queue differs. Want %d, got %d",
			4, len(eq.entries))
	}
}
