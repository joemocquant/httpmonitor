package w3chttpd

import (
	"testing"
	"time"
)

func TestParseLine(t *testing.T) {

	logSamples := [][]byte{
		[]byte(`64.242.88.11 user-identifier frank [07/Mar/2004:16:05:49 -0800] "GET /twiki/ HTTP/1.1" 401 12846`),
		[]byte(`127.0.0.1 - - [07/Mar/2004:16:06:51 -0800] "GET /twiki HTTP/1.1" 200 4523`),
		[]byte(`10.0.2.50 - john [07/Mar/2004:16:10:02 -0800] "POST /mailman/listinfo/ HTTP/2.0" 200 6291`),
	}

	expectedResults := []*Entry{
		&Entry{
			Ip:         []byte("64.242.88.11"),
			ProtocolId: []byte("user-identifier"),
			UserId:     []byte("frank"),
			Req: Request{
				Method:   []byte("GET"),
				Resource: []byte("/twiki/"),
				Protocol: []byte("HTTP/1.1"),
			},
			StatusCode: 401,
			Size:       12846,
		},

		&Entry{
			Ip:         []byte("127.0.0.1"),
			ProtocolId: []byte("-"),
			UserId:     []byte("-"),
			Req: Request{
				Method:   []byte("GET"),
				Resource: []byte("/twiki"),
				Protocol: []byte("HTTP/1.1"),
			},
			StatusCode: 200,
			Size:       4523,
		},
		&Entry{
			Ip:         []byte("10.0.2.50"),
			ProtocolId: []byte("-"),
			UserId:     []byte("john"),
			Req: Request{
				Method:   []byte("POST"),
				Resource: []byte("/mailman/listinfo/"),
				Protocol: []byte("HTTP/2.0"),
			},
			StatusCode: 200,
			Size:       6291,
		},
	}

	expectedResults[0].Timestamp, _ =
		time.Parse("02/Jan/2006:15:04:05 -0700", "07/Mar/2004:16:05:49 -0800")
	expectedResults[1].Timestamp, _ =
		time.Parse("02/Jan/2006:15:04:05 -0700", "07/Mar/2004:16:06:51 -0800")
	expectedResults[2].Timestamp, _ =
		time.Parse("02/Jan/2006:15:04:05 -0700", "07/Mar/2004:16:10:02 -0800")

	for i, line := range logSamples {

		e := &Entry{}
		err := ParseLine(line, e)

		if err != nil {
			t.Errorf("[%s] An error occured: %v", string(line), err)
		}

		if string(e.Ip) != string(expectedResults[i].Ip) {
			t.Errorf("ip field differs. Want \"%s\", got \"%s\"",
				expectedResults[i].Ip, e.Ip)
		}

		if string(e.ProtocolId) != string(expectedResults[i].ProtocolId) {
			t.Errorf("protocolId field differs. Want \"%s\", got \"%s\"",
				expectedResults[i].ProtocolId, e.ProtocolId)
		}

		if string(e.UserId) != string(expectedResults[i].UserId) {
			t.Errorf("userId field differs. Want \"%s\", got \"%s\"",
				expectedResults[i].UserId, e.UserId)
		}

		if e.Timestamp != expectedResults[i].Timestamp {
			t.Errorf("timestamp field differs. Want \"%s\", got \"%s\"",
				expectedResults[i].Timestamp, e.Timestamp)
		}

		if string(e.Req.Method) != string(expectedResults[i].Req.Method) {
			t.Errorf("method field differs. Want \"%s\", got \"%s\"",
				expectedResults[i].Req.Method, e.Req.Method)
		}

		if string(e.Req.Resource) != string(expectedResults[i].Req.Resource) {
			t.Errorf("method field differs. Want \"%s\", got \"%s\"",
				expectedResults[i].Req.Resource, e.Req.Resource)
		}

		if string(e.Req.Protocol) != string(expectedResults[i].Req.Protocol) {
			t.Errorf("method field differs. Want \"%s\", got \"%s\"",
				expectedResults[i].Req.Protocol, e.Req.Protocol)
		}

		if string(e.StatusCode) != string(expectedResults[i].StatusCode) {
			t.Errorf("statusCode field differs. Want %d, got %d",
				expectedResults[i].StatusCode, e.StatusCode)
		}

		if string(e.Size) != string(expectedResults[i].Size) {
			t.Errorf("size field differs. Want %d, got %d",
				expectedResults[i].Size, e.Size)
		}
	}

	e := &Entry{}
	line := []byte("adfdfaf asdfa")
	err := ParseLine(line, e)

	if err == nil {
		t.Errorf("[%s] should return an error", string(line))
	}

}

func BenchmarkParseLine(b *testing.B) {

	e := &Entry{}

	for i := 0; i < b.N; i++ {
		ParseLine([]byte(`10.0.2.50 - john [07/Mar/2004:16:10:02 -0800] "POST /mailman/listinfo/hsdivision HTTP/2.0" 200 6291`), e)
	}
}

func TestConvertByteToInt(t *testing.T) {

	conversionTable := []struct {
		buf []byte
		val int
	}{
		{[]byte("123"), 123},
		{[]byte("4"), 4},
		{[]byte("54534343"), 54534343},
		{[]byte("0"), 0},
	}

	for _, rec := range conversionTable {

		got := convertByteToInt(rec.buf)

		if got != rec.val {
			t.Errorf("want %d, got %d", rec.val, got)
		}
	}
}
