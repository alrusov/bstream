package bstream

import (
	"fmt"
	"reflect"
	"unsafe"
)

//----------------------------------------------------------------------------------------------------------------------------//

type (
	// Stream --
	Stream struct {
		buf      []byte
		capacity int
		readPos  int
	}

	// Type --
	Type int
)

const (
	// Bool --
	Bool Type = iota
	// Byte --
	Byte
	// Int --
	Int
	// Float --
	Float
	// String --
	String
)

//----------------------------------------------------------------------------------------------------------------------------//

// New --
//func New(withLock bool, capacity int) *Stream {
func New(capacity int) *Stream {
	s := &Stream{
		buf:      make([]byte, 0, capacity),
		capacity: capacity,
		readPos:  0,
	}
	return s
}

//----------------------------------------------------------------------------------------------------------------------------//

// Reset --
func (s *Stream) Reset() {
	s.buf = make([]byte, 0, s.capacity)
	s.readPos = 0
}

// Len --
func (s *Stream) Len() int {
	return len(s.buf)
}

// ReadPos --
func (s *Stream) ReadPos() int {
	return s.readPos
}

// ResetReadPos --
func (s *Stream) ResetReadPos() {
	s.SetReadPos(0)
}

// SetReadPos --
func (s *Stream) SetReadPos(pos int) {
	s.readPos = pos
}

//----------------------------------------------------------------------------------------------------------------------------//

// PutBool --
func (s *Stream) PutBool(v bool) {
	b := byte(0)
	if v {
		b = byte(1)
	}
	s.buf = append(s.buf, b)
}

// GetBool --
func (s *Stream) GetBool() (bool, error) {
	const ln = 1
	if len(s.buf)-s.readPos < ln {
		return false, fmt.Errorf("GetBool: Requires %d bytes, has %d", ln, len(s.buf)-s.readPos)
	}

	b := s.buf[s.readPos]
	s.readPos += ln
	return b != 0, nil
}

//----------------------------------------------------------------------------------------------------------------------------//

// PutByte --
func (s *Stream) PutByte(v byte) {
	s.buf = append(s.buf, v)
}

// GetByte --
func (s *Stream) GetByte() (byte, error) {
	const ln = 1
	if len(s.buf)-s.readPos < ln {
		return 0, fmt.Errorf("GetByte: Requires %d bytes, has %d", ln, len(s.buf)-s.readPos)
	}

	b := s.buf[s.readPos]
	s.readPos += ln
	return b, nil
}

//----------------------------------------------------------------------------------------------------------------------------//

// PutInt --
func (s *Stream) PutInt(v int64) {
	const ln = 8
	b := *(*[ln]byte)(unsafe.Pointer(&v))
	s.buf = append(s.buf, b[:ln]...)
}

// GetInt --
func (s *Stream) GetInt() (int64, error) {
	const ln = 8
	if len(s.buf)-s.readPos < ln {
		return 0, fmt.Errorf("GetInt64: Requires %d bytes, has %d", ln, len(s.buf)-s.readPos)
	}

	v := *(*int64)(unsafe.Pointer(&s.buf[s.readPos]))
	s.readPos += ln
	return v, nil
}

//----------------------------------------------------------------------------------------------------------------------------//

// PutFloat --
func (s *Stream) PutFloat(v float64) {
	const ln = 8
	b := *(*[ln]byte)(unsafe.Pointer(&v))
	s.buf = append(s.buf, b[:ln]...)
}

// GetFloat --
func (s *Stream) GetFloat() (float64, error) {
	const ln = 8
	if len(s.buf)-s.readPos < ln {
		return 0, fmt.Errorf("GetFloat64: Requires %d bytes, has %d", ln, len(s.buf)-s.readPos)
	}

	v := *(*float64)(unsafe.Pointer(&s.buf[s.readPos]))
	s.readPos += ln
	return v, nil
}

//----------------------------------------------------------------------------------------------------------------------------//

