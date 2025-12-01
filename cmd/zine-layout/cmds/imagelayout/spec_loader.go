package imagelayoutcmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-go-golems/zine-layout/pkg/imagelayout"
	"gopkg.in/yaml.v3"
)

type layoutSpecDocument struct {
	Layout imagelayout.LayoutRequest `json:"layout" yaml:"layout"`
	Image  struct {
		Width  int `json:"width" yaml:"width"`
		Height int `json:"height" yaml:"height"`
	} `json:"image" yaml:"image"`
}

func loadLayoutSpec(path string) (imagelayout.LayoutRequest, imagelayout.ImageMeta, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return imagelayout.LayoutRequest{}, imagelayout.ImageMeta{}, fmt.Errorf("read spec: %w", err)
	}
	var doc layoutSpecDocument
	if err := unmarshalByExt(path, data, &doc); err != nil {
		return imagelayout.LayoutRequest{}, imagelayout.ImageMeta{}, fmt.Errorf("parse spec: %w", err)
	}
	layout := doc.Layout
	meta := imagelayout.ImageMeta{Width: doc.Image.Width, Height: doc.Image.Height}
	return layout, meta, nil
}

func unmarshalByExt(path string, data []byte, out any) error {
	ext := filepath.Ext(path)
	switch ext {
	case ".yaml", ".yml":
		return yaml.Unmarshal(data, out)
	case ".json":
		return json.Unmarshal(data, out)
	default:
		return fmt.Errorf("unsupported spec extension: %s", ext)
	}
}
