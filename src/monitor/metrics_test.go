package monitor

import (
	"testing"
	"time"
	"w3chttpd"
)

func TestGetMetricsForEntries(t *testing.T) {

	entries := []*w3chttpd.Entry{
		&w3chttpd.Entry{
			Req:        w3chttpd.Request{Resource: []byte("/")},
			StatusCode: 200,
			Size:       350,
		},
		&w3chttpd.Entry{
			Req:        w3chttpd.Request{Resource: []byte("/toto")},
			StatusCode: 400,
			Size:       125,
		},
		&w3chttpd.Entry{
			Req:        w3chttpd.Request{Resource: []byte("/test")},
			StatusCode: 400,
			Size:       125,
		},
		&w3chttpd.Entry{
			Req:        w3chttpd.Request{Resource: []byte("/toto#aa")},
			StatusCode: 200,
			Size:       500,
		},
		&w3chttpd.Entry{
			Req:        w3chttpd.Request{Resource: []byte("/tata")},
			StatusCode: 200,
			Size:       500,
		},
		&w3chttpd.Entry{
			Req:        w3chttpd.Request{Resource: []byte("/toto")},
			StatusCode: 200,
			Size:       500,
		},
		&w3chttpd.Entry{
			Req:        w3chttpd.Request{Resource: []byte("/test")},
			StatusCode: 400,
			Size:       125,
		},
	}

	metrics := getMetricsForEntries(entries, time.Now(), time.Now())

	if len(metrics.Rank) != 3 {
		t.Errorf("Length of metrics differs. Want %d, got %d",
			3, len(metrics.Rank))
	}

	expectedRank := []Rank{
		Rank{3, "toto"},
		Rank{2, "test"},
		Rank{1, "tata"},
	}

	for i, rec := range metrics.Rank {

		if rec.HitCount != expectedRank[i].HitCount &&
			rec.Section != expectedRank[i].Section {
			t.Errorf("ranked in the wrong order: want %v, got %+v",
				expectedRank, metrics.Rank)
			break
		}
	}

	totalTraffic, requestCount, errorCount := 0, 0, 0
	for _, e := range entries {
		totalTraffic += e.Size
		requestCount++
		if e.StatusCode >= 400 {
			errorCount++
		}
	}

	if metrics.RequestCount != requestCount {
		t.Errorf("RequestCount field differs. Want %d, got %d",
			requestCount, metrics.RequestCount)
	}

	if metrics.ErrorCount != errorCount {
		t.Errorf("ErrorCount field differs. Want %d, got %d",
			errorCount, metrics.ErrorCount)
	}

	if metrics.TotalTraffic != totalTraffic {
		t.Errorf("totalTraffic field differs. Want %d, got %d",
			totalTraffic, metrics.TotalTraffic)
	}
}

func TestGetSection(t *testing.T) {

	resourceTable := []struct {
		resource []byte
		section  []byte
	}{
		{[]byte("/twiki/bin/rdiff/TWiki/NewUserTemplate?rev1=1.3&rev2=1.2"), []byte("twiki")},
		{[]byte("/mailman/listinfo/hsdivision"), []byte("mailman")},
		{[]byte("/1#subsection/2/3"), []byte("1")},
		{[]byte("/twiki?test=data"), []byte("twiki")},
		{[]byte("/twiki"), []byte("twiki")},
		{[]byte("/twiki///"), []byte("twiki")},
		{[]byte("///twiki"), []byte("twiki")},
		{[]byte("////"), nil},
		{[]byte("/"), nil},
	}

	for _, rec := range resourceTable {

		got := getSection(rec.resource)

		if string(got) != string(rec.section) {
			t.Errorf("want \"%s\", got \"%s\"", string(rec.section), string(got))
		}
	}
}
