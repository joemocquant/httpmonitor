package monitor

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"time"
	"w3chttpd"
)

// Logs are written to AccessLog in chronological order
// MetricsFrequency must be multiple of ReadFrequency
// Delay must be smaller than readFrequency
type Config struct {
	AccessLog        *io.Reader
	ReadFrequency    time.Duration
	MetricsFrequency time.Duration
	TrafficWindow    time.Duration
	Threshold        int
	BufferPoolSize   int
	BufferSize       int
	EntryPoolSize    int
	Delay            time.Duration
	AlertsChan       chan<- []*Alert
	MetricsChan      chan<- *Metrics

	// Internal parameters
	brd   *bufio.Reader
	bpool *bufferPool
	w     window
}

func Monitor(conf *Config) {

	if conf.MetricsFrequency%conf.ReadFrequency != 0 {
		panic(fmt.Errorf("MetricsFrequency should be a multiple of ReadFrequency"))
	}

	if conf.Delay >= conf.ReadFrequency {
		panic(fmt.Errorf("Delay must be smaller than ReadFrequency"))
	}

	conf.brd = bufio.NewReaderSize(*conf.AccessLog,
		conf.BufferPoolSize*conf.BufferSize)
	conf.w.init(conf.TrafficWindow, conf.Threshold, conf.EntryPoolSize)

	conf.bpool = &bufferPool{}
	conf.bpool.init(conf.BufferPoolSize, conf.BufferSize)

	unprocessedBytes := []byte{}
	frequency := conf.ReadFrequency
	var previousRun int64 = -1

	for {
		now := time.Now().UnixNano()
		nextRun := now - (now % int64(frequency)) + int64(frequency)
		time.Sleep(time.Duration(nextRun - now))

		if previousRun != -1 && now-previousRun >= int64(frequency) {
			log.Println("Potentially missing logs! Please adjust parameters")
		}

		// ProcessLog is blocking and should take less time
		// to execute than readFrequency,
		// otherwise parameters need adjustment
		unprocessedBytes = processLog(nextRun, unprocessedBytes, conf)

		time.Sleep(conf.Delay)
		previousRun = nextRun
	}
}

func processLog(now int64, unprocessedBytes []byte, conf *Config) []byte {

	unprocessedBytes = processBuffer(conf.brd, conf.bpool,
		conf.w.queue, unprocessedBytes)

	startMetrics := time.Unix(0, now-int64(conf.MetricsFrequency))
	startTrafficWindow := time.Unix(0, now-int64(conf.TrafficWindow))

	// Period = [ start - now [
	end := time.Unix(-1, now)

	// Remove outdated entries
	deleted := conf.w.queue.removeOutdatedEntries(conf.w.edge,
		startMetrics, startTrafficWindow)

	// Trigger metrics computation if needed
	if (now % int64(conf.MetricsFrequency)) == 0 {
		entries := conf.w.queue.getEntriesInWindow(startMetrics, end)

		copied := make([]*w3chttpd.Entry, len(entries))
		copy(copied, entries)

		go func() {
			conf.MetricsChan <- getMetricsForEntries(copied, startMetrics, end)
		}()
	}

	// Check Alerts at every readFrequency
	alerts := conf.w.getNewAlerts(end, deleted)
	if len(alerts) != 0 {
		conf.AlertsChan <- alerts
	}

	return unprocessedBytes
}

func Display(w io.Writer, alertsChan <-chan []*Alert,
	metricsChan <-chan *Metrics) {

	alertHistory := []*Alert{}

	for {

		select {

		case alerts := <-alertsChan:
			alertHistory = append(alertHistory, alerts...)
			for _, a := range alerts {
				fmt.Fprintln(w, a)
			}

		case metrics := <-metricsChan:
			fmt.Fprintln(w, metrics)
			for _, alert := range alertHistory {
				fmt.Fprintln(w, alert)
			}
		}
	}
}
