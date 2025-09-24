package engine

import (
	"fmt"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// TraceEntry represents a structured log message captured during compute/render.
type TraceEntry struct {
	Stage   string `json:"stage"`
	Message string `json:"message"`
}

// Tracer wraps zerolog and mirrors logs into a trace slice for later display.
type Tracer struct {
	logger  zerolog.Logger
	entries *[]TraceEntry
}

// NewTracer builds a tracer for a spread/phase pair.
func NewTracer(spread, phase string, entries *[]TraceEntry) Tracer {
	logger := log.With().Str("spread", spread).Str("phase", phase).Logger()
	return Tracer{logger: logger, entries: entries}
}

// Log writes the formatted message into the trace buffer and through zerolog.
func (t Tracer) Log(stage, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if t.entries != nil {
		*t.entries = append(*t.entries, TraceEntry{Stage: stage, Message: msg})
	}
	evt := t.logger.Info().Str("stage", stage)
	evt.Msg(msg)
}
