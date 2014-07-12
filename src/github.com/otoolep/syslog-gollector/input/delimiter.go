package input

import (
	"io"
	"regexp"
	"strings"
)

const (
	SYSLOG_DELIMITER = `<[0-9]{1,3}>[0-9]\s$`
)

var startRegex *regexp.Regexp
var runRegex *regexp.Regexp

type Reader interface {
	ReadByte() (byte, error)
}

func init() {
	startRegex = regexp.MustCompile(SYSLOG_DELIMITER)
	runRegex = regexp.MustCompile(`\n` + SYSLOG_DELIMITER)
}

// A Delimiter detects when Syslog lines start.
type Delimiter struct {
	buffer []byte
	regex  *regexp.Regexp
}

// NewDelimiter returns an initialized Delimiter.
func NewDelimiter(maxSize int) *Delimiter {
	self := &Delimiter{}
	self.buffer = make([]byte, 0, maxSize)
	self.regex = startRegex
	return self
}

// Push a byte into the Delimiter. If the byte results in a
// a new Syslog message, it'll be flagged via the bool.
func (self *Delimiter) Push(b byte) (string, bool) {
	self.buffer = append(self.buffer, b)
	delimiter := self.regex.FindIndex(self.buffer)
	if delimiter == nil {
		return "", false
	}

	if self.regex == startRegex {
		// First match -- switch to the regex for embedded lines, and
		// drop any leading characters.
		self.buffer = self.buffer[delimiter[0]:]
		self.regex = runRegex
		return "", false
	}

	dispatch := strings.TrimRight(string(self.buffer[:delimiter[0]]), "\r")
	self.buffer = self.buffer[delimiter[0]+1:]
	return dispatch, true
}

// Vestige returns the bytes which have been pushed to Delimiter, since
// the last Syslog message was returned.
func (self *Delimiter) Vestige() (string, bool) {
	if len(self.buffer) == 0 {
		return "", false
	}
	dispatch := strings.TrimRight(string(self.buffer), "\r\n")
	self.buffer = nil
	return dispatch, true
}

// Stream returns a channel, on which the delimited Syslog messages
// are emitted.
func (self *Delimiter) Stream(reader Reader) chan string {
	eventChan := make(chan string)

	go func() {
		for {
			b, err := reader.ReadByte()
			if err != nil {
				if err != io.EOF {
					panic(err)
				} else {
					close(eventChan)
					return
				}
			}

			event, match := self.Push(b)
			if match {
				eventChan <- event
			}
		}
	}()
	return eventChan
}
