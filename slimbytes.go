package slimarray

import "errors"

var (
	BytesTooLarge = errors.New("total bytes exceeds max value of uint32")
	TooManyRows   = errors.New("row count exceeds max value of int32")
)

// NewBytes creates a SlimBytes from a series of records.
//
// Since 0.1.14
func NewBytes(records [][]byte) (*SlimBytes, error) {

	n := int64(len(records))
	size := int64(0)
	for _, rec := range records {
		size += int64(len(rec))
	}

	if n > 0x7fffffff {
		return nil, TooManyRows
	}

	if size > 0xffffffff {
		return nil, BytesTooLarge
	}

	packed := make([]byte, 0, size)
	pos := make([]uint32, 0, n+1)

	for _, rec := range records {
		pos = append(pos, uint32(len(packed)))
		packed = append(packed, rec...)
	}
	pos = append(pos, uint32(len(packed)))

	posArr := NewU32(pos)
	bs := &SlimBytes{
		Positions: posArr,
		Records:   packed,
	}

	return bs, nil
}

// Get the i-th record.
//
// Performance
// ~ 24 ns
// Two slimarray query cost about 8 ns * 2
//
// Since 0.1.14
func (b *SlimBytes) Get(i int32) []byte {

	byteOffset := b.Positions.Get(i)
	byteEnd := b.Positions.Get(i + 1)

	return b.Records[byteOffset:byteEnd]
}
