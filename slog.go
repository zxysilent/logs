package logs

import (
	"context"
	"io"
	"log/slog"

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

// SlogHandler adapts the logs library as a slog.Handler, writing key=value
// logfmt output through the given io.Writer.
//
// Usage:
//
//	h := logs.NewSlogHandler(os.Stderr, &logs.SlogHandlerOptions{Level: slog.LevelInfo})
//	logger := slog.New(h)
//	logger.Info("hello", "key", "value")
type SlogHandler struct {
	w     io.Writer
	level slog.Leveler
	group string      // current group prefix (dot-separated), from WithGroup
	attrs []slog.Attr // unencoded attrs from WithAttrs
	buf   []byte      // reusable encode buffer
}

// SlogHandlerOptions configures a SlogHandler.
type SlogHandlerOptions struct {
	// Level is the minimum level to log. Defaults to slog.LevelInfo.
	Level slog.Leveler
}

// NewSlogHandler creates a SlogHandler that writes logfmt output to w.
func NewSlogHandler(w io.Writer, opts *SlogHandlerOptions) *SlogHandler {
	h := &SlogHandler{w: w, level: slog.LevelInfo}
	if opts != nil && opts.Level != nil {
		h.level = opts.Level
	}
	return h
}

// SlogHandler returns a slog.Handler that writes through this Logger's config.
func (l *Logger) SlogHandler() slog.Handler {
	switch l.cfg.level {
	case LevelDebug, LevelInfo, LevelWarn, LevelError:
		return &SlogHandler{w: l.cfg.out, level: logSlogLeveler(l.cfg.level)}
	default:
		return &internalSlogHandler{cfg: l.cfg, level: l.cfg.level}
	}
}

func logSlogLeveler(lv Level) slog.Level {
	switch lv {
	case LevelDebug:
		return slog.LevelDebug
	case LevelInfo:
		return slog.LevelInfo
	case LevelWarn:
		return slog.LevelWarn
	default:
		return slog.LevelError
	}
}

// Enabled reports whether the handler handles records at the given level.
func (h *SlogHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level.Level()
}

// Handle formats the slog.Record as logfmt and writes it to the output.
func (h *SlogHandler) Handle(_ context.Context, r slog.Record) error {
	buf := h.buf[:0]

	buf = append(buf, "time="...)
	buf = textenc.PutTime(buf, r.Time)

	buf = append(buf, " level="...)
	buf = append(buf, slogLevelString(r.Level)...)

	// pre-attrs (from WithAttrs)
	for _, a := range h.attrs {
		buf = h.appendAttr(buf, a, h.group)
	}

	// record attrs (from the log call)
	r.Attrs(func(a slog.Attr) bool {
		buf = h.appendAttr(buf, a, h.group)
		return true
	})

	buf = append(buf, " msg="...)
	buf = textenc.PutStringQuote(buf, r.Message)
	buf = append(buf, '\n')

	_, err := h.w.Write(buf)
	h.buf = buf[:0]
	return err
}

// WithAttrs returns a new Handler with the given attrs stored.
func (h *SlogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}
	clone := *h
	clone.attrs = make([]slog.Attr, 0, len(h.attrs)+len(attrs))
	clone.attrs = append(clone.attrs, h.attrs...)
	clone.attrs = append(clone.attrs, attrs...)
	return &clone
}

// WithGroup returns a new Handler with the given group appended.
func (h *SlogHandler) WithGroup(name string) slog.Handler {
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

// appendAttr encodes a single slog.Attr in logfmt key=value format.
// PutKeyRaw already adds a leading space separator.
func (h *SlogHandler) appendAttr(dst []byte, a slog.Attr, group string) []byte {
	key := a.Key
	if group != "" {
		key = group + "." + key
	}
	if a.Value.Kind() == slog.KindGroup {
		for _, sub := range a.Value.Group() {
			dst = h.appendAttr(dst, sub, key)
		}
		return dst
	}
	dst = textenc.PutKeyRaw(dst, key)
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

// ---------------------------------------------------------------------------
// internal fallback handler for non-standard levels (LevelMute / custom)
// ---------------------------------------------------------------------------

type internalSlogHandler struct {
	cfg   *config
	level Level
}

func (h *internalSlogHandler) Enabled(_ context.Context, level slog.Level) bool {
	return mapSlogLevel(level) >= h.level
}

func (h *internalSlogHandler) Handle(_ context.Context, r slog.Record) error {
	if !h.Enabled(nil, r.Level) {
		return nil
	}
	buf := getb()
	defer putb(buf)

	*buf = append(*buf, "time="...)
	*buf = textenc.PutTime(*buf, r.Time)
	putLevel(buf, mapSlogLevel(r.Level))

	r.Attrs(func(a slog.Attr) bool {
		*buf = append(*buf, ' ')
		*buf = appendSlogAttrKV(*buf, a, "")
		return true
	})

	*buf = append(*buf, " msg="...)
	*buf = textenc.PutStringQuote(*buf, r.Message)
	*buf = append(*buf, '\n')
	h.cfg.out.Write(*buf)
	return nil
}

func (h *internalSlogHandler) WithAttrs(attrs []slog.Attr) slog.Handler { return h }
func (h *internalSlogHandler) WithGroup(name string) slog.Handler       { return h }

func mapSlogLevel(lv slog.Level) Level {
	switch {
	case lv <= slog.LevelDebug:
		return LevelDebug
	case lv <= slog.LevelInfo:
		return LevelInfo
	case lv <= slog.LevelWarn:
		return LevelWarn
	default:
		return LevelError
	}
}

// appendSlogAttrKV is appendSlogValue + key without leading space.
func appendSlogAttrKV(dst []byte, a slog.Attr, group string) []byte {
	key := a.Key
	if group != "" {
		key = group + "." + key
	}
	if a.Value.Kind() == slog.KindGroup {
		for _, sub := range a.Value.Group() {
			dst = appendSlogAttrKV(dst, sub, key)
		}
		return dst
	}
	dst = textenc.PutKeyRaw(dst, key)
	return appendSlogValue(dst, a.Value)
}
