package media

import (
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"

	"github.com/go-go-golems/zine-layout/cmd/experiments/spread-algorithm-actually-from-sonnet/internal/engine"
)

// Metadata reads dimensions of an image without decoding the full pixel data.
func Metadata(path string) (engine.SourceMeta, error) {
	f, err := os.Open(path)
	if err != nil {
		return engine.SourceMeta{}, err
	}
	defer f.Close()
	cfg, _, err := image.DecodeConfig(f)
	if err != nil {
		return engine.SourceMeta{}, err
	}
	return engine.SourceMeta{Width: cfg.Width, Height: cfg.Height}, nil
}

// Load decodes an image from disk.
func Load(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	return img, err
}
