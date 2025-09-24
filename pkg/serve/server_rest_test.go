package serve

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-go-golems/zine-layout/pkg/projects"
	"github.com/go-go-golems/zine-layout/pkg/spread"
	simple "github.com/go-go-golems/zine-layout/pkg/spread/simple"
)

func writeTestPNG(t *testing.T, path string, width, height int) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir images: %v", err)
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create image: %v", err)
	}
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{R: 200, G: 100, B: 50, A: 255})
		}
	}
	if err := png.Encode(f, img); err != nil {
		_ = f.Close()
		t.Fatalf("encode png: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close png: %v", err)
	}
}

func setupServer(t *testing.T) (*Server, *projects.Project, func()) {
	t.Helper()
	dataRoot := t.TempDir()
	settings := Settings{
		Root:     dataRoot,
		DataRoot: dataRoot,
		Addr:     "127.0.0.1:0",
	}
	srv := New(settings)
	if err := srv.prepare(); err != nil {
		t.Fatalf("prepare server: %v", err)
	}
	project, err := projects.CreateProject(srv.projectsRoot, "rest-test")
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	imgPath := filepath.Join(projects.ProjectImagesDir(srv.projectsRoot, project.ID), "0001.png")
	writeTestPNG(t, imgPath, 16, 24)
	project.Images = []string{"0001.png"}
	project.Order = []string{"0001.png"}
	project.UpdatedAt = time.Now().UTC()
	if err := projects.WriteProject(srv.projectsRoot, project); err != nil {
		t.Fatalf("write project: %v", err)
	}
	cleanup := func() {
		if srv.db != nil {
			_ = srv.db.Close()
		}
	}
	return srv, project, cleanup
}

