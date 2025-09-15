package doc

import "embed"

// DocsFS embeds all markdown topics for the help system.
//
//go:embed topics/*
var DocsFS embed.FS
