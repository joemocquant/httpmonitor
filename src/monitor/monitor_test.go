package monitor

import (
	"bufio"
	"bytes"
	"io"
	"strings"
	"testing"
	"time"
)

func TestProccessLog(t *testing.T) {

	logSamples := [][]byte{
		[]byte(`127.0.0.1 - - [07/Mar/2004:16:00:00 -0800] "GET /twiki/ HTTP/1.1" 401 12846`),
		[]byte(`127.0.0.1 - - [07/Mar/2004:16:00:01 -0800] "GET /twiki HTTP/1.1" 200 4523`),
		[]byte(`10.0.2.50 - - [07/Mar/2004:16:10:02 -0800] "POST /mailman/listinfo/ HTTP/2.0" 200 6291`),
	}

	buffer := bytes.Join(logSamples, []byte("\n"))
	var rd io.Reader = bytes.NewReader(buffer)

	alertsChan := make(chan []*Alert)
	metricsChan := make(chan *Metrics)

	conf := &Config{
		MetricsFrequency: 10 * time.Second,
		TrafficWindow:    2 * time.Minute,
		Threshold:        500,
		BufferPoolSize:   10,
		BufferSize:       100,
		EntryPoolSize:    10,
		AlertsChan:       alertsChan,
		MetricsChan:      metricsChan,
	}

	go func() {
		for {
			select {
			case <-alertsChan:
			case <-metricsChan:
			}
		}
	}()

	conf.bpool = &bufferPool{}
	conf.bpool.init(conf.BufferPoolSize, conf.BufferSize)

	conf.brd = bufio.NewReaderSize(rd, conf.BufferSize)
	conf.w.init(conf.TrafficWindow, conf.Threshold, conf.EntryPoolSize)

	now := time.Unix(0, 0).UnixNano()
	unprocessed := []byte(nil)
	unprocessed = processLog(now, unprocessed, conf)

	if len(conf.w.queue.entries) != 2 {
		t.Errorf("Length of queue differs. Want %d, got %d",
			2, len(conf.w.queue.entries))
	}

	if len(unprocessed) != len(logSamples[2]) {
		t.Errorf("Length of unprocessed buffer differs. Want %d, got %d",
			len(logSamples[2]), len(unprocessed))
	}

	rd = strings.NewReader("\n")
	conf.brd = bufio.NewReaderSize(rd, conf.BufferSize)
	unprocessed = processLog(now, unprocessed, conf)

	if len(conf.w.queue.entries) != 3 {
		t.Errorf("Length of queue differs. Want %d, got %d",
			3, len(conf.w.queue.entries))
	}

	if len(unprocessed) != 0 {
		t.Errorf("Length of unprocessed buffer differs. Want %d, got %d",
			0, len(unprocessed))
	}

}
