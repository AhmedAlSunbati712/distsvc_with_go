package log

import (
	"fmt"
	"os"
	"strconv"
	"strings"
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
	files, err := os.ReadDir((l.Dir))
	if err != nil {
		return err
	}
	offsets := make([]uint64, 10)
	for _, file := range files {
		parts := strings.Split(file.Name(), ".")
		file_name, file_extension := parts[0], parts[1]
		if file_extension == ".store" {
			file_base_offset, err := strconv.ParseUint(file_name, 10, 0)
			if err != nil {
				return err
			}
			offsets = append(offsets, file_base_offset)
		}
	}

	for _, offset := range offsets {
		l.newSegment(offset)
	}
	return fmt.Errorf("Some placeholder")
}

func (l *Log) newSegment(offset uint64) error {
	segment, err := newSegment(l.Dir, offset, l.c)
	if err != nil {
		return err
	}
	l.activeSegment = segment
	l.segments = append(l.segments, segment)
	return nil
}
