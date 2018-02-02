package monitor

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"
	"time"
	"w3chttpd"
)

var logSamples = [][]byte{
	[]byte(`64.242.88.11 - - [07/Mar/2004:16:05:49 -0800] "GET /twiki/ HTTP/1.1" 401 12846`),
	[]byte(`127.0.0.1 - - [07/Mar/2004:16:06:51 -0800] "GET /twiki HTTP/1.1" 200 4523`),
	[]byte(`10.0.2.50 - - [07/Mar/2004:16:10:02 -0800] "POST /mailman/listinfo/ HTTP/2.0" 200 6291`),
}

func TestUnprocessedBytesAdd(t *testing.T) {

	ub := &unprocessedBytes{
		&sync.RWMutex{},
		[][]byte{},
	}

	ub.add(4, []byte("toto"))

	if len(ub.unprocessed) != 5 {
		t.Errorf("Slice length differs. Want %d, got %d", 5, len(ub.unprocessed))
	}

	ub.add(8, []byte("tata"))

	if len(ub.unprocessed) != 9 {
		t.Errorf("Slice length differs. Want %d, got %d", 9, len(ub.unprocessed))
	}
}

func TestConcatenateAll(t *testing.T) {

	ub := &unprocessedBytes{
		&sync.RWMutex{},
		[][]byte{},
	}

	ub.add(4, []byte("toto"))
	ub.add(8, []byte("tata"))

	res := ub.concatenateAll()

	expectedLength := 0
	for _, fragment := range ub.unprocessed {
		expectedLength += len(fragment)
	}

	if len(res) != expectedLength {
		t.Errorf("Slice length differs. Want %d, got %d",
			expectedLength, len(ub.unprocessed))
	}

	expectedSlice := []byte("tototata")
	if string(res) != string(expectedSlice) {
		t.Errorf("Slice differs. Want %s, got %s",
			string(expectedSlice), string(res))
	}
}

func TestReadBuffer(t *testing.T) {

	str := "abcdef"

	for i := 1; i < 10; i++ {

		bpool := &bufferPool{}
		bpool.init(10, i)

		rd := strings.NewReader(str)
		brd := bufio.NewReaderSize(rd, i)
		j := 0

		for {

			buf, err := readBuffer(brd, bpool)
			if err == io.EOF {
				break
			}

			end := j + i
			if j+i > len(str) {
				end = len(str)
			}

			expected := str[j:end]
			if string(buf) != expected {
				t.Errorf("Buffer differs. Want \"%s\", got \"%s\"", expected, string(buf))
			}
			j += i
		}
	}
}

func benchmarkReadBuffer(b *testing.B, bufferSize int) {

	str := "abcdef"

	bpool := &bufferPool{}
	bpool.init(10, bufferSize)

	for i := 0; i < b.N; i++ {

		rd := strings.NewReader(str)
		brd := bufio.NewReaderSize(rd, bufferSize)

		for {
			_, err := readBuffer(brd, bpool)

			if err == io.EOF {
				break
			}
		}
	}
}

func BenchmarkReadBuffer1(b *testing.B) {
	benchmarkReadBuffer(b, 1)
}

func BenchmarkReadBuffer2(b *testing.B) {
	benchmarkReadBuffer(b, 2)
}

func BenchmarkReadBuffer3(b *testing.B) {
	benchmarkReadBuffer(b, 3)
}

func BenchmarkReadBuffer4(b *testing.B) {
	benchmarkReadBuffer(b, 4)
}

func BenchmarkReadBuffer5(b *testing.B) {
	benchmarkReadBuffer(b, 5)
}

func BenchmarkReadBuffer6(b *testing.B) {
	benchmarkReadBuffer(b, 6)
}

func BenchmarkReadBuffer7(b *testing.B) {
	benchmarkReadBuffer(b, 7)
}

