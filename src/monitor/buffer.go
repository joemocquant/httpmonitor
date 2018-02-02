package monitor

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"sort"
	"sync"
	"unicode/utf8"
	"w3chttpd"
)

type unprocessedBytes struct {
	*sync.RWMutex
	unprocessed [][]byte
}

func (u *unprocessedBytes) add(bufPos int, fragment []byte) {

	if len(fragment) == 0 {
		return
	}

	u.Lock()
	defer u.Unlock()

	if bufPos < len(u.unprocessed) {
		u.unprocessed[bufPos] = fragment

	} else {

		tmp := make([][]byte, bufPos+1)
		copy(tmp, u.unprocessed)
		tmp[bufPos] = fragment
		u.unprocessed = tmp
	}
}

func (u *unprocessedBytes) concatenateAll() []byte {

	u.RLock()
	defer u.RUnlock()

	res := []byte{}

	for _, someBytes := range u.unprocessed {
		res = append(res, someBytes...)
	}

	return res
}

func processBuffer(brd *bufio.Reader, bpool *bufferPool,
	queue *entryQueue, unprocessed []byte) []byte {

	tempQueue := &entryQueue{
		&sync.RWMutex{},
		make([]*w3chttpd.Entry, 0),
		queue.epool,
	}

	ub := &unprocessedBytes{
		&sync.RWMutex{},
		[][]byte{},
	}
	ub.add(0, unprocessed)

	bufPos := 1
	wg := sync.WaitGroup{}

	for {

		buf, err := readBuffer(brd, bpool)
		if err == io.EOF || err != nil {
			break
		}

		wg.Add(1)
		go func(buf []byte, err error, bufPos int) {

			defer wg.Done()

			up := extractLines(buf, tempQueue)
			ub.add(bufPos, append([]byte{}, up...))
			bpool.recycle(buf)

		}(buf, err, bufPos)

		bufPos++
	}

	wg.Wait()

	buf := ub.concatenateAll()
	buf = extractLines(buf, tempQueue)
	unprocessed = extractLine(buf, tempQueue)

	sort.Sort(tempQueue)
	queue.entries = append(queue.entries, tempQueue.entries...)

	return unprocessed
}

func readBuffer(brd *bufio.Reader, bpool *bufferPool) ([]byte, error) {

	buffer := bpool.get()
	n, err := brd.Read(buffer)

	if err != nil && err == io.EOF {
		return nil, err
	}

	if err != nil {
		return nil, fmt.Errorf("bufio.Reader.Read: %v", err)
	}

	return buffer[0:n], nil
}

// Extract lines between first \n and last \n in the buffer
// Since it is impossible to align buffer size with a log line sizes
// (a log line having a variable size)
func extractLines(buffer []byte, queue *entryQueue) []byte {

	startBuf, endBuf := 0, 0

	for i := 0; i < len(buffer) && startBuf == 0; {
		r, s := utf8.DecodeRune(buffer[i:])
		if r == '\n' {
			startBuf = i + s
		}
		i += s
	}

	for i := len(buffer) - 1; i >= 0 && endBuf == 0; {
		r, _ := utf8.DecodeRune(buffer[i:])
		if r == '\n' {
			endBuf = i
		}
		i -= 1
	}

	if endBuf <= startBuf {
		return append([]byte{}, buffer...)
	}

	currentStart := startBuf
	for i := currentStart; i <= endBuf; {

		r, s := utf8.DecodeRune(buffer[i:])

		if r == '\n' {

			e := queue.epool.get()
			err := w3chttpd.ParseLine(buffer[currentStart:i], e)

			if err != nil {
				log.Printf("ParseLine: %v", err)
				queue.epool.recycle(e)

			} else {
				queue.add(e)
			}
			currentStart = i + s
		}
		i += s
	}

	if endBuf+1 < len(buffer) {
		res := append([]byte{}, buffer[:startBuf]...)
		return append(res, buffer[endBuf+1:]...)
	}

	return append([]byte{}, buffer[:startBuf]...)
}

func extractLine(buffer []byte, queue *entryQueue) []byte {

	if len(buffer) == 0 {
		return nil
	}

	for i := 0; i <= len(buffer); {

		r, s := utf8.DecodeRune(buffer[i:])

		if r == '\n' {

			e := queue.epool.get()
			err := w3chttpd.ParseLine(buffer[:i], e)

			if err != nil {
				log.Printf("ParseLine: %v", err)
				queue.epool.recycle(e)

			} else {
				queue.add(e)
			}

			if i+s < len(buffer) {
				return append([]byte{}, buffer[i+s:]...)
			}
			return nil
		}
		i += s
	}

	return append([]byte{}, buffer...)
}
