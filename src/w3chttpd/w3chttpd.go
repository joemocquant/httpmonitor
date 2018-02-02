package w3chttpd

import (
	"fmt"
	"time"
	"unicode/utf8"
)

// https://en.wikipedia.org/wiki/Common_Log_Format
// using []byte instead of string to avoid multiple memory allocations

type Request struct {
	Method   []byte
	Resource []byte
	Protocol []byte
}

type Entry struct {
	line       []byte
	Ip         []byte
	ProtocolId []byte
	UserId     []byte
	Timestamp  time.Time
	Req        Request
	StatusCode int
	Size       int
}

func (e *Entry) String() string {

	str := fmt.Sprintf("%s %s %s [%s] \"%s %s %s\" %d %d",
		string(e.Ip), string(e.ProtocolId), string(e.UserId),
		e.Timestamp.String(), string(e.Req.Method), string(e.Req.Resource),
		string(e.Req.Protocol), e.StatusCode, e.Size)

	return str
}

// parse a e.line of log with only 2 allocation (e.line and timestamp)
func ParseLine(line []byte, e *Entry) error {

	e.line = append([]byte{}, line...)

	// ip
	start := 0
	i, s := parseField(e.line, start, ' ')
	if i == -1 {
		return fmt.Errorf("parseField: wrong format: \"%s\"", string(e.line))
	}
	e.Ip = e.line[start:i]
	start = i + s

	// protocolId
	i, s = parseField(e.line, start, ' ')
	if i == -1 {
		return fmt.Errorf("parseField: wrong format: \"%s\"", string(e.line))
	}
	e.ProtocolId = e.line[start:i]
	start = i + s

	// userId
	i, s = parseField(e.line, start, ' ')
	if i == -1 {
		return fmt.Errorf("parseField: wrong format: \"%s\"", string(e.line))
	}
	e.UserId = e.line[start:i]
	start = i + s

	// timestamp
	start += 1 // eating '['
	i, s = parseField(e.line, start, ']')
	if i == -1 {
		return fmt.Errorf("parseField: wrong format: \"%s\"", string(e.line))
	}
	t, err := time.Parse("02/Jan/2006:15:04:05 -0700", string(e.line[start:i]))
	if err != nil {
		return fmt.Errorf("time.Parse: %v", err)
	}
	e.Timestamp = t
	start = i + s + 1 // eating ']'

	// method
	start += 1 // eating '"'
	i, s = parseField(e.line, start, ' ')
	if i == -1 {
		return fmt.Errorf("parseField: wrong format: \"%s\"", string(e.line))
	}
	e.Req.Method = e.line[start:i]
	start = i + s

	// resource
	i, s = parseField(e.line, start, ' ')
	if i == -1 {
		return fmt.Errorf("parseField: wrong format: \"%s\"", string(e.line))
	}
	e.Req.Resource = e.line[start:i]
	start = i + s

	// protocol
	i, s = parseField(e.line, start, '"')
	if i == -1 {
		return fmt.Errorf("parseField: wrong format: \"%s\"", string(e.line))
	}
	e.Req.Protocol = e.line[start:i]
	start = i + s

	// statusCode
	start += 1 // eating '"'
	i, s = parseField(e.line, start, ' ')
	if i == -1 {
		return fmt.Errorf("parseField: wrong format: \"%s\"", string(e.line))
	}
	val := convertByteToInt(e.line[start:i])
	e.StatusCode = val
	start = i + s

	val = convertByteToInt(e.line[start:])
	e.Size = val

	return nil
}

func parseField(line []byte, start int, delimiter rune) (int, int) {

	for i := start; i < len(line); {
		r, s := utf8.DecodeRune(line[i:])
		if r == delimiter {
			return i, s
		}
		i += s
	}
	return -1, -1
}

func convertByteToInt(buf []byte) int {

	var n int = 0
	var scale = 1

	for i := len(buf) - 1; i >= 0; i-- {
		n += scale * (int(buf[i]) - int('0'))
		scale *= 10
	}

	return n
}