func TestProcessBuffer(t *testing.T) {

	buffer := []byte{}
	now := time.Now()

	for i := 0; i < 300; i++ {

		line := fmt.Sprintf("- - - [%s] \"GET /twiki/ HTTP/1.1\" 401 12846\n",
			now.Add(time.Duration(i)*time.Second).Format("02/Jan/2006:15:04:05 -0700"))
		buffer = append(buffer, line...)
	}

	truncated := append([]byte(nil), buffer[0:5]...)
	midBuffer := buffer[5 : len(buffer)-7]

	for i := 45; i < 51; i++ {

		eq := &entryQueue{
			&sync.RWMutex{},
			make([]*w3chttpd.Entry, 0),
			&entryPool{},
		}

		eq.epool.init(10)

		bpool := &bufferPool{}
		bpool.init(10, i)

		rd := bytes.NewReader(midBuffer)
		brd := bufio.NewReaderSize(rd, i)

		unprocessed := processBuffer(brd, bpool, eq, truncated)

		if len(eq.entries) != 299 {
			t.Errorf("Lengh of queue differs. Want %d, got %d", 299, len(eq.entries))
		}

		buffer = buffer[len(buffer)-7:]
		rd = bytes.NewReader(buffer)
		brd = bufio.NewReaderSize(rd, i)
		unprocessed = processBuffer(brd, bpool, eq, unprocessed)

		if len(eq.entries) != 300 {
			t.Errorf("Lengh of queue differs. Want %d, got %d", 300, len(eq.entries))
		}

		if len(unprocessed) != 0 {
			t.Errorf("Lengh of unprocessed differs. Want %d, got %d",
				0, len(unprocessed))
		}

		var previous time.Time
		for i, e := range eq.entries {
			if i != 0 && e.Timestamp.Before(previous) {
				t.Error("Queue not sorted")
			}
			previous = e.Timestamp
		}
	}
}

func benchmarkProcessBuffer(b *testing.B, bufferSize int) {

	bpool := &bufferPool{}
	bpool.init(10, bufferSize)

	for i := 0; i < b.N; i++ {

		buffer := bytes.Join(logSamples, []byte("\n"))
		buffer = append(buffer, []byte("\n")...)

		rd := bytes.NewReader(buffer)
		brd := bufio.NewReaderSize(rd, bufferSize)

		eq := &entryQueue{
			&sync.RWMutex{},
			make([]*w3chttpd.Entry, 0),
			&entryPool{},
		}

		eq.epool.init(10)

		processBuffer(brd, bpool, eq, nil)
	}
}

func BenchmarkProcessBuffer10(b *testing.B) {
	benchmarkProcessBuffer(b, 10)
}

func BenchmarkProcessBuffer100(b *testing.B) {
	benchmarkProcessBuffer(b, 100)
}

func BenchmarkProcessBuffer1000(b *testing.B) {
	benchmarkProcessBuffer(b, 1000)
}

func BenchmarkProcessBuffer10000(b *testing.B) {
	benchmarkProcessBuffer(b, 10000)
}

func TestExtractLines(t *testing.T) {

	eq := &entryQueue{
		&sync.RWMutex{},
		make([]*w3chttpd.Entry, 0),
		&entryPool{},
	}

	eq.epool.init(10)

	buffer := bytes.Join(logSamples, []byte("\n"))
	buffer = append(buffer, []byte("\n")...)
	unprocessed := extractLines(buffer, eq)

	expected := buffer[:len(logSamples[0])+1]

	if string(unprocessed) != string(expected) {
		t.Errorf("Unprocessed bytes differ. Want\"%s\", got \"%s\"",
			string(expected), string(unprocessed))
	}

	if len(eq.entries) != 2 {
		t.Errorf("Lengh of queue differs. Want %d, got %d", 2, len(eq.entries))
	}

	eq = &entryQueue{
		&sync.RWMutex{},
		make([]*w3chttpd.Entry, 0),
		&entryPool{},
	}

	eq.epool.init(10)

	buffer = bytes.Join(logSamples, []byte("\n"))
	unprocessed = extractLines(buffer, eq)

	expected = append(buffer[:len(logSamples[0])], []byte("\n")...)
	expected = append(expected, logSamples[2]...)

	if string(unprocessed) != string(expected) {
		t.Errorf("Unprocessed bytes differ. Want \"%s\", got \"%s\"",
			string(expected), string(unprocessed))
	}

	if len(eq.entries) != 1 {
		t.Errorf("Lengh of queue differs. Want %d, got %d", 1, len(eq.entries))
	}

	eq = &entryQueue{
		&sync.RWMutex{},
		make([]*w3chttpd.Entry, 0),
		&entryPool{},
	}

	eq.epool.init(10)

	buffer = logSamples[0]
	unprocessed = extractLines(buffer, eq)

	expected = logSamples[0]

	if string(unprocessed) != string(expected) {
		t.Errorf("Unprocessed bytes differ. Want \"%s\", got \"%s\"",
			string(expected), string(unprocessed))
	}

	if len(eq.entries) != 0 {
		t.Errorf("Lengh of queue differs. Want %d, got %d", 1, len(eq.entries))
	}
}
