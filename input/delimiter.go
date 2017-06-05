package input

import (
	"io"
	"regexp"
	"strings"
)

const (
	SYSLOG_DELIMITER = `<[0-9]{1,3}>[0-9]\s`
)

var syslogRegex *regexp.Regexp
var startRegex *regexp.Regexp
var runRegex *regexp.Regexp

// Reader is the interface objects passed to the Delimiter must support.
type Reader interface {
	ReadByte() (byte, error)
}

func init() {
	syslogRegex = regexp.MustCompile(SYSLOG_DELIMITER)
	startRegex = regexp.MustCompile(SYSLOG_DELIMITER + `$`)
	runRegex = regexp.MustCompile(`\n` + SYSLOG_DELIMITER)
}

// A Delimiter detects when Syslog lines start.
type Delimiter struct {
	buffer []byte
	regex  *regexp.Regexp
}

// NewDelimiter returns an initialized Delimiter.
func NewDelimiter(maxSize int) *Delimiter {
	d := &Delimiter{}
	d.buffer = make([]byte, 0, maxSize)
	d.regex = startRegex
	return d
}

// Push a byte into the Delimiter. If the byte results in a
// a new Syslog message, it'll be flagged via the bool.
func (d *Delimiter) Push(b byte) (string, bool) {
	d.buffer = append(d.buffer, b)
	delimiter := d.regex.FindIndex(d.buffer)
	if delimiter == nil {
		return "", false
	}

	if d.regex == startRegex {
		// First match -- switch to the regex for embedded lines, and
		// drop any leading characters.
		d.buffer = d.buffer[delimiter[0]:]
		d.regex = runRegex
		return "", false
	}

	dispatch := strings.TrimRight(string(d.buffer[:delimiter[0]]), "\r")
	d.buffer = d.buffer[delimiter[0]+1:]
	return dispatch, true
}

// Vestige returns the bytes which have been pushed to Delimiter, since
// the last Syslog message was returned, but only if the buffer appears
// to be a valid syslog message.
func (d *Delimiter) Vestige() (string, bool) {
	delimiter := syslogRegex.FindIndex(d.buffer)
	if delimiter == nil {
		d.buffer = nil
		return "", false
	}
	dispatch := strings.TrimRight(string(d.buffer), "\r\n")
	d.buffer = nil
	return dispatch, true
}

// Stream returns a channel, on which the delimited Syslog messages
// are emitted.
func (d *Delimiter) Stream(reader Reader) chan string {
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

			event, match := d.Push(b)
			if match {
				eventChan <- event
			}
		}
	}()
	return eventChan
}
