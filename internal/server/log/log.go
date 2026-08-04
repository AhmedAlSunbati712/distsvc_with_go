package log

import (
	"fmt"
	"sync"
)

type Log struct {
	mu            sync.Mutex
	Dir           string
	c             Config
	activeSegment *segment
	segments      []*segment
}

func newLog(dir string, c Config) (*Log, error) {
	if c.Segment.MaxIndexBytes == 0 {
		c.Segment.MaxIndexBytes = 1024
	}
	if c.Segment.MaxStoreBytes == 0 {
		c.Segment.MaxStoreBytes = 1024
	}
	l := &Log{
		Dir: dir,
		c:   c,
	}
	return l, l.setup()
}

func (l *Log) setup() error {
	return fmt.Errorf("Some placeholder")
}
