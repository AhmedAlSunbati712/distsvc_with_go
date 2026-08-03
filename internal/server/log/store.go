package log

import (
	"bufio"
	"encoding/binary"
	"os"
	"sync"
)

var (
	enc = binary.BigEndian
)

const (
	lenWidth = 8
)

type store struct {
	mu   sync.Mutex
	f    *os.File
	buf  *bufio.Writer
	size uint64
}

func newStore(f *os.File) (*store, error) {
	s := &store{}
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	s.buf = bufio.NewWriter(f)
	s.size = uint64(info.Size())
	s.f = f
	return s, nil
}

func (s *store) Append(record []byte) (n uint64, pos uint64, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	// First, get the length of the record
	// Write it to the file with binary.Write(buf, enc, size)
	// Then write the record itself
	// Make sure to grab the file size first since that will give you the offset
	offset := s.size

	if err := binary.Write(s.buf, enc, uint64(len(record))); err != nil {
		return 0, 0, err
	}
	w, err := s.buf.Write(record)
	if err != nil {
		return 0, 0, err
	}
	w += lenWidth
	s.size += uint64(w)
	return uint64(w), offset, nil
}

func (s *store) Read(pos uint64) (record []byte, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Make sure to flush the buffer in case we are trying to read from somewhere in the buffer
	// that is not yet on disk
	if err := s.buf.Flush(); err != nil {
		return nil, err
	}

	// Read the size of the record, first lenWidth bytes
	size_b := make([]byte, lenWidth)
	if _, err := s.f.ReadAt(size_b, int64(pos)); err != nil {
		return nil, err
	}
	// Convert it into big endian representation of uint64
	size := enc.Uint64(size_b)

	// Now read the record itself
	record_b := make([]byte, size)
	if _, err := s.f.ReadAt(record_b, int64(pos+lenWidth)); err != nil {
		return nil, err
	}

	return record_b, nil
}

func (s *store) ReadAt(b []byte, off int64) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.buf.Flush(); err != nil {
		return 0, err
	}
	return s.f.ReadAt(b, off)
}

func (s *store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.buf.Flush(); err != nil {
		return err
	}
	if err := s.f.Close(); err != nil {
		return err
	}
	return nil
}

func (s *store) Name() string {
	return s.f.Name()
}
