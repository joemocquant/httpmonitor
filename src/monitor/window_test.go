package monitor

import (
	"testing"
	"time"
	"w3chttpd"
)

func testUpdateEdge(t *testing.T, w window, wepos int) {

	if w.edgePos != wepos {
		t.Errorf("edgePos field differs. Want %v, got %v",
			wepos, w.edgePos)
	}
}

func TestUpdateEdge(t *testing.T) {

	conf := &Config{
		TrafficWindow: 2 * time.Minute,
		EntryPoolSize: 10,
	}

	conf.w.init(conf.TrafficWindow, 0, conf.EntryPoolSize)
	conf.w.queue.add(&w3chttpd.Entry{Timestamp: time.Unix(2, 0), Size: 100})
	conf.w.queue.add(&w3chttpd.Entry{Timestamp: time.Unix(2, 0), Size: 120})
	conf.w.queue.add(&w3chttpd.Entry{Timestamp: time.Unix(3, 0), Size: 300})
	conf.w.queue.add(&w3chttpd.Entry{Timestamp: time.Unix(122, 0), Size: 7})
	conf.w.queue.add(&w3chttpd.Entry{Timestamp: time.Unix(180, 0), Size: 8})
	conf.w.queue.add(&w3chttpd.Entry{Timestamp: time.Unix(250, 0), Size: 11})

	eq := conf.w.queue

	conf.w.updateEdge(eq.entries[0].Timestamp)
	testUpdateEdge(t, conf.w, -1)

	conf.w.updateEdge(eq.entries[1].Timestamp)
	testUpdateEdge(t, conf.w, -1)

	conf.w.updateEdge(eq.entries[2].Timestamp)
	testUpdateEdge(t, conf.w, -1)

	conf.w.updateEdge(eq.entries[3].Timestamp)
	testUpdateEdge(t, conf.w, 1)

	conf.w.updateEdge(eq.entries[4].Timestamp)
	testUpdateEdge(t, conf.w, 2)

	conf.w.updateEdge(eq.entries[5].Timestamp)
	testUpdateEdge(t, conf.w, 3)
}

func testGetNewAlerts(t *testing.T, w window, expectedAlerts []*Alert,
	expectedSize, expectedEdgePos int) {

	if w.size != expectedSize {
		t.Errorf("size value differs. Want %d, got %d",
			expectedSize, w.size)
	}

	if w.edgePos != expectedEdgePos {
		t.Errorf("edgePos value differs. Want %d, got %d",
			expectedEdgePos, w.edgePos)
	}

	if len(w.alerts) != len(expectedAlerts) {
		t.Errorf("Should want %d alerts. Got %d ",
			len(expectedAlerts), len(w.alerts))
		return
	}

	for i, a := range expectedAlerts {

		if w.alerts[i].Status != a.Status {
			t.Errorf("Alert should be of type %v", a.Status)
		}

		if w.alerts[i].Timestamp != a.Timestamp {
			t.Errorf("Alert timestamp differs. Want %v, got %v",
				a.Timestamp, w.alerts[i].Timestamp)
		}

		if w.alerts[i].Total != a.Total {
			t.Errorf("Alert value differs. Want %d, got %d",
				a.Total, w.alerts[i].Total)
		}
	}
}

