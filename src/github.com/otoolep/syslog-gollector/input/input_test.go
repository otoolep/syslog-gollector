package input

import (
	"strings"
	"testing"

	. "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) {
	TestingT(t)
}

type InputSuite struct{}

var _ = Suite(&InputSuite{})

/*
 * Delimiter tests.
 */

func (s *InputSuite) Test_Simple(c *C) {
	line := "<11>1 sshd is down\n<22>1 sshd is up\n<67>2 password accepted"
	d := NewDelimiter(256)
	ch := d.Stream(strings.NewReader(line))
	c.Assert(<-ch, Equals, "<11>1 sshd is down")
	c.Assert(<-ch, Equals, "<22>1 sshd is up")
}

func (s *InputSuite) Test_Leading(c *C) {
	line := "password accepted for user root<12>1 sshd is down\n<145>1 sshd is up\n<67>2 password accepted"
	d := NewDelimiter(256)
	ch := d.Stream(strings.NewReader(line))

	c.Assert(<-ch, Equals, "<12>1 sshd is down")
	c.Assert(<-ch, Equals, "<145>1 sshd is up")
}

func (s *InputSuite) Test_CRLF(c *C) {
	line := "<12>1 sshd is down\r\n<145>1 sshd is up\r\n<67>2 password accepted"
	d := NewDelimiter(256)
	ch := d.Stream(strings.NewReader(line))

	c.Assert(<-ch, Equals, "<12>1 sshd is down")
	c.Assert(<-ch, Equals, "<145>1 sshd is up")
}

func (s *InputSuite) Test_Stacktrace(c *C) {
	line := "<12>1 sshd is down\n<145>1 OOM on line 42, dummy.java\n\tclass_loader.jar\n<67>2 password accepted"
	d := NewDelimiter(256)
	ch := d.Stream(strings.NewReader(line))

	c.Assert(<-ch, Equals, "<12>1 sshd is down")
	c.Assert(<-ch, Equals, "<145>1 OOM on line 42, dummy.java\n\tclass_loader.jar")
}

func (s *InputSuite) Test_Embedded(c *C) {
	line := "<12>1 sshd is <down>\n<145>1 sshd is up<33>4\n<67>2 password accepted"
	d := NewDelimiter(256)
	ch := d.Stream(strings.NewReader(line))

	c.Assert(<-ch, Equals, "<12>1 sshd is <down>")
	c.Assert(<-ch, Equals, "<145>1 sshd is up<33>4")
}

func (s *InputSuite) Test_VestigeZero(c *C) {
	d := NewDelimiter(256)
	m, b := d.Vestige()
	c.Assert(b, Equals, false)
	c.Assert(m, Equals, "")
}

func (s *InputSuite) Test_Vestige(c *C) {
	d := NewDelimiter(256)
	d.Push('1')
	d.Push('2')
	d.Push('\n')
	m, b := d.Vestige()
	c.Assert(b, Equals, true)
	c.Assert(m, Equals, "12")
}

/*
 * Rfc5424 parer tests
 */

func (s *InputSuite) Test_SuccessfulParsing(c *C) {
	p := NewRfc5424Parser()

	m := p.Parse("<134>1 2013-09-04T10:25:52.618085 ubuntu sshd 1999 - password accepted")
	e := ParsedMessage{Priority: 134, Version: 1, Timestamp: "2013-09-04T10:25:52.618085", Host: "ubuntu", App: "sshd", Pid: 1999, MsgId: "-", Message: "password accepted"}
	c.Assert(*m, Equals, e)

	m = p.Parse("<33>5 2013-09-04T10:25:52.618085 test.com cron 304 - password accepted")
	e = ParsedMessage{Priority: 33, Version: 5, Timestamp: "2013-09-04T10:25:52.618085", Host: "test.com", App: "cron", Pid: 304, MsgId: "-", Message: "password accepted"}
	c.Assert(*m, Equals, e)

	m = p.Parse("<1>0 2013-09-04T10:25:52.618085 test.com cron 65535 - password accepted")
	e = ParsedMessage{Priority: 1, Version: 0, Timestamp: "2013-09-04T10:25:52.618085", Host: "test.com", App: "cron", Pid: 65535, MsgId: "-", Message: "password accepted"}
	c.Assert(*m, Equals, e)

	m = p.Parse("<1>0 2013-09-04T10:25:52.618085 test.com cron 65535 msgid1234 password accepted")
	e = ParsedMessage{Priority: 1, Version: 0, Timestamp: "2013-09-04T10:25:52.618085", Host: "test.com", App: "cron", Pid: 65535, MsgId: "msgid1234", Message: "password accepted"}
	c.Assert(*m, Equals, e)

	m = p.Parse("<1>0 2013-09-04T10:25:52.618085 test.com cron 65535 - JVM NPE\nsome_file.java:48\n\tsome_other_file.java:902")
	e = ParsedMessage{Priority: 1, Version: 0, Timestamp: "2013-09-04T10:25:52.618085", Host: "test.com", App: "cron", Pid: 65535, MsgId: "-", Message: "JVM NPE\nsome_file.java:48\n\tsome_other_file.java:902"}
	c.Assert(*m, Equals, e)
}

func (s *InputSuite) Test_FailedParsing(c *C) {
	p := NewRfc5424Parser()

	m := p.Parse("<134> 2013-09-04T10:25:52.618085 ubuntu sshd 1999 - password accepted")
	c.Assert(m, IsNil)

	m = p.Parse("<33> 7 2013-09-04T10:25:52.618085 test.com cron 304 - password accepted")
	c.Assert(m, IsNil)

	m = p.Parse("<33> 7 2013-09-04T10:25:52.618085 test.com cron 304 $ password accepted")
	c.Assert(m, IsNil)

	m = p.Parse("<33> 7 2013-09-04T10:25:52.618085 test.com cron 304 - - password accepted")
	c.Assert(m, IsNil)

	m = p.Parse("<33>7 2013-09-04T10:25:52.618085 test.com cron not_a_pid - password accepted")
	c.Assert(m, IsNil)

	m = p.Parse("5:52.618085 test.com cron 65535 - password accepted")
	c.Assert(m, IsNil)
}
