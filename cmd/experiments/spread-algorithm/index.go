package main

import (
    "fmt"
    "html"
    "os"
    "path/filepath"
    "strings"
)

func WriteHTMLIndex(path string, results []SpreadOutput) error {
    var b strings.Builder
    b.WriteString("<!DOCTYPE html>\n<html><head><meta charset=\"utf-8\">\n")
    b.WriteString("<title>Spread Renders</title>\n")
    b.WriteString("<style>\n")
    b.WriteString("body{font-family:system-ui,-apple-system,Segoe UI,Roboto,Ubuntu,\"Helvetica Neue\",Arial,sans-serif;margin:24px;background:#f8f9fb;color:#222}\n")
    b.WriteString("h1{font-size:20px;margin:0 0 16px}\n")
    b.WriteString(".spread{background:#fff;border:1px solid #e5e7eb;border-radius:8px;margin:16px 0;padding:16px;box-shadow:0 1px 2px rgba(0,0,0,0.04)}\n")
    b.WriteString(".meta{font-size:12px;color:#666;margin-bottom:8px}\n")
    b.WriteString(".imgs{display:flex;gap:12px;flex-wrap:wrap}\n")
    b.WriteString(".imgs img{max-width:100%;height:auto;border:1px solid #999;background:#fafafa;padding:2px;border-radius:4px}\n")
    b.WriteString("details{margin-top:8px}\n")
    b.WriteString("pre{background:#0b1021;color:#e6edf3;padding:12px;border-radius:6px;overflow:auto;font-size:12px;line-height:1.4}\n")
    b.WriteString("summary{cursor:pointer;color:#0366d6}\n")
    b.WriteString("</style>\n</head><body>\n")
    b.WriteString("<h1>Spread Renders</h1>\n")

    // Images are expected to be in the same output directory passed to the CLI
    baseDir := filepath.Dir(path)

    for _, r := range results {
        b.WriteString("<div class=\"spread\">\n")
        b.WriteString(fmt.Sprintf("<div class=\"meta\"><strong>%s</strong></div>\n", html.EscapeString(r.Name)))
        b.WriteString("<div class=\"imgs\">\n")
        for i, fp := range r.PanelFiles {
            // Ensure we build a correct link relative to index.html
            rel := fp
            if !filepath.IsAbs(rel) {
                rel = filepath.Join(baseDir, rel)
            }
            if r2, err := filepath.Rel(baseDir, rel); err == nil {
                rel = r2
            }
            alt := "image"
            if i < len(r.PanelLabels) { alt = r.PanelLabels[i] }
            b.WriteString(fmt.Sprintf("<figure><img src=\"%s\" alt=\"%s\"><figcaption style=\"text-align:center;font-size:12px;color:#555\">%s</figcaption></figure>\n",
                html.EscapeString(rel), html.EscapeString(alt), html.EscapeString(alt)))
        }
        b.WriteString("</div>\n")
        if len(r.Logs) > 0 {
            b.WriteString("<details><summary>Algorithm trace</summary><pre>")
            for _, line := range r.Logs {
                b.WriteString(html.EscapeString(line))
                b.WriteString("\n")
            }
            b.WriteString("</pre></details>\n")
        }
        b.WriteString("</div>\n")
    }

    b.WriteString("</body></html>\n")
    return os.WriteFile(path, []byte(b.String()), 0o644)
}


