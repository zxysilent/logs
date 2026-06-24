package textenc

import (
	"fmt"
	"unicode/utf8"
)

const hex = "0123456789abcdef"

var noEscapeTable = [256]bool{}

func init() {
	for i := 0; i <= 0x7e; i++ {
		noEscapeTable[i] = i >= 0x20 && i != '\\' && i != '"'
	}
}

// PutString encodes the input string and appends
// the encoded string to the input byte slice without quoting spaces/tabs.
func PutString(dst []byte, s string) []byte {
	return quoteString(dst, s, false)
}

// PutStringQuote encodes the input string and appends
// the encoded string to the input byte slice, quoting if it contains spaces/tabs.
func PutStringQuote(dst []byte, s string) []byte {
	return quoteString(dst, s, true)
}

func quoteString(dst []byte, s string, quote bool) []byte {
	// Single pass: find the first byte that needs escaping while tracking
	// whether the string contains a space/tab (which forces quoting).
	// Most keys/values are clean ASCII, so this returns via the fast path.
	needQuote := false
	for i := 0; i < len(s); i++ {
		b := s[i]
		if noEscapeTable[b] {
			// space (0x20) is in noEscapeTable but still forces quoting.
			if quote && b == ' ' {
				needQuote = true
			}
			continue
		}
		// b needs escaping (or is a tab). Re-scan the rest only when we still
		// don't know whether quoting is required, to keep output identical.
		if quote && !needQuote {
			for j := i; j < len(s); j++ {
				if c := s[j]; c == ' ' || c == '\t' {
					needQuote = true
					break
				}
			}
		}
		if needQuote {
			dst = append(dst, '"')
		}
		dst = appendStringComplex(dst, s, i)
		if needQuote {
			dst = append(dst, '"')
		}
		return dst
	}
	// No escape characters — fast path: just copy the string, with quotes if needed.
	if needQuote {
		dst = append(dst, '"')
	}
	dst = append(dst, s...)
	if needQuote {
		dst = append(dst, '"')
	}
	return dst
}

// PutStringer encodes the input Stringer to json and appends the
// encoded Stringer value to the input byte slice.
func PutStringer(dst []byte, val fmt.Stringer) []byte {
	if val == nil {
		return PutNil(dst)
	}
	return PutStringQuote(dst, val.String())
}

// appendStringComplex takes over from quoteString when a character that
// needs escaping is encountered, encoding the remainder byte-by-byte.
func appendStringComplex(dst []byte, s string, i int) []byte {
	start := 0
	for i < len(s) {
		b := s[i]
		if b >= utf8.RuneSelf {
			r, size := utf8.DecodeRuneInString(s[i:])
			if r == utf8.RuneError && size == 1 {
				// In case of error, first append previous simple characters to
				// the byte slice if any and append a replacement character code
				// in place of the invalid sequence.
				if start < i {
					dst = append(dst, s[start:i]...)
				}
				dst = append(dst, `\ufffd`...)
				i += size
				start = i
				continue
			}
			i += size
			continue
		}
		if noEscapeTable[b] {
			i++
			continue
		}
		// We encountered a character that needs to be encoded.
		// Let's append the previous simple characters to the byte slice
		// and switch our operation to read and encode the remainder
		// characters byte-by-byte.
		if start < i {
			dst = append(dst, s[start:i]...)
		}
		switch b {
		case '"', '\\':
			dst = append(dst, '\\', b)
		case '\b':
			dst = append(dst, '\\', 'b')
		case '\f':
			dst = append(dst, '\\', 'f')
		case '\n':
			dst = append(dst, '\\', 'n')
		case '\r':
			dst = append(dst, '\\', 'r')
		case '\t':
			dst = append(dst, '\\', 't')
		default:
			dst = append(dst, '\\', 'u', '0', '0', hex[b>>4], hex[b&0xF])
		}
		i++
		start = i
	}
	if start < len(s) {
		dst = append(dst, s[start:]...)
	}
	return dst
}
