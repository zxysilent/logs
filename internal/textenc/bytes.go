package textenc

import (
	"unicode/utf8"
)

// PutBytes encodes []byte without quoting spaces/tabs.
func PutBytes(dst, s []byte) []byte {
	return putBytes(dst, s, false)
}

// PutBytesQuote encodes []byte and quotes if it contains spaces/tabs.
func PutBytesQuote(dst, s []byte) []byte {
	return putBytes(dst, s, true)
}

func putBytes(dst, s []byte, quote bool) []byte {
	// Single pass: find the first byte that needs escaping while tracking
	// whether the slice contains a space/tab (which forces quoting).
	needQuote := false
	for i := 0; i < len(s); i++ {
		b := s[i]
		if noEscapeTable[b] {
			if quote && b == ' ' {
				needQuote = true
			}
			continue
		}
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
		dst = appendBytesComplex(dst, s, i)
		if needQuote {
			dst = append(dst, '"')
		}
		return dst
	}
	if needQuote {
		dst = append(dst, '"')
	}
	dst = append(dst, s...)
	if needQuote {
		dst = append(dst, '"')
	}
	return dst
}

// appendBytesComplex is a mirror of the appendStringComplex
// with []byte arg
func appendBytesComplex(dst, s []byte, i int) []byte {
	start := 0
	for i < len(s) {
		b := s[i]
		if b >= utf8.RuneSelf {
			r, size := utf8.DecodeRune(s[i:])
			if r == utf8.RuneError && size == 1 {
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
