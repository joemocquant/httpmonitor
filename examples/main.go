package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"monitor"
	"os"
	"time"
)

func generateLogs(path string) {

	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	sections := []string{
		"contact", "help", "admin", "support", "pricing", "team",
	}

	rand.Seed(time.Now().Unix())

	for {

		var lines []byte

		for i := 0; i < 5000; i++ {

			j := rand.Intn(len(sections))
			size := rand.Intn(40)
			ip := rand.Intn(10) + 1

			line := fmt.Sprintf("127.0.0.%d - - [%s] \"GET /%s HTTP/1.1\" 200 %d\n", ip,
				time.Now().Format("02/Jan/2006:15:04:05 -0700"), sections[j], size)
			lines = append(lines, []byte(line)...)
		}

		f.Write(lines)
		j := rand.Intn(3)
		time.Sleep(time.Duration(j) * time.Second)
	}
}

func main() {

	path := flag.String("path", "access.log",
		"log file path")

	readFrequency := flag.Int("read-frequency", 1000,
		"Frequency at which file is accessed (in milliseconds)")

	metricsFrequency := flag.Int("metrics-frequency", 10,
		"Frequency at which metrics are displayed (in seconds)")

	trafficWindow := flag.Int("traffic-window", 120,
		"Sliding window for alerting (in seconds)")

	threshold := flag.Int("treshold", 250,
		"Value for which an alert is triggered (in bytes)")

	bufferPoolSize := flag.Int("buffer-pool-size", 20,
		"number of buffers in buffer pool")

	bufferSize := flag.Int("buffer-size", 1024*1024, // 1mb
		"Minimum size for buffer (in bytes)")

	entryPoolSize := flag.Int("entry-pool-size", 12000000,
		"number of entries in entry pool")

	delay := flag.Int("delay", 0,
		"delay to retrieve logs (in milliseconds)")

	flag.Parse()

	os.Remove(*path)

	f, err := os.OpenFile(*path, os.O_CREATE|os.O_RDONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	accessLog := io.Reader(f)
	alertsChan := make(chan []*monitor.Alert)
	metricsChan := make(chan *monitor.Metrics)

	conf := &monitor.Config{
		AccessLog:        &accessLog,
		ReadFrequency:    time.Duration(*readFrequency) * time.Millisecond,
		MetricsFrequency: time.Duration(*metricsFrequency) * time.Second,
		TrafficWindow:    time.Duration(*trafficWindow) * time.Second,
		Threshold:        *threshold,
		BufferPoolSize:   *bufferPoolSize,
		BufferSize:       *bufferSize,
		EntryPoolSize:    *entryPoolSize,
		Delay:            time.Duration(*delay) * time.Millisecond,
		AlertsChan:       alertsChan,
		MetricsChan:      metricsChan,
	}

	//go generateLogs(*path)
	go monitor.Display(os.Stdout, alertsChan, metricsChan)
	monitor.Monitor(conf)
}

// echo "- - - [`date "+%d/%b/%Y:%H:%M:%S %z"`] \"GET /twiki/ HTTP/1.1\" 401 12846" >> access.log
// echo -n "- - - [`date "+%d/%b/%Y:%H:%M:%S %z"`] \"GET /twiki/ HTTP/1.1\"" >> access.log
// echo " 401 12846" >> access.log
