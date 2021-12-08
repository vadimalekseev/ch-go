package proto

import (
	"bufio"
	"encoding/binary"
	"io"
	"unicode/utf8"

	"github.com/go-faster/errors"
)

// Decoder implements decoding from Reader.
type Decoder interface {
	Decode(r *Reader) error
}

// AwareDecoder implements encoding to Buffer that depends on version.
type AwareDecoder interface {
	DecodeAware(r *Reader, version int) error
}

// Reader implements ClickHouse protocol decoding from buffered reader.
type Reader struct {
	s *bufio.Reader
	b *Buffer
}

// Decode value.
func (r *Reader) Decode(v Decoder) error {
	return v.Decode(r)
}

// ReadRaw reads raw n bytes.
func (r *Reader) ReadRaw(n int) ([]byte, error) {
	b := make([]byte, n)

	if _, err := io.ReadFull(r.s, b); err != nil {
		return nil, errors.Wrap(err, "read full")
	}

	return b, nil
}

// UVarInt reads uint64 from internal reader.
func (r *Reader) UVarInt() (uint64, error) {
	n, err := binary.ReadUvarint(r.s)
	if err != nil {
		return 0, errors.Wrap(err, "read")
	}
	return n, nil
}

const maxStrSize = 10 * 1024 * 1024 // 10 MB

// StrRaw decodes string to internal buffer and returns it directly.
//
// Do not retain returned slice.
func (r *Reader) StrRaw() ([]byte, error) {
	n, err := r.Int()
	if err != nil {
		return nil, errors.Wrap(err, "read length")
	}

	if n < 0 {
		return nil, errors.Errorf("size %d is invalid", n)
	}
	if n > maxStrSize {
		// Protecting from possible OOM.
		return nil, errors.Errorf("size %d too big (%d is maximum)", n, maxStrSize)
	}

	r.b.Ensure(n)
	if _, err := io.ReadFull(r.s, r.b.Buf); err != nil {
		return nil, errors.Wrap(err, "read str")
	}

	return r.b.Buf, nil
}

// StrAppend decodes string and appends it to provided buf.
func (r *Reader) StrAppend(buf []byte) ([]byte, error) {
	defer r.b.Reset()

	str, err := r.StrRaw()
	if err != nil {
		return nil, errors.Wrap(err, "raw")
	}

	return append(buf, str...), nil
}

// StrBytes decodes string and allocates new byte slice with result.
func (r *Reader) StrBytes() ([]byte, error) {
	return r.StrAppend(nil)
}

// Str decodes string.
func (r *Reader) Str() (string, error) {
	s, err := r.StrBytes()
	if err != nil {
		return "", errors.Wrap(err, "bytes")
	}
	if !utf8.Valid(s) {
		return "", errors.New("invalid utf8")
	}

	return string(s), err
}

// Int decodes uvarint as int.
func (r *Reader) Int() (int, error) {
	n, err := r.UVarInt()
	if err != nil {
		return 0, errors.Wrap(err, "uvarint")
	}
	return int(n), nil
}

// Int32 decodes int32 value.
func (r *Reader) Int32() (int32, error) {
	r.b.Ensure(32 / 8)
	if _, err := io.ReadFull(r.s, r.b.Buf); err != nil {
		return 0, errors.Wrap(err, "read")
	}
	v := bin.Uint32(r.b.Buf)
	return int32(v), nil
}

// Int64 decodes int64 value.
func (r *Reader) Int64() (int64, error) {
	r.b.Ensure(64 / 8)
	if _, err := io.ReadFull(r.s, r.b.Buf); err != nil {
		return 0, errors.Wrap(err, "read")
	}
	v := bin.Uint64(r.b.Buf)
	return int64(v), nil
}

// UInt8 decodes uint8 value.
func (r *Reader) UInt8() (uint8, error) {
	r.b.Ensure(1)
	if _, err := io.ReadFull(r.s, r.b.Buf); err != nil {
		return 0, errors.Wrap(err, "read")
	}
	return r.b.Buf[0], nil
}

// UInt32 decodes uint32 value.
func (r *Reader) UInt32() (uint32, error) {
	r.b.Ensure(32 / 8)
	if _, err := io.ReadFull(r.s, r.b.Buf); err != nil {
		return 0, errors.Wrap(err, "read")
	}
	v := bin.Uint32(r.b.Buf)
	return v, nil
}

// Bool decodes bool as uint8.
func (r *Reader) Bool() (bool, error) {
	v, err := r.UInt8()
	if err != nil {
		return false, errors.Wrap(err, "uint8")
	}
	switch v {
	case boolTrue:
		return true, nil
	case boolFalse:
		return false, nil
	default:
		return false, errors.Errorf("unexpected value %d for boolean", v)
	}
}

const defaultReaderSize = 1024 // 1kb

// NewReader initializes new Reader from provided io.Reader.
func NewReader(r io.Reader) *Reader {
	return &Reader{
		s: bufio.NewReaderSize(r, defaultReaderSize),
		b: &Buffer{},
	}
}
