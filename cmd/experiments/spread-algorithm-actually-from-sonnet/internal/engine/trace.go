package engine

import "fmt"

// TraceEntry represents a structured log message captured during compute/render.
type TraceEntry struct {
	Stage   string `json:"stage"`
	Message string `json:"message"`
}

func addTrace(trace *[]TraceEntry, stage, format string, args ...interface{}) {
	entry := TraceEntry{Stage: stage, Message: fmt.Sprintf(format, args...)}
	*trace = append(*trace, entry)
}
