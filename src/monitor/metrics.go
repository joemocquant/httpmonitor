package monitor

import (
	"bufio"
	"bytes"
	"fmt"
	"sort"
	"text/tabwriter"
	"time"
	"unicode/utf8"
	"w3chttpd"
)

type Rank struct {
	HitCount int
	Section  string
}

type Metrics struct {
	Rank           []Rank
	RequestCount   int
	ErrorCount     int
	TotalTraffic   int
	PeriodStart    time.Time
	PeriodEnd      time.Time
	UniqueVisitors int
	AvgPageViews   float32
}

type ranking []Rank

func (r ranking) Len() int           { return len(r) }
func (r ranking) Less(i, j int) bool { return r[i].HitCount > r[j].HitCount }
func (r ranking) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }

func (m *Metrics) String() string {

	str := fmt.Sprintf("\n[%s - %s] Requests: %d | Errors: %d"+
		" | Traffic: %d\n", m.PeriodStart.Format("02/01/2006:15:04:05"),
		m.PeriodEnd.Format("02/01/2006:15:04:05"), m.RequestCount,
		m.ErrorCount, m.TotalTraffic)

	str += fmt.Sprintf("Unique visitors: %d (Avg page views per visitor: %.2f) \n",
		m.UniqueVisitors, m.AvgPageViews)

	if len(m.Rank) == 0 {
		return str
	}

	tables := bytes.Buffer{}
	w := bufio.NewWriter(&tables)
	tw := tabwriter.NewWriter(w, 80, 0, 0, '.', tabwriter.TabIndent)

	line := ""
	for i := 0; i < 83; i++ {
		line += "\\"
	}
	fmt.Fprintf(tw, line+"\n")

	line = ""
	for i := 0; i < 73; i++ {
		line += " "
	}
	fmt.Fprintf(tw, "SECTION"+line+"HITS\n")

	line = ""
	for i := 0; i < 84; i++ {
		line += "-"
	}
	fmt.Fprintf(tw, line+"\n")

	for _, r := range m.Rank {
		fmt.Fprintf(tw, "%s\t%d\n", r.Section, r.HitCount)
	}
	line = ""
	for i := 0; i < 80; i++ {
		line += "/"
	}

	fmt.Fprintf(tw, line+"\n")
	tw.Flush()
	w.Flush()

	return str + tables.String()
}

func getMetricsForEntries(entries []*w3chttpd.Entry,
	periodStart, periodEnd time.Time) *Metrics {

	m := &Metrics{
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
	}

	hits := make(map[string]int, len(entries))
	visitors := make(map[string]int, 0)

	for _, e := range entries {

		m.TotalTraffic += e.Size
		m.RequestCount++
		if e.StatusCode >= 400 {
			m.ErrorCount++
		}

		visitors[string(e.Ip)]++

		section := getSection(e.Req.Resource)
		if section == nil {
			continue
		}

		hits[string(section)]++
	}

	m.UniqueVisitors = len(visitors)
	m.AvgPageViews = float32(m.RequestCount) / float32(m.UniqueVisitors)

	r := make(ranking, len(hits))
	i := 0
	for section, hitCount := range hits {
		r[i] = Rank{hitCount, section}
		i++
	}

	sort.Sort(r)
	m.Rank = r
	return m
}

func getSection(resource []byte) []byte {

	start := 0
	for start < len(resource) {

		r, s := utf8.DecodeRune(resource[start:])

		if r == '/' {
			start += s
		} else {
			break
		}
	}

	if start == len(resource) {
		return nil
	}

	for i := start; i < len(resource); {

		r, s := utf8.DecodeRune(resource[i:])

		if r == '/' || r == '#' || r == '?' {

			return resource[start:i]
		}
		i += s
	}

	return resource[start:]
}
