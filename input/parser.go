package input

import (
	"encoding/json"
	"regexp"
	"strconv"

	metrics "github.com/rcrowley/go-metrics"
)

// A Rfc5424Parser parses Syslog messages.
type Rfc5424Parser struct {
	regex    *regexp.Regexp
	registry metrics.Registry
	parsed   metrics.Counter
	dropped  metrics.Counter
}

// ParsedMessage represents a fully parsed Syslog message.
type ParsedMessage struct {
	Priority  int    `json:"priority"`
	Version   int    `json:"version"`
	Timestamp string `json:"timestamp"`
	Host      string `json:"host"`
	App       string `json:"app"`
	Pid       int    `json:"pid"`
	MsgId     string `json:"msgid"`
	Message   string `json:"message"`
}

// NewRfc5424Parser Returns an initialized Rfc5424Parser.
func NewRfc5424Parser() *Rfc5424Parser {
	leading := `(?s)`
	pri := `<([0-9]{1,3})>`
	ver := `([0-9])`
	ts := `([^ ]+)`
	host := `([^ ]+)`
	app := `([^ ]+)`
	pid := `([0-9]{1,5})`
	id := `([\w-]+)`
	msg := `(.+$)`

	p := &Rfc5424Parser{}
	r := regexp.MustCompile(leading + pri + ver + `\s` + ts + `\s` + host + `\s` + app + `\s` + pid + `\s` + id + `\s` + msg)
	p.regex = r

	// Initialize metrics
	p.registry = metrics.NewRegistry()
	p.parsed = metrics.NewCounter()
	p.dropped = metrics.NewCounter()
	p.registry.Register("events.parsed", p.parsed)
	p.registry.Register("events.dropped", p.dropped)
	return p
}

// Statistics returns an object storing statistics, which supports JSON
// marshalling.
func (p *Rfc5424Parser) Statistics() (metrics.Registry, error) {
	return p.registry, nil
}

// StreamingParse emits parsed Syslog messages on the returned channel. If
// there are any parsing errors, the message is dropped.
func (p *Rfc5424Parser) StreamingParse(in chan string) (chan string, error) {
	ch := make(chan string)

	go func() {
		for m := range in {
			parsed := p.Parse(m)
			if parsed == nil {
				continue
			}
			b, err := json.Marshal(*parsed)
			if err != nil {
				continue
			}
			event := string(b)
			ch <- event
		}
	}()
	return ch, nil
}

// Parse takes a raw message and returns a parsed message. If no match,
// nil is returned.
func (p *Rfc5424Parser) Parse(raw string) *ParsedMessage {
	m := p.regex.FindStringSubmatch(raw)
	if m == nil || len(m) != 9 {
		p.dropped.Inc(1)
		return nil
	}
	p.parsed.Inc(1)

	// Errors are ignored, because the regex shouldn't match if the
	// following ain't numbers.
	pri, _ := strconv.Atoi(m[1])
	ver, _ := strconv.Atoi(m[2])
	pid, _ := strconv.Atoi(m[6])

	return &ParsedMessage{pri, ver, m[3], m[4], m[5], pid, m[7], m[8]}
}
