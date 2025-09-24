package gallery

import (
	"html/template"
	"os"
	"path/filepath"
)

type Entry struct {
	Index    int
	Spread   string
	Panel    string
	RelPath  string
	FileName string
	Trace    []string
}

type PageData struct {
	Title   string
	Entries []Entry
}

const tmpl = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8" />
<title>{{ .Title }}</title>
<style>
body { font-family: sans-serif; background: #f7f7f7; color: #222; }
main { display: grid; grid-template-columns: repeat(auto-fill, minmax(320px, 1fr)); gap: 24px; padding: 24px; }
figure { background: #fff; border: 1px solid #ccc; border-radius: 8px; padding: 12px; box-shadow: 0 2px 6px rgba(0,0,0,0.08); }
figure img { width: 100%; height: auto; border: 4px solid #e3e3e3; box-sizing: border-box; background: #fff; }
figcaption { margin-top: 8px; font-size: 0.9rem; line-height: 1.3; }
figcaption strong { display: block; }
details { margin-top: 8px; font-size: 0.8rem; }
details summary { cursor: pointer; font-weight: 600; }
details ul { margin: 8px 0 0 16px; padding: 0; }
</style>
</head>
<body>
<header>
<h1>{{ .Title }}</h1>
<p>Total spreads: {{ len .Entries }}</p>
</header>
<main>
{{ range .Entries }}
<figure>
<img src="{{ .RelPath }}" alt="{{ .FileName }}" />
<figcaption>
<strong>{{ .Spread }} — {{ .Panel }}</strong>
<span>#{{ .Index }}</span>
<small>{{ .FileName }}</small>
</figcaption>
{{ if .Trace }}
<details>
<summary>Trace</summary>
<ul>
{{ range .Trace }}
<li>{{ . }}</li>
{{ end }}
</ul>
</details>
{{ end }}
</figure>
{{ end }}
</main>
</body>
</html>`

// Write saves a gallery page to disk.
func Write(path, title string, entries []Entry) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	tpl, err := template.New("gallery").Parse(tmpl)
	if err != nil {
		return err
	}
	data := PageData{Title: title, Entries: entries}
	return tpl.Execute(f, data)
}