func TestPagesAndSpreadsLifecycle(t *testing.T) {
	srv, project, cleanup := setupServer(t)
	defer cleanup()

	ts := httptest.NewServer(srv.Routes())
	defer ts.Close()

	// Prime the project repository by listing projects (mirrors UI behaviour)
	resp, err := http.Get(ts.URL + "/api/projects")
	if err != nil {
		t.Fatalf("list projects: %v", err)
	}
	_ = resp.Body.Close()

	// Seed asset repository by hitting the images endpoint
	imagesURL := ts.URL + "/api/projects/" + project.ID + "/images"
	resp, err = http.Get(imagesURL)
	if err != nil {
		t.Fatalf("get images: %v", err)
	}
	var imagesPayload struct {
		Images []map[string]any `json:"images"`
		Order  []string         `json:"order"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&imagesPayload); err != nil {
		t.Fatalf("decode images: %v", err)
	}
	_ = resp.Body.Close()
	if len(imagesPayload.Images) != 1 {
		t.Fatalf("expected 1 image, got %d", len(imagesPayload.Images))
	}
	if len(imagesPayload.Order) != 1 || imagesPayload.Order[0] != "0001.png" {
		t.Fatalf("unexpected order: %#v", imagesPayload.Order)
	}

	// Create page 1
	pageURL := ts.URL + "/api/projects/" + project.ID + "/pages/1"
	settings := spread.DefaultSettings()
	settings.Export.OutDir = ""
	result := &simple.Result{EffectiveSpreadW: 42}
	assetID := "0001.png"
	pageBody := struct {
		AssetID  *string         `json:"asset_id,omitempty"`
		Settings spread.Settings `json:"settings"`
		Result   *simple.Result  `json:"result,omitempty"`
	}{
		AssetID:  &assetID,
		Settings: settings,
		Result:   result,
	}
	buf, _ := json.Marshal(pageBody)
	req, _ := http.NewRequest(http.MethodPut, pageURL, bytes.NewReader(buf))
	req.Header.Set("Content-Type", "application/json")
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("put page: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("put page status: %s", resp.Status)
	}
	var pageResp struct {
		Page struct {
			PageNumber int             `json:"page_number"`
			AssetID    *string         `json:"asset_id"`
			Settings   spread.Settings `json:"settings"`
			Result     *simple.Result  `json:"result"`
		} `json:"page"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&pageResp); err != nil {
		t.Fatalf("decode page: %v", err)
	}
	_ = resp.Body.Close()
	if pageResp.Page.PageNumber != 1 {
		t.Fatalf("expected page 1, got %d", pageResp.Page.PageNumber)
	}
	if pageResp.Page.AssetID == nil || *pageResp.Page.AssetID != "0001.png" {
		t.Fatalf("unexpected asset id: %#v", pageResp.Page.AssetID)
	}

	// List pages returns our record
	resp, err = http.Get(ts.URL + "/api/projects/" + project.ID + "/pages")
	if err != nil {
		t.Fatalf("list pages: %v", err)
	}
	var pagesPayload struct {
		Pages []struct {
			PageNumber int `json:"page_number"`
		} `json:"pages"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&pagesPayload); err != nil {
		t.Fatalf("decode pages list: %v", err)
	}
	_ = resp.Body.Close()
	if len(pagesPayload.Pages) != 1 || pagesPayload.Pages[0].PageNumber != 1 {
		t.Fatalf("unexpected pages payload: %#v", pagesPayload)
	}

	// Upsert spread 1 referencing page 1
	spreadURL := ts.URL + "/api/projects/" + project.ID + "/spreads/1"
	left := 1
	spreadBody := struct {
		LeftPageNumber  *int            `json:"left_page_number,omitempty"`
		RightPageNumber *int            `json:"right_page_number,omitempty"`
		Settings        spread.Settings `json:"settings"`
		Result          *simple.Result  `json:"result,omitempty"`
	}{
		LeftPageNumber: &left,
		Settings:       settings,
		Result:         result,
	}
	buf, _ = json.Marshal(spreadBody)
	req, _ = http.NewRequest(http.MethodPut, spreadURL, bytes.NewReader(buf))
	req.Header.Set("Content-Type", "application/json")
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("put spread: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("put spread status: %s", resp.Status)
	}
	var spreadResp struct {
		Spread struct {
			SpreadNumber   int  `json:"spread_number"`
			LeftPageNumber *int `json:"left_page_number"`
		} `json:"spread"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&spreadResp); err != nil {
		t.Fatalf("decode spread: %v", err)
	}
	_ = resp.Body.Close()
	if spreadResp.Spread.SpreadNumber != 1 {
		t.Fatalf("expected spread 1, got %d", spreadResp.Spread.SpreadNumber)
	}
	if spreadResp.Spread.LeftPageNumber == nil || *spreadResp.Spread.LeftPageNumber != 1 {
		t.Fatalf("unexpected left page number: %#v", spreadResp.Spread.LeftPageNumber)
	}

	// List spreads
	resp, err = http.Get(ts.URL + "/api/projects/" + project.ID + "/spreads")
	if err != nil {
		t.Fatalf("list spreads: %v", err)
	}
	var spreadsPayload struct {
		Spreads []struct {
			SpreadNumber int `json:"spread_number"`
		} `json:"spreads"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&spreadsPayload); err != nil {
		t.Fatalf("decode spreads list: %v", err)
	}
	_ = resp.Body.Close()
	if len(spreadsPayload.Spreads) != 1 || spreadsPayload.Spreads[0].SpreadNumber != 1 {
		t.Fatalf("unexpected spreads payload: %#v", spreadsPayload)
	}

	// Delete spread then page
	req, _ = http.NewRequest(http.MethodDelete, spreadURL, nil)
	if resp, err = http.DefaultClient.Do(req); err != nil {
		t.Fatalf("delete spread: %v", err)
	} else {
		resp.Body.Close()
	}

	req, _ = http.NewRequest(http.MethodDelete, pageURL, nil)
	if resp, err = http.DefaultClient.Do(req); err != nil {
		t.Fatalf("delete page: %v", err)
	} else {
		resp.Body.Close()
	}

	// Ensure lists are empty again
	resp, err = http.Get(ts.URL + "/api/projects/" + project.ID + "/spreads")
	if err != nil {
		t.Fatalf("list spreads (post-delete): %v", err)
	}
	if err := json.NewDecoder(resp.Body).Decode(&spreadsPayload); err != nil {
		t.Fatalf("decode spreads post-delete: %v", err)
	}
	_ = resp.Body.Close()
	if len(spreadsPayload.Spreads) != 0 {
		t.Fatalf("expected no spreads, got %#v", spreadsPayload)
	}

	resp, err = http.Get(ts.URL + "/api/projects/" + project.ID + "/pages")
	if err != nil {
		t.Fatalf("list pages (post-delete): %v", err)
	}
	if err := json.NewDecoder(resp.Body).Decode(&pagesPayload); err != nil {
		t.Fatalf("decode pages post-delete: %v", err)
	}
	_ = resp.Body.Close()
	if len(pagesPayload.Pages) != 0 {
		t.Fatalf("expected no pages, got %#v", pagesPayload)
	}
}
