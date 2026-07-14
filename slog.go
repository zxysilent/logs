package logs

import (
	"context"
	"log/slog"
	"runtime"
	"strconv"

	"github.com/zxysilent/logs/internal/textenc"
)

// slogLevelString maps slog.Level to the short logfmt level string.
func slogLevelString(lv slog.Level) string {
	switch {
	case lv <= slog.LevelDebug:
		return "DBG"
	case lv <= slog.LevelInfo:
		return "INF"
	case lv <= slog.LevelWarn:
		return "WRN"
	default:
		return "ERR"
	}
}

// slogHandler adapts the logs library as a slog.Handler, writing key=value
// logfmt output through the configured writer.
//
// Usage:
//
//	logger := slog.New(logs.NewSlogHandler())
//	logger.Info("hello", "key", "value")
type slogHandler struct {
	cfg   *config
	group string
	attrs []byte
}

// NewSlogHandler returns a handler backed by the package-level logs config.
func NewSlogHandler() slog.Handler {
	return l.NewSlogHandler()
}

// NewSlogHandler returns a slog.Handler that writes through this Logger's config,
// inheriting level, caller, and separator settings.
func (l *Logger) NewSlogHandler() slog.Handler {
	return &slogHandler{cfg: l.cfg}
}

// Enabled reports whether the handler handles records at the given level.
func (h *slogHandler) Enabled(_ context.Context, level slog.Level) bool {
	if h.cfg.level == LevelMute {
		return false
	}
	return Level(level) >= h.cfg.level
}

// Handle formats the slog.Record as logfmt and writes it to the output.
func (h *slogHandler) Handle(_ context.Context, r slog.Record) error {
	buf := getb()
	defer putb(buf)

	*buf = append(*buf, "time="...)
	*buf = textenc.PutTime(*buf, r.Time)

	*buf = append(*buf, " level="...)
	*buf = append(*buf, slogLevelString(r.Level)...)

	if h.cfg.caller && r.PC != 0 {
		*buf = putSlogCaller(*buf, r.PC, h.cfg.sep)
	}

	if len(h.attrs) > 0 {
		*buf = append(*buf, ' ')
		*buf = append(*buf, h.attrs...)
	}

	r.Attrs(func(a slog.Attr) bool {
		*buf = appendSlogAttr(*buf, a, h.group)
		return true
	})

	*buf = append(*buf, " msg="...)
	*buf = textenc.PutStringQuote(*buf, r.Message)
	*buf = append(*buf, '\n')

	_, err := h.cfg.out.Write(*buf)
	return err
}

func putSlogCaller(dst []byte, pc uintptr, sep []string) []byte {
	file := "###"
	line := 0
	pc-- // slog.Record.PC is the return PC after the call instruction.
	if fn := runtime.FuncForPC(pc); fn != nil {
		file, line = fn.FileLine(pc)
		if slash := lastSep(file, sep); slash >= 0 {
			file = file[slash:]
		}
	}
	dst = append(dst, " caller="...)
	dst = textenc.PutString(dst, file)
	dst = append(dst, ':')
	return strconv.AppendInt(dst, int64(line), 10)
}

// WithAttrs returns a new Handler with the given attrs stored.
func (h *slogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}
	clone := *h
	clone.attrs = make([]byte, len(h.attrs), len(h.attrs)+len(attrs)*16)
	clone.attrs = append(clone.attrs, h.attrs...)
	for _, attr := range attrs {
		clone.attrs = appendSlogAttr(clone.attrs, attr, h.group)
	}
	return &clone
}

// WithGroup returns a new Handler with the given group appended.
func (h *slogHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	clone := *h
	if clone.group == "" {
		clone.group = name
	} else {
		clone.group = clone.group + "." + name
	}
	return &clone
}

// appendSlogAttr encodes a single slog.Attr in logfmt key=value format.
func appendSlogAttr(dst []byte, a slog.Attr, group string) []byte {
	if a.Equal(slog.Attr{}) {
		return dst
	}
	a.Value = a.Value.Resolve()
	key := a.Key
	if group != "" {
		key = group + "." + key
	}
	if a.Value.Kind() == slog.KindGroup {
		for _, sub := range a.Value.Group() {
			dst = appendSlogAttr(dst, sub, key)
		}
		return dst
	}
	dst = textenc.PutKey(dst, key)
	return appendSlogValue(dst, a.Value)
}

// appendSlogValue encodes a slog.Value by its Kind.
func appendSlogValue(dst []byte, v slog.Value) []byte {
	switch v.Kind() {
	case slog.KindString:
		return textenc.PutStringQuote(dst, v.String())
	case slog.KindInt64:
		return textenc.PutInt64(dst, v.Int64())
	case slog.KindUint64:
		return textenc.PutUint64(dst, v.Uint64())
	case slog.KindFloat64:
		return textenc.PutFloat64(dst, v.Float64())
	case slog.KindBool:
		return textenc.PutBool(dst, v.Bool())
	case slog.KindDuration:
		return textenc.PutDuration(dst, v.Duration())
	case slog.KindTime:
		return textenc.PutTime(dst, v.Time())
	case slog.KindAny:
		return textenc.PutAny(dst, v.Any())
	default:
		return textenc.PutNil(dst)
	}
}
