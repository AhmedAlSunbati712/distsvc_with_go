package log

import (
	"io"
	"os"

	"github.com/tysonmote/gommap"
)

var (
	offWidth uint64 = 4
	posWidth uint64 = 8
	entWidth        = (offWidth + posWidth)
)

type index struct {
	f    *os.File
	mmap gommap.MMap
	size uint64
}

func newIndex(file *os.File, c Config) (*index, error) {
	idx := &index{
		f: file,
	}

	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	idx.size = uint64(stat.Size())
	if err := os.Truncate(file.Name(), int64(c.Segment.MaxIndexBytes)); err != nil {
		return nil, err
	}

	idx.mmap, err = gommap.Map(file.Fd(), gommap.PROT_READ|gommap.PROT_WRITE, gommap.MAP_SHARED)
	if err != nil {
		return nil, err
	}
	return idx, nil
}

func (idx *index) Read(in int64) (out uint32, pos uint64, err error) {
	if idx.size == 0 {
		return 0, 0, io.EOF
	}

	// Get the relative index from the top of the index file
	var entry_idx uint32
	if in == -1 {
		entry_idx = uint32((idx.size / entWidth) - 1)
	} else {
		entry_idx = uint32(in)
	}

	// Now calculate the byte offset
	var entryOffset uint64 = uint64(uint64(entry_idx) * entWidth)
	if idx.size < entryOffset+entWidth {
		return 0, 0, io.EOF
	}

	// Now read the actual record offset
	out = enc.Uint32(idx.mmap[entryOffset : entryOffset+offWidth])
	pos = enc.Uint64(idx.mmap[entryOffset+offWidth : entryOffset+entWidth])
	return out, pos, nil

}

func (idx *index) Write(offset uint32, pos uint64) error {
	if uint64(len(idx.mmap)) < idx.size+entWidth {
		return io.EOF
	}
	enc.PutUint32(idx.mmap[idx.size:idx.size+offWidth], offset)
	enc.PutUint64(idx.mmap[idx.size+offWidth:idx.size+entWidth], pos)
	idx.size += entWidth
	return nil
}

func (idx *index) Name() string {
	return idx.f.Name()
}

func (idx *index) Close() error {
	if err := idx.mmap.Sync(gommap.MS_ASYNC); err != nil {
		return err
	}
	if err := idx.f.Sync(); err != nil {
		return err
	}
	if err := idx.f.Truncate(int64(idx.size)); err != nil {
		return err
	}
	return idx.f.Close()
}
