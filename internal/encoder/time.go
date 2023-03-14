package encoder

import (
	"strconv"
	"time"
)

const (
	timeFormatUnix      = ""
	timeFormatUnixMs    = "UNIXMS"
	timeFormatUnixMicro = "UNIXMICRO"
	timeFormatUnixNano  = "UNIXNANO"
	timeFieldFormat     = "2006/01/02 15:04:05.000"
)

// PutTime formats the input time with the given format
// and appends the encoded string to the input byte slice.
func (e Encoder) PutTime(dst []byte, t time.Time, format string) []byte {
	switch format {
	case timeFormatUnix:
		return e.PutInt64(dst, t.Unix())
	case timeFormatUnixMs:
		return e.PutInt64(dst, t.UnixNano()/1000000)
	case timeFormatUnixMicro:
		return e.PutInt64(dst, t.UnixNano()/1000)
	case timeFormatUnixNano:
		return e.PutInt64(dst, t.UnixNano())
	case timeFieldFormat:
		return e.PutTimeFast(dst, t)
	}
	return append(t.AppendFormat(append(dst, '"'), format), '"')
}

// PutTime formats the input time with the given format
// and appends the encoded string to the input byte slice.
func (e Encoder) PutTimeFast(dst []byte, t time.Time) []byte {
	return appendFormatFast(dst, t)
}

// PutTimes converts the input times with the given format
// and appends the encoded string list to the input byte slice.
func (Encoder) PutTimes(dst []byte, vals []time.Time, format string) []byte {
	switch format {
	case timeFormatUnix:
		return appendUnixTimes(dst, vals)
	case timeFormatUnixMs:
		return appendUnixNanoTimes(dst, vals, 1000000)
	case timeFormatUnixMicro:
		return appendUnixNanoTimes(dst, vals, 1000)
	case timeFormatUnixNano:
		return appendUnixNanoTimes(dst, vals, 1)
	}
	if len(vals) == 0 {
		return append(dst, '[', ']')
	}
	dst = append(dst, '[')
	dst = append(vals[0].AppendFormat(append(dst, '"'), format), '"')
	if len(vals) > 1 {
		for _, t := range vals[1:] {
			dst = append(t.AppendFormat(append(dst, ',', '"'), format), '"')
		}
	}
	dst = append(dst, ']')
	return dst
}

func appendUnixTimes(dst []byte, vals []time.Time) []byte {
	if len(vals) == 0 {
		return append(dst, '[', ']')
	}
	dst = append(dst, '[')
	dst = strconv.AppendInt(dst, vals[0].Unix(), 10)
	if len(vals) > 1 {
		for _, t := range vals[1:] {
			dst = strconv.AppendInt(append(dst, ','), t.Unix(), 10)
		}
	}
	dst = append(dst, ']')
	return dst
}

func appendUnixNanoTimes(dst []byte, vals []time.Time, div int64) []byte {
	if len(vals) == 0 {
		return append(dst, '[', ']')
	}
	dst = append(dst, '[')
	dst = strconv.AppendInt(dst, vals[0].UnixNano()/div, 10)
	if len(vals) > 1 {
		for _, t := range vals[1:] {
			dst = strconv.AppendInt(append(dst, ','), t.UnixNano()/div, 10)
		}
	}
	dst = append(dst, ']')
	return dst
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
func appendFormatFast(b []byte, t time.Time) []byte {
	b = append(b, '"')
	// Format date.
	year, month, day := t.Date()
	b = appendInt(b, year, 4)
	b = append(b, '-')
	b = appendInt(b, int(month), 2)
	b = append(b, '-')
	b = appendInt(b, day, 2)

	b = append(b, ' ')

	// Format time.
	hour, min, sec := t.Clock()
	b = appendInt(b, hour, 2)
	b = append(b, ':')
	b = appendInt(b, min, 2)
	b = append(b, ':')
	b = appendInt(b, sec, 2)

	b = append(b, '.')
	ms := t.Nanosecond() / 1e6
	b = appendInt(b, ms, 3)
	b = append(b, '"')
	return b
}

// PutDuration formats the input duration with the given unit & format
// and appends the encoded string to the input byte slice.
func (e Encoder) PutDuration(dst []byte, d time.Duration, unit time.Duration) []byte {
	return e.PutFloat64(dst, float64(d)/float64(unit))
}

// PutDurations formats the input durations with the given unit & format
// and appends the encoded string list to the input byte slice.
func (e Encoder) PutDurations(dst []byte, vals []time.Duration, unit time.Duration) []byte {
	if len(vals) == 0 {
		return append(dst, '[', ']')
	}
	dst = append(dst, '[')
	dst = e.PutDuration(dst, vals[0], unit)
	if len(vals) > 1 {
		for _, d := range vals[1:] {
			dst = e.PutDuration(append(dst, ','), d, unit)
		}
	}
	dst = append(dst, ']')
	return dst
}