func TestGetNewAlerts(t *testing.T) {

	conf := &Config{
		TrafficWindow: 2 * time.Minute,
		Threshold:     400,
		EntryPoolSize: 10,
	}

	conf.w.init(conf.TrafficWindow, conf.Threshold, conf.EntryPoolSize)
	conf.w.queue.add(&w3chttpd.Entry{Timestamp: time.Unix(0, 0), Size: 200})

	conf.w.getNewAlerts(time.Unix(200, 0), 0)
	expectedAlerts := []*Alert{}
	testGetNewAlerts(t, conf.w, expectedAlerts, 200, -1)

	conf = &Config{
		TrafficWindow: 2 * time.Minute,
		Threshold:     400,
		EntryPoolSize: 10,
	}

	conf.w.init(conf.TrafficWindow, conf.Threshold, conf.EntryPoolSize)

	conf.w.queue.add(&w3chttpd.Entry{Timestamp: time.Unix(0, 0), Size: 200})
	conf.w.queue.add(&w3chttpd.Entry{Timestamp: time.Unix(1, 0), Size: 300})
	conf.w.queue.add(&w3chttpd.Entry{Timestamp: time.Unix(2, 0), Size: 10})

	conf.w.getNewAlerts(time.Unix(2, 0), 0)

	expectedAlerts = []*Alert{
		&Alert{time.Unix(1, 0), 500, StatusExceed},
	}
	testGetNewAlerts(t, conf.w, expectedAlerts, 510, -1)

	conf = &Config{
		TrafficWindow: 2 * time.Minute,
		Threshold:     400,
		EntryPoolSize: 10,
	}

	conf.w.init(conf.TrafficWindow, conf.Threshold, conf.EntryPoolSize)

	conf.w.queue.add(&w3chttpd.Entry{Timestamp: time.Unix(2, 0), Size: 391})
	conf.w.queue.add(&w3chttpd.Entry{Timestamp: time.Unix(5, 0), Size: 10})
	conf.w.queue.add(&w3chttpd.Entry{Timestamp: time.Unix(6, 0), Size: 1})
	conf.w.queue.add(&w3chttpd.Entry{Timestamp: time.Unix(125, 0), Size: 15})

	conf.w.getNewAlerts(time.Unix(125, 0), 0)

	expectedAlerts = []*Alert{
		&Alert{time.Unix(5, 0), 401, StatusExceed},
		&Alert{time.Unix(122, 0), 11, StatusRecovered},
	}
	testGetNewAlerts(t, conf.w, expectedAlerts, 16, 1)

	conf = &Config{
		TrafficWindow: 2 * time.Minute,
		Threshold:     400,
		EntryPoolSize: 10,
	}

	conf.w.init(conf.TrafficWindow, conf.Threshold, conf.EntryPoolSize)

	conf.w.queue.add(&w3chttpd.Entry{Timestamp: time.Unix(1, 0), Size: 16})
	conf.w.queue.add(&w3chttpd.Entry{Timestamp: time.Unix(2, 0), Size: 390})    // exceed 2 at 406
	conf.w.queue.add(&w3chttpd.Entry{Timestamp: time.Unix(22, 0), Size: 15})    // window size 421
	conf.w.queue.add(&w3chttpd.Entry{Timestamp: time.Unix(25, 0), Size: 15})    // window size 436
	conf.w.queue.add(&w3chttpd.Entry{Timestamp: time.Unix(119, 0), Size: 15})   // window size 451
	conf.w.queue.add(&w3chttpd.Entry{Timestamp: time.Unix(120, 0), Size: 15})   // window size 466
	conf.w.queue.add(&w3chttpd.Entry{Timestamp: time.Unix(121, 0), Size: 15})   // window size 465
	conf.w.queue.add(&w3chttpd.Entry{Timestamp: time.Unix(122, 0), Size: 15})   // window size 90 recover
	conf.w.queue.add(&w3chttpd.Entry{Timestamp: time.Unix(123, 0), Size: 15})   //
	conf.w.queue.add(&w3chttpd.Entry{Timestamp: time.Unix(124, 0), Size: 1500}) // exceed window size 1605
	conf.w.queue.add(&w3chttpd.Entry{Timestamp: time.Unix(125, 0), Size: 15})

	conf.w.getNewAlerts(time.Unix(360, 0), 0) // recover 244

	expectedAlerts = []*Alert{
		&Alert{time.Unix(2, 0), 406, StatusExceed},
		&Alert{time.Unix(122, 0), 90, StatusRecovered},
		&Alert{time.Unix(124, 0), 1605, StatusExceed},
		&Alert{time.Unix(124+120, 0), 15, StatusRecovered},
	}
	testGetNewAlerts(t, conf.w, expectedAlerts, 15, 9)

	conf = &Config{
		TrafficWindow: 5 * time.Second,
		Threshold:     400,
		EntryPoolSize: 10,
	}

	toAdd := []*w3chttpd.Entry{
		&w3chttpd.Entry{Timestamp: time.Unix(1, 0), Size: 1},
		&w3chttpd.Entry{Timestamp: time.Unix(2, 0), Size: 540},
		&w3chttpd.Entry{Timestamp: time.Unix(3, 0), Size: 1},
		&w3chttpd.Entry{Timestamp: time.Unix(4, 0), Size: 1},
		&w3chttpd.Entry{Timestamp: time.Unix(5, 0), Size: 1},
		&w3chttpd.Entry{Timestamp: time.Unix(6, 0), Size: 1},
		&w3chttpd.Entry{Timestamp: time.Unix(6, 0), Size: 1},
		&w3chttpd.Entry{Timestamp: time.Unix(7, 0), Size: 1},
		&w3chttpd.Entry{Timestamp: time.Unix(7, 0), Size: 5},
		&w3chttpd.Entry{Timestamp: time.Unix(11, 0), Size: 390},
		&w3chttpd.Entry{Timestamp: time.Unix(11, 0), Size: 10},
		&w3chttpd.Entry{Timestamp: time.Unix(11, 0), Size: 2},
	}

	conf.w.init(conf.TrafficWindow, conf.Threshold, conf.EntryPoolSize)

	conf.w.queue.add(toAdd[0])
	conf.w.getNewAlerts(time.Unix(1, 0), 0)
	expectedAlerts = []*Alert{}
	testGetNewAlerts(t, conf.w, expectedAlerts, 1, -1)

	conf.w.queue.add(toAdd[1])
	conf.w.getNewAlerts(time.Unix(2, 0), 0)
	expectedAlerts = []*Alert{
		&Alert{time.Unix(2, 0), 541, StatusExceed},
	}
	testGetNewAlerts(t, conf.w, expectedAlerts, 541, -1)

	conf.w.queue.add(toAdd[2])
	conf.w.getNewAlerts(time.Unix(3, 0), 0)
	expectedAlerts = []*Alert{}
	testGetNewAlerts(t, conf.w, expectedAlerts, 542, -1)

	conf.w.queue.add(toAdd[3])
	conf.w.getNewAlerts(time.Unix(4, 0), 0)
	expectedAlerts = []*Alert{}
	testGetNewAlerts(t, conf.w, expectedAlerts, 543, -1)

	conf.w.queue.add(toAdd[4])
	conf.w.getNewAlerts(time.Unix(5, 0), 0)
	expectedAlerts = []*Alert{}
	testGetNewAlerts(t, conf.w, expectedAlerts, 544, -1)

	conf.w.queue.add(toAdd[5])
	conf.w.queue.add(toAdd[6])
	conf.w.getNewAlerts(time.Unix(6, 0), 0)
	expectedAlerts = []*Alert{}
	testGetNewAlerts(t, conf.w, expectedAlerts, 545, 0)

	conf.w.queue.add(toAdd[7])
	conf.w.queue.add(toAdd[8])
	conf.w.getNewAlerts(time.Unix(7, 0), 0)
	expectedAlerts = []*Alert{
		&Alert{time.Unix(7, 0), 11, StatusRecovered},
	}
	testGetNewAlerts(t, conf.w, expectedAlerts, 11, 1)

	conf.w.queue.add(toAdd[9])
	conf.w.queue.add(toAdd[10])
	conf.w.queue.add(toAdd[11])
	conf.w.getNewAlerts(time.Unix(11, 0), 0)
	expectedAlerts = []*Alert{
		&Alert{time.Unix(11, 0), 408, StatusExceed},
	}
	testGetNewAlerts(t, conf.w, expectedAlerts, 408, 6)

	conf = &Config{
		TrafficWindow: 5 * time.Second,
		Threshold:     250,
		EntryPoolSize: 10,
	}

	conf.w.init(conf.TrafficWindow, conf.Threshold, conf.EntryPoolSize)

	conf.w.queue.add(&w3chttpd.Entry{Timestamp: time.Unix(54, 0), Size: 30})
	conf.w.queue.add(&w3chttpd.Entry{Timestamp: time.Unix(54, 0), Size: 19})
	conf.w.queue.add(&w3chttpd.Entry{Timestamp: time.Unix(54, 0), Size: 5})
	conf.w.queue.add(&w3chttpd.Entry{Timestamp: time.Unix(54, 0), Size: 18}) // s = 72

	conf.w.queue.add(&w3chttpd.Entry{Timestamp: time.Unix(55, 0), Size: 27}) // s = 27

	conf.w.queue.add(&w3chttpd.Entry{Timestamp: time.Unix(56, 0), Size: 26})
	conf.w.queue.add(&w3chttpd.Entry{Timestamp: time.Unix(56, 0), Size: 24})
	conf.w.queue.add(&w3chttpd.Entry{Timestamp: time.Unix(56, 0), Size: 35})
	conf.w.queue.add(&w3chttpd.Entry{Timestamp: time.Unix(56, 0), Size: 19})
	conf.w.queue.add(&w3chttpd.Entry{Timestamp: time.Unix(56, 0), Size: 13}) // s = 117

	conf.w.queue.add(&w3chttpd.Entry{Timestamp: time.Unix(58, 0), Size: 14})
	conf.w.queue.add(&w3chttpd.Entry{Timestamp: time.Unix(58, 0), Size: 38})
	conf.w.queue.add(&w3chttpd.Entry{Timestamp: time.Unix(58, 0), Size: 14})
	conf.w.queue.add(&w3chttpd.Entry{Timestamp: time.Unix(58, 0), Size: 4})
	conf.w.queue.add(&w3chttpd.Entry{Timestamp: time.Unix(58, 0), Size: 12}) // s = 82, ws = 298 // exceed!!

	//59 :  ws = 298 - 72 = 226 // recover

	conf.w.queue.add(&w3chttpd.Entry{Timestamp: time.Unix(60, 0), Size: 39})
	conf.w.queue.add(&w3chttpd.Entry{Timestamp: time.Unix(60, 0), Size: 39})
	conf.w.queue.add(&w3chttpd.Entry{Timestamp: time.Unix(60, 0), Size: 9})
	conf.w.queue.add(&w3chttpd.Entry{Timestamp: time.Unix(60, 0), Size: 1})
	conf.w.queue.add(&w3chttpd.Entry{Timestamp: time.Unix(60, 0), Size: 6})
	conf.w.queue.add(&w3chttpd.Entry{Timestamp: time.Unix(60, 0), Size: 36})
	conf.w.queue.add(&w3chttpd.Entry{Timestamp: time.Unix(60, 0), Size: 39})
	conf.w.queue.add(&w3chttpd.Entry{Timestamp: time.Unix(60, 0), Size: 36})
	conf.w.queue.add(&w3chttpd.Entry{Timestamp: time.Unix(60, 0), Size: 32})
	conf.w.queue.add(&w3chttpd.Entry{Timestamp: time.Unix(60, 0), Size: 1}) // s = 238, ws = 226 + 238 -27 = 437

	conf.w.getNewAlerts(time.Unix(60, 0), 0)
	expectedAlerts = []*Alert{
		&Alert{time.Unix(58, 0), 298, StatusExceed},
		&Alert{time.Unix(59, 0), 226, StatusRecovered},
		&Alert{time.Unix(60, 0), 437, StatusExceed},
	}
	testGetNewAlerts(t, conf.w, expectedAlerts, 437, 4)
}
