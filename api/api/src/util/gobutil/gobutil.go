package gobutil

import (
	"bytes"
	"encoding/gob"
	"sync"
)

// GobCodec encodes and decodes values of type T using the gob package.
type GobCodec[T any] struct {
	buff *bytes.Buffer
	m    *sync.Mutex
}

// NewGobCodec creates a GobCodec for the specified type.
func NewGobCodec[T any]() *GobCodec[T] {
	return &GobCodec[T]{
		buff: &bytes.Buffer{},
		m:    &sync.Mutex{},
	}
}

// Encode a value of type T into a slice of bytes.
func (ed *GobCodec[T]) Encode(val T) ([]byte, error) {
	ed.m.Lock()
	defer ed.m.Unlock()
	defer ed.buff.Reset()

	enc := gob.NewEncoder(ed.buff)
	err := enc.Encode(val)
	src := ed.buff.Bytes()
	dst := make([]byte, len(src))
	copy(dst, src)

	return dst, err
}

// Decode a slice of bytes into a value of type T.
func (ed *GobCodec[T]) Decode(raw []byte) (T, error) {
	ed.m.Lock()
	defer ed.m.Unlock()
	defer ed.buff.Reset()

	dec := gob.NewDecoder(ed.buff)
	ed.buff.Write(raw)
	var v T
	err := dec.Decode(&v)

	return v, err
}
