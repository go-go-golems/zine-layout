package main

import (
    "io"
    "os"
    "time"

    "github.com/rs/zerolog"
    "github.com/rs/zerolog/log"
)

func initLogger(verbose bool) zerolog.Logger {
    // default: pretty console writer
    var w io.Writer = os.Stderr
    cw := zerolog.ConsoleWriter{Out: w, TimeFormat: time.RFC3339Nano}
    cw.PartsOrder = []string{"time", "level", "caller", "message"}
    logger := zerolog.New(cw).With().Timestamp().Caller().Logger()

    // log level
    level := zerolog.InfoLevel
    if verbose {
        level = zerolog.DebugLevel
    }
    zerolog.SetGlobalLevel(level)
    log.Logger = logger
    return logger
}


