package textenc

import (
	"time"
)

// PutTime formats the input time as YYYY-MM-DDTHH:MM:SS.MMM
// and appends the encoded string to the input byte slice.
func PutTime(dst []byte, t time.Time) []byte {
	year, month, day := t.Date()
	hour, min, sec := t.Clock()
	ms := t.Nanosecond() / 1e6

	// Ensure capacity for the fixed 23-byte layout, then write directly.
	n := len(dst)
	if n+23 <= cap(dst) {
		dst = dst[:n+23]
	} else {
		dst = append(dst, make([]byte, 23)...)
	}

	// YYYY-MM-DDTHH:MM:SS.MMM
	// 01234567890123456789012

	// Year: 4 digits (arithmetic).
	put4(dst[n:], year)
	dst[n+4] = '-'
	// Month: 2 digits
	put2(dst[n+5:], int(month))
	dst[n+7] = '-'
	// Day: 2 digits
	put2(dst[n+8:], day)
	dst[n+10] = 'T'
	// Hour: 2 digits
	put2(dst[n+11:], hour)
	dst[n+13] = ':'
	// Minute: 2 digits
	put2(dst[n+14:], min)
	dst[n+16] = ':'
	// Second: 2 digits
	put2(dst[n+17:], sec)
	dst[n+19] = '.'
	// Millisecond: 3 digits
	put3(dst[n+20:], ms)

	return dst
}

func put2(dst []byte, v int) {
	dst[0] = '0' + byte(v/10)
	dst[1] = '0' + byte(v%10)
}

func put3(dst []byte, v int) {
	dst[0] = '0' + byte(v/100)
	dst[1] = '0' + byte((v/10)%10)
	dst[2] = '0' + byte(v%10)
}

func put4(dst []byte, v int) {
	dst[0] = '0' + byte(v/1000)
	dst[1] = '0' + byte((v/100)%10)
	dst[2] = '0' + byte((v/10)%10)
	dst[3] = '0' + byte(v%10)
}

// PutDuration formats the input duration with the given unit & format
// and appends the encoded string to the input byte slice.
func PutDuration(dst []byte, d time.Duration) []byte {
	return PutString(dst, d.String())
}