// PutString --
func (s *Stream) PutString(v string) {
	const lnln = 4
	ln := len(v)
	ln32 := int32(ln)
	b := *(*[lnln]byte)(unsafe.Pointer(&ln32))
	s.buf = append(s.buf, b[:lnln]...)

	if ln == 0 {
		return
	}

	/*
		// Простой, надежный, тупой и медленный вариант. Преобразование строки в []byte относительно долгая операция и с дополнительным выделением памяти
		s.buf = append(s.buf, []byte(v)...)
	*/
	/*
		// Нет выигрыша при предварительном выделении памяти. Если без выделения, то около 10%
		buf := new(bytes.Buffer)
		buf.WriteString(v)

		s.buf = append(s.buf, buf.Bytes()...)
	*/

	// Быстрее примерно в три раза (если в New() указать достаточный размер, иначе выигрыш 35%) так как не копируются данные.
	// Но стрёмно... Надо на винде проверить...
	bb := *(*[]byte)(
		unsafe.Pointer(
			&reflect.SliceHeader{
				Data: (*reflect.StringHeader)(unsafe.Pointer(&v)).Data,
				Len:  ln,
				Cap:  ln,
			},
		),
	)

	s.buf = append(s.buf, bb...)
}

// GetString --
func (s *Stream) GetString() (string, error) {
	const lnln = 4
	tailLen := len(s.buf) - s.readPos
	if tailLen < lnln {
		return "", fmt.Errorf("GetString: Requires %d bytes at least, has %d", lnln, tailLen)
	}

	v := ""
	ln := *(*uint32)(unsafe.Pointer(&s.buf[s.readPos]))
	if ln > 0 {
		if tailLen < int(lnln+ln) {
			return "", fmt.Errorf("GetString: Requires %d bytes, has %d", lnln+ln, tailLen)
		}

		v = *(*string)(
			unsafe.Pointer(
				&reflect.StringHeader{
					Data: uintptr(unsafe.Pointer(&s.buf[s.readPos+lnln])),
					Len:  int(ln),
				},
			),
		)
	}

	s.readPos += int(lnln + ln)
	return v, nil
}

//----------------------------------------------------------------------------------------------------------------------------//

// Marshal --
func (s *Stream) Marshal(types []Type, data [][]any) (err error) {
	fn := "Marshal"

	setError := func(bi int, i int, v any, vv any) {
		err = fmt.Errorf(`%s: data[%d][%d]="%v" is %T, expected %T`, fn, bi, i, v, v, vv)
	}

	for bi, block := range data {
		if len(block) != len(types) {
			err = fmt.Errorf("%s: %d (len(block)=%d) != (len(types)=%d)", fn, bi, len(block), len(types))
			return
		}

		for i, v := range block {
			switch types[i] {
			case Bool:
				var vv bool
				switch v := v.(type) {
				case bool:
					vv = v
				default:
					setError(bi, i, v, vv)
					return
				}
				s.PutBool(vv)
			case Int:
				var vv int64
				switch v := v.(type) {
				case int:
					vv = int64(v)
				case int32:
					vv = int64(v)
				case int64:
					vv = v
				case uint:
					vv = int64(v)
				case uint32:
					vv = int64(v)
				case uint64:
					vv = int64(v)
				default:
					setError(bi, i, v, vv)
					return
				}
				s.PutInt(vv)
			case Float:
				var vv float64
				switch v := v.(type) {
				case float32:
					vv = float64(v)
				case float64:
					vv = v
				default:
					setError(bi, i, v, vv)
					return
				}
				s.PutFloat(vv)
			case String:
				var vv string
				switch v := v.(type) {
				case string:
					vv = v
				default:
					setError(bi, i, v, vv)
					return
				}
				s.PutString(vv)
			default:
			}
		}
	}

	return
}

//----------------------------------------------------------------------------------------------------------------------------//

// Unmarshal --
func (s *Stream) Unmarshal(types []Type) (data [][]any, err error) {
	//fn := "Unmarshal"

	defer func() {
		if err != nil {
			data = nil
		}
	}()

	s.ResetReadPos()

	data = [][]any{}
	blockLen := len(types)
	ln := s.Len()

	for {
		block := make([]any, blockLen)

		for i, t := range types {
			switch t {
			case Bool:
				block[i], err = s.GetBool()
			case Int:
				block[i], err = s.GetInt()
			case Float:
				block[i], err = s.GetFloat()
			case String:
				block[i], err = s.GetString()
			}

			if err != nil {
				return
			}
		}

		data = append(data, block)

		if s.ReadPos() == ln {
			break
		}
	}

	return
}

//----------------------------------------------------------------------------------------------------------------------------//
