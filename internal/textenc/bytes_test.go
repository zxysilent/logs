package textenc

import (
	"testing"
)

func TestPutBytes(t *testing.T) {
	for _, tt := range encodeStringTests {
		b := PutBytes([]byte{}, []byte(tt.in))
		if got, want := string(b), tt.out; got != want {
			t.Errorf("appendBytes(%q) = %#q, want %#q", tt.in, got, want)
		}
	}
}

// func TestStringBytes(t *testing.T) {
// 	t.Parallel()
// 	// Test that encodeState.stringBytes and encodeState.string use the same encoding.
// 	var r []rune
// 	for i := '\u0000'; i <= unicode.MaxRune; i++ {
// 		r = append(r, i)
// 	}
// 	s := string(r) + "\xff\xff\xffhello" // some invalid UTF-8 too

// 	encStr := string(PutString([]byte{}, s))
// 	encBytes := string(PutBytes([]byte{}, []byte(s)))

// 	if encStr != encBytes {
// 		i := 0
// 		for i < len(encStr) && i < len(encBytes) && encStr[i] == encBytes[i] {
// 			i++
// 		}
// 		encStr = encStr[i:]
// 		encBytes = encBytes[i:]
// 		i = 0
// 		for i < len(encStr) && i < len(encBytes) && encStr[len(encStr)-i-1] == encBytes[len(encBytes)-i-1] {
// 			i++
// 		}
// 		encStr = encStr[:len(encStr)-i]
// 		encBytes = encBytes[:len(encBytes)-i]

// 		if len(encStr) > 20 {
// 			encStr = encStr[:20] + "..."
// 		}
// 		if len(encBytes) > 20 {
// 			encBytes = encBytes[:20] + "..."
// 		}

// 		t.Errorf("encodings differ at %#q vs %#q", encStr, encBytes)
// 	}
// }

func TestPutBytesQuote(t *testing.T) {
	tests := []struct {
		in  string
		out string
	}{
		{"hello world", `"hello world"`},
		{"a b", `"a b"`},
		{"\thello", `"\thello"`},
		{"hello\tworld", `"hello\tworld"`},
		{"a b\tc", `"a b\tc"`},
		{`a"b c`, `"a\"b c"`}, // space after escape — needQuote must not be missed
		{"hello", `hello`},
		{"hello\nworld", `hello\nworld`},
	}
	for _, tt := range tests {
		got := string(PutBytesQuote([]byte{}, []byte(tt.in)))
		if got != tt.out {
			t.Errorf("PutBytesQuote(%q) = %#q, want %#q", tt.in, got, tt.out)
		}
		//Verify PutStringQuote matching
		strGot := string(PutStringQuote([]byte{}, tt.in))
		if got != strGot {
			t.Errorf("PutBytesQuote(%q)=%#q != PutStringQuote(%q)=%#q", tt.in, got, tt.in, strGot)
		}
	}
}

func BenchmarkPutBytes(b *testing.B) {
	tests := map[string]string{
		"NoEncoding":       `aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa`,
		"EncodingFirst":    `"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa`,
		"EncodingMiddle":   `aaaaaaaaaaaaaaaaaaaaaaaaa"aaaaaaaaaaaaaaaaaaaaaaaa`,
		"EncodingLast":     `aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"`,
		"MultiBytesFirst":  `❤️aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa`,
		"MultiBytesMiddle": `aaaaaaaaaaaaaaaaaaaaaaaaa❤️aaaaaaaaaaaaaaaaaaaaaaaa`,
		"MultiBytesLast":   `aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa❤️`,
	}
	for name, str := range tests {
		byt := []byte(str)
		b.Run(name, func(b *testing.B) {
			buf := make([]byte, 0, 100)
			for i := 0; i < b.N; i++ {
				_ = PutBytes(buf, byt)
			}
		})
	}
}
