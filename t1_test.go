package bstream

import (
	"testing"
)

//----------------------------------------------------------------------------------------------------------------------------//

func Test1(t *testing.T) {
	// TODO
}

//----------------------------------------------------------------------------------------------------------------------------//

func TestPutGet(t *testing.T) {
	s := New(0)

	data := []any{
		false,
		true,
		int64(0),
		int64(-1),
		int64(0x0102030405060708),
		int64(0x1020304050607080),
		float64(0),
		float64(123456789e39),
		float64(-123456789e39),
		"",
		"qwertyuiopйцукенгшщй世界",
		"1111111111222222222233333333334444444444555555555566666666667777777777888888888899999999990000000000qqqqqqqqqqwwwwwwwwwweeeeeeeeeerrrrrrrrrr",
	}

	for i, v := range data {
		switch v := v.(type) {
		case bool:
			s.PutBool(v)
		case int64:
			s.PutInt(v)
		case float64:
			s.PutFloat(v)
		case string:
			s.PutString(v)
		default:
			t.Errorf(`[%d] Illegal type "%T"`, i, v)
		}
	}

	for i, v := range data {
		switch v := v.(type) {
		case bool:
			vv, err := s.GetBool()
			if err != nil {
				t.Error(err)
			} else if vv != v {
				t.Errorf("[%d] Got %v, expected %v", i, vv, v)
			}
		case int64:
			vv, err := s.GetInt()
			if err != nil {
				t.Error(err)
			} else if vv != v {
				t.Errorf("[%d] Got %v, expected %v", i, vv, v)
			}
		case float64:
			vv, err := s.GetFloat()
			if err != nil {
				t.Error(err)
			} else if vv != v {
				t.Errorf("[%d] Got %v, expected %v", i, vv, v)
			}
		case string:
			vv, err := s.GetString()
			if err != nil {
				t.Error(err)
			} else if vv != v {
				t.Errorf(`[%d] Got "%v", expected "%v"`, i, vv, v)
			}
		}
	}
}

//----------------------------------------------------------------------------------------------------------------------------//

func BenchmarkPutBool(b *testing.B) {
	s := New(b.N)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s.PutBool(true)
	}
}

func BenchmarkGetBool(b *testing.B) {
	s := New(b.N)

	for i := 0; i < b.N; i++ {
		s.PutBool(true)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := s.GetBool()
		if err != nil {
			b.Fatal(err)
		}
	}
}

//----------------------------------------------------------------------------------------------------------------------------//

func BenchmarkPutByte(b *testing.B) {
	s := New(b.N)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s.PutByte(0)
	}
}

func BenchmarkGetByte(b *testing.B) {
	s := New(b.N)

	for i := 0; i < b.N; i++ {
		s.PutByte(0)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := s.GetByte()
		if err != nil {
			b.Fatal(err)
		}
	}
}

//----------------------------------------------------------------------------------------------------------------------------//

func BenchmarkPutInt64(b *testing.B) {
	s := New(b.N * 8)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s.PutInt(0)
	}
}

func BenchmarkGetInt64(b *testing.B) {
	s := New(b.N * 8)

	for i := 0; i < b.N; i++ {
		s.PutInt(0)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := s.GetInt()
		if err != nil {
			b.Fatal(err)
		}
	}
}

//----------------------------------------------------------------------------------------------------------------------------//

func BenchmarkPutFloat64(b *testing.B) {
	s := New(b.N * 8)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s.PutFloat(0)
	}
}

func BenchmarkGetFloat64(b *testing.B) {
	s := New(b.N * 8)

	for i := 0; i < b.N; i++ {
		s.PutFloat(0)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := s.GetFloat()
		if err != nil {
			b.Fatal(err)
		}
	}
}

//----------------------------------------------------------------------------------------------------------------------------//

func BenchmarkPutString(b *testing.B) {
	ln := 128
	s := New(b.N * (ln + 4))

	bb := make([]byte, ln)
	for i := 0; i < len(bb); i++ {
		bb[i] = byte(i)
	}

	ss := string(bb)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s.PutString(ss)
	}
}

func BenchmarkGetString(b *testing.B) {
	ln := 128

	s := New(b.N * (ln + 4))

	bb := make([]byte, ln)
	for i := 0; i < len(bb); i++ {
		bb[i] = byte(i)
	}

	ss := string(bb)

	for i := 0; i < b.N; i++ {
		s.PutString(ss)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := s.GetString()
		if err != nil {
			b.Fatalf("[%d] B.N=%d, len=%d, cap=%d, readPos=%d %s", i, b.N, len(s.buf), cap(s.buf), s.readPos, err)
		}
	}
}

//----------------------------------------------------------------------------------------------------------------------------//

func BenchmarkMarshal(b *testing.B) {
	types, data := makeTestData()
	blockCount := len(data)

	s := make([]*Stream, b.N)

	for i := 0; i < b.N; i++ {
		s[i] = New(blockCount * (1*1 + 1*1 + 6*8 + (4 + 0) + (4 + 36) + (4 + 140) + 40*8))
		//s[i] = New(0)

	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := s[i].Marshal(types, data)
		if err != nil {
			b.Fatalf("[%d] %s", i, err)
		}
	}
}

//----------------------------------------------------------------------------------------------------------------------------//

func BenchmarkUnmarshal(b *testing.B) {
	types, data := makeTestData()

	s := New(0)

	err := s.Marshal(types, data)
	if err != nil {
		b.Fatalf("[Marshal] %s", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := s.Unmarshal(types)
		if err != nil {
			b.Fatalf("[%d] %s", i, err)
		}
	}
}

//----------------------------------------------------------------------------------------------------------------------------//

func makeTestData() (types []Type, data [][]any) {
	blockCount := 1000
	types = []Type{
		Bool,
		Byte,
		Int, Int, Int, Int, Int, Int,
		String, String, String,
		Float, Float, Float, Float, Float, Float, Float, Float, Float, Float,
		Float, Float, Float, Float, Float, Float, Float, Float, Float, Float,
		Float, Float, Float, Float, Float, Float, Float, Float, Float, Float,
		Float, Float, Float, Float, Float, Float, Float, Float, Float, Float,
	}

	data = make([][]any, blockCount)

	for i1 := 0; i1 < blockCount; i1++ {
		block := []any{
			i1%2 == 0,
			byte(i1 % 256),
			int(i1),
			uint(i1),
			int32(i1),
			uint32(i1),
			int64(i1),
			uint64(i1),
			"",
			"qwertyuiopйцукенгшщй世界",
			"1111111111222222222233333333334444444444555555555566666666667777777777888888888899999999990000000000qqqqqqqqqqwwwwwwwwwweeeeeeeeeerrrrrrrrrr",
		}

		for i2 := 0; i2 < 20; i2++ {
			block = append(block, float32(i2))
			block = append(block, float64(i2))
		}

		data[i1] = block
	}

	return
}

//----------------------------------------------------------------------------------------------------------------------------//
