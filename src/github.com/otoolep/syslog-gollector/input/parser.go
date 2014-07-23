package input

import (
	"encoding/json"
	"regexp"
	"strconv"
)

// A Rfc5424Parser parses Syslog messages.
type Rfc5424Parser struct {
	regex *regexp.Regexp
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

// Returns an initialized Rfc5424Parser.
func NewRfc5424Parser() *Rfc5424Parser {
	p := &Rfc5424Parser{}
	r := regexp.MustCompile(`(?s)<([0-9]{1,3})>([0-9])\s(.+)\s(.+)\s(.+)\s([0-9]{1,5})\s([\w-]+)\s(.+$)`)
	p.regex = r
	return p
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
		return nil
	}

	// Errors are ignored, because the regex shouldn't match if the
	// following ain't numbers.
	pri, _ := strconv.Atoi(m[1])
	ver, _ := strconv.Atoi(m[2])
	pid, _ := strconv.Atoi(m[6])

	return &ParsedMessage{pri, ver, m[3], m[4], m[5], pid, m[7], m[8]}
}
