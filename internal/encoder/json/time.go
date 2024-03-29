package json

import (
	"time"
)

// PutTime formats the input time with the given format
// and appends the encoded string to the input byte slice.
func (e Encoder) PutTime(dst []byte, t time.Time) []byte {
	dst = append(dst, '"')
	// Format date.
	year, month, day := t.Date()
	dst = appendInt(dst, year, 4)
	dst = append(dst, '-')
	dst = appendInt(dst, int(month), 2)
	dst = append(dst, '-')
	dst = appendInt(dst, day, 2)

	dst = append(dst, 'T')

	// Format time.
	hour, min, sec := t.Clock()
	dst = appendInt(dst, hour, 2)
	dst = append(dst, ':')
	dst = appendInt(dst, min, 2)
	dst = append(dst, ':')
	dst = appendInt(dst, sec, 2)

	dst = append(dst, '.')
	ms := t.Nanosecond() / 1e6
	dst = appendInt(dst, ms, 3)
	dst = append(dst, '"')
	return dst
}

// PutDuration formats the input duration with the given unit & format
// and appends the encoded string to the input byte slice.
func (e Encoder) PutDuration(dst []byte, d time.Duration) []byte {
	return e.PutString(dst, d.String())
}

// appendInt appends the decimal form of x to b and returns the result.
// If the decimal form (excluding sign) is shorter than width, the result is padded with leading 0's.
// Duplicates functionality in strconv, but avoids dependency.
func appendInt(b []byte, x int, width int) []byte {
	u := uint(x)
	if x < 0 {
		b = append(b, '-')
		u = uint(-x)
	}

	// 2-digit and 4-digit fields are the most common in time formats.
	utod := func(u uint) byte { return '0' + byte(u) }
	switch {
	case width == 2 && u < 1e2:
		return append(b, utod(u/1e1), utod(u%1e1))
	case width == 4 && u < 1e4:
		return append(b, utod(u/1e3), utod(u/1e2%1e1), utod(u/1e1%1e1), utod(u%1e1))
	}

	// Compute the number of decimal digits.
	var n int
	if u == 0 {
		n = 1
	}
	for u2 := u; u2 > 0; u2 /= 10 {
		n++
	}

	// Add 0-padding.
	for pad := width - n; pad > 0; pad-- {
		b = append(b, '0')
	}

	// Ensure capacity.
	if len(b)+n <= cap(b) {
		b = b[:len(b)+n]
	} else {
		b = append(b, make([]byte, n)...)
	}

	// Assemble decimal in reverse order.
	i := len(b) - 1
	for u >= 10 && i > 0 {
		q := u / 10
		b[i] = utod(u - q*10)
		u = q
		i--
	}
	b[i] = utod(u)
	return b
}
