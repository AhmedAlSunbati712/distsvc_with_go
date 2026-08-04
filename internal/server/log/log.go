package log

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	log_v1 "github.com/ahmedalsunbati712/proglog/api/v1"
)

type Log struct {
	mu            sync.RWMutex
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
		if err := l.newSegment(offset); err != nil {
			return err
		}
	}

	if l.segments == nil {
		if err := l.newSegment(l.c.Segment.InitialOffset); err != nil {
			return nil
		}

	}
	return nil
}

func (l *Log) Append(record *log_v1.Record) (uint64, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	off, err := l.activeSegment.Append(record)
	if err != nil {
		return 0, nil
	}

	if l.activeSegment.IsMaxed() {
		err = l.newSegment(off + 1)
	}
	return off, err
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

func (l *Log) Read(offset uint64) (*log_v1.Record, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	var target_segment *segment
	for _, segment := range l.segments {
		if segment.baseOffset <= offset && offset < segment.nextOffset {
			target_segment = segment
			break
		}
	}
	if target_segment == nil {
		return nil, fmt.Errorf("Offset out of range: %d", offset)
	}
	return target_segment.Read(offset)
}
