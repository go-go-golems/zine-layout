package namer

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type TemplateData struct {
	Index         int
	Name          string
	Panel         string
	ImageBasename string
	Extension     string
}

var tokenPattern = regexp.MustCompile(`\{([a-z_]+)(:[^}]*)?\}`)

// Build generates a filename from the template and data.
func Build(template string, data TemplateData) (string, error) {
	if template == "" {
		template = "{index}-{name}-{panel}.{ext}"
	}
	var err error
	result := tokenPattern.ReplaceAllStringFunc(template, func(m string) string {
		matches := tokenPattern.FindStringSubmatch(m)
		if matches == nil {
			return m
		}
		key := matches[1]
		spec := ""
		if len(matches) > 2 {
			spec = matches[2]
		}
		switch key {
		case "index":
			return formatIndex(data.Index, spec)
		case "name":
			return sanitize(data.Name)
		case "panel":
			return sanitize(data.Panel)
		case "image_basename":
			return sanitize(data.ImageBasename)
		case "ext":
			return strings.TrimPrefix(data.Extension, ".")
		default:
			err = fmt.Errorf("unknown filename token %q", key)
			return m
		}
	})
	if err != nil {
		return "", err
	}
	return filepath.Clean(result), nil
}

func formatIndex(index int, spec string) string {
	if spec == "" {
		return strconv.Itoa(index)
	}
	// spec like :03d
	format := "%" + strings.TrimPrefix(spec, ":")
	return fmt.Sprintf(format, index)
}

func sanitize(s string) string {
	s = strings.TrimSpace(s)
	replacer := strings.NewReplacer(
		" ", "-",
		"/", "-",
		"\\", "-",
	)
	return replacer.Replace(s)
}
