package monitor

import "w3chttpd"

type entryPool struct {
	pool chan *w3chttpd.Entry
}

func (ep *entryPool) init(poolSize int) {

	ep.pool = make(chan *w3chttpd.Entry, poolSize)
}

func (ep *entryPool) get() *w3chttpd.Entry {

	var e *w3chttpd.Entry

	select {
	case e = <-ep.pool:
	default:
		e = &w3chttpd.Entry{}
	}

	return e
}

func (ep *entryPool) recycle(e *w3chttpd.Entry) {

	select {
	case ep.pool <- e:
	default:
	}
}

type bufferPool struct {
	bufferSize     int
	pool           chan []byte
	recyclingQueue [][]byte
}

func (bp *bufferPool) init(poolSize, bufferSize int) {

	bp.pool = make(chan []byte, poolSize)
	bp.bufferSize = bufferSize
}

func (bp *bufferPool) get() []byte {

	var buf []byte

	select {
	case buf = <-bp.pool:
	default:
		buf = make([]byte, bp.bufferSize)
	}

	return buf
}

func (bp *bufferPool) recycle(buf []byte) {

	if cap(buf) != bp.bufferSize {
		return
	}

	select {
	case bp.pool <- buf:
	default:
	}
}
