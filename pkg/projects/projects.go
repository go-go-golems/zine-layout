package projects

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
)

// SavedImage describes an image persisted on disk for a project.
type SavedImage struct {
	Filename string
	Bytes    int64
	Width    int
	Height   int
}

// ProjectDir returns the absolute path for a project root directory.
func ProjectDir(projectsRoot, id string) string {
	return filepath.Join(projectsRoot, id)
}

// ProjectImagesDir returns the directory where project images are stored.
func ProjectImagesDir(projectsRoot, id string) string {
	return filepath.Join(ProjectDir(projectsRoot, id), "images")
}

// EnsureProjectDirs makes sure project and image directories exist.
func EnsureProjectDirs(projectsRoot, id string) error {
	if err := os.MkdirAll(ProjectDir(projectsRoot, id), 0o755); err != nil {
		return err
	}
	return os.MkdirAll(ProjectImagesDir(projectsRoot, id), 0o755)
}

// SavePNGImage persists a PNG upload to the project images directory and
// returns basic metadata used for asset records.
func SavePNGImage(projectsRoot, projectID string, fh *multipart.FileHeader) (*SavedImage, error) {
	if fh == nil {
		return nil, fmt.Errorf("file header is nil")
	}
	if fh.Size == 0 {
		return nil, fmt.Errorf("empty file upload")
	}
	lower := strings.ToLower(fh.Filename)
	if !strings.HasSuffix(lower, ".png") {
		return nil, fmt.Errorf("only .png uploads are supported (got %s)", fh.Filename)
	}
	if err := EnsureProjectDirs(projectsRoot, projectID); err != nil {
		return nil, err
	}
	imgDir := ProjectImagesDir(projectsRoot, projectID)
	next := nextImageNumber(imgDir)
	filename := fmt.Sprintf("%04d.png", next)
	dstPath := filepath.Join(imgDir, filename)

	src, err := fh.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	dst, err := os.Create(dstPath)
	if err != nil {
		return nil, err
	}
	bytesWritten, copyErr := io.Copy(dst, src)
	closeErr := dst.Close()
	if copyErr != nil {
		_ = os.Remove(dstPath)
		return nil, copyErr
	}
	if closeErr != nil {
		return nil, closeErr
	}

	width, height, err := readImageSize(dstPath)
	if err != nil {
		_ = os.Remove(dstPath)
		return nil, err
	}

	return &SavedImage{
		Filename: filename,
		Bytes:    bytesWritten,
		Width:    width,
		Height:   height,
	}, nil
}

func nextImageNumber(dir string) int {
	max := 0
	entries, _ := os.ReadDir(dir)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".png") {
			base := strings.TrimSuffix(name, ".png")
			var num int
			if _, err := fmt.Sscanf(base, "%d", &num); err == nil && num > max {
				max = num
			}
		}
	}
	return max + 1
}

func readImageSize(path string) (int, int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, 0, err
	}
	defer f.Close()
	cfg, _, err := image.DecodeConfig(f)
	if err != nil {
		return 0, 0, err
	}
	return cfg.Width, cfg.Height, nil
}
