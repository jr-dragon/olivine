package resp

import "testing"

func TestSimpleString_Marshal(t *testing.T) {
	testcases := []struct {
		data   []byte
		expect []byte
	}{
		{
			data:   []byte("OK"),
			expect: []byte("+OK\r\n"),
		},
		{
			data:   []byte("PONG"),
			expect: []byte("+PONG\r\n"),
		},
	}

	for _, tc := range testcases {
		t.Run(string(tc.data), func(t *testing.T) {
			s := SimpleString(tc.data)

			marshaled := s.Marshal()
			if string(marshaled) != string(tc.expect) {
				t.Errorf("expect '%s', got '%s'", tc.expect, marshaled)
			}
		})
	}
}

func BenchmarkSimpleString_bufferMarshal(b *testing.B) {
	s := SimpleString("OK")

	b.StartTimer()
	for b.Loop() {
		_ = s.bufferMarshal()
	}
	b.StopTimer()
}

func BenchmarkSimpleString_copyMarshal(b *testing.B) {
	s := SimpleString("OK")

	b.StartTimer()
	for b.Loop() {
		_ = s.copyMarshaler()
	}
	b.StopTimer()
}
