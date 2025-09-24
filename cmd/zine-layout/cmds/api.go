package cmds

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "os"
    "os/exec"
    "path/filepath"
    "strings"

    "github.com/spf13/cobra"
    "github.com/go-go-golems/zine-layout/cmd/zine-layout/cmds/api"
)

// NewAPICobraCommand creates the `api` command group with subcommands to hit server endpoints.
func NewAPICobraCommand() (*cobra.Command, error) {
    root := &cobra.Command{
        Use:   "api",
        Short: "API client utilities for the zine-layout server",
    }

    // yaml-render
    var yamlFile, baseDir, server, out string
    var open bool
    cmdYamlRender := &cobra.Command{
        Use:   "yaml-render",
        Short: "Render a Simple YAML payload via /api/v1/yaml/render and save HTML",
        RunE: func(cmd *cobra.Command, args []string) error {
            if yamlFile == "" {
                return fmt.Errorf("--yaml-file is required")
            }
            if out == "" {
                out = filepath.Join("/home/manuel/tmp", "simple-yaml-render.html")
            }
            data, err := os.ReadFile(yamlFile)
            if err != nil {
                return err
            }
            body := map[string]any{"yaml": string(data)}
            if baseDir != "" {
                body["base_dir"] = baseDir
            }
            respBytes, err := httpPostJSON(fmt.Sprintf("%s/api/v1/yaml/render", server), body)
            if err != nil {
                return err
            }
            var resp struct{
                HTML string `json:"html"`
            }
            _ = json.Unmarshal(respBytes, &resp)
            if strings.TrimSpace(resp.HTML) == "" {
                // fallback: write entire JSON for debugging
                resp.HTML = fmt.Sprintf("<pre>%s</pre>", string(respBytes))
            }
            if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
                return err
            }
            if err := os.WriteFile(out, []byte(resp.HTML), 0o644); err != nil {
                return err
            }
            fmt.Printf("HTML saved: %s\n", out)
            if open {
                _ = exec.Command("xdg-open", out).Start()
            }
            return nil
        },
    }
    cmdYamlRender.Flags().StringVar(&yamlFile, "yaml-file", "", "Path to Simple YAML file")
    cmdYamlRender.Flags().StringVar(&baseDir, "base-dir", "", "Base directory for resolving image paths")
    cmdYamlRender.Flags().StringVar(&server, "server", "http://localhost:8088", "Server base URL")
    cmdYamlRender.Flags().StringVar(&out, "out", filepath.Join("/home/manuel/tmp", "simple-yaml-render.html"), "Output HTML file path")
    cmdYamlRender.Flags().BoolVar(&open, "open", false, "Open HTML after saving")
    root.AddCommand(cmdYamlRender)

    // compute
    var jsonFile string
    cmdCompute := &cobra.Command{
        Use:   "compute",
        Short: "POST /api/v1/compute with a JSON request file and print the result",
        RunE: func(cmd *cobra.Command, args []string) error {
            if jsonFile == "" {
                return fmt.Errorf("--json-file is required")
            }
            b, err := os.ReadFile(jsonFile)
            if err != nil {
                return err
            }
            respBytes, err := httpPostRaw(fmt.Sprintf("%s/api/v1/compute", server), b)
            if err != nil {
                return err
            }
            var out map[string]any
            _ = json.Unmarshal(respBytes, &out)
            pretty, _ := json.MarshalIndent(out, "", "  ")
            fmt.Println(string(pretty))
            return nil
        },
    }
    cmdCompute.Flags().StringVar(&jsonFile, "json-file", "", "Path to JSON file (ComputeRequest)")
    cmdCompute.Flags().StringVar(&server, "server", "http://localhost:8088", "Server base URL")
    root.AddCommand(cmdCompute)

    // preview
    var previewOut string
    cmdPreview := &cobra.Command{
        Use:   "preview",
        Short: "POST /api/v1/preview with a JSON request file and save the image",
        RunE: func(cmd *cobra.Command, args []string) error {
            if jsonFile == "" {
                return fmt.Errorf("--json-file is required")
            }
            if previewOut == "" {
                previewOut = filepath.Join("/home/manuel/tmp", "preview.png")
            }
            b, err := os.ReadFile(jsonFile)
            if err != nil {
                return err
            }
            url := fmt.Sprintf("%s/api/v1/preview", server)
            req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))
            if err != nil {
                return err
            }
            req.Header.Set("Content-Type", "application/json")
            resp, err := http.DefaultClient.Do(req)
            if err != nil {
                return err
            }
            defer resp.Body.Close()
            if resp.StatusCode < 200 || resp.StatusCode >= 300 {
                data, _ := io.ReadAll(resp.Body)
                return fmt.Errorf("preview http %d: %s", resp.StatusCode, string(data))
            }
            if err := os.MkdirAll(filepath.Dir(previewOut), 0o755); err != nil {
                return err
            }
            f, err := os.Create(previewOut)
            if err != nil {
                return err
            }
            defer f.Close()
            if _, err := io.Copy(f, resp.Body); err != nil {
                return err
            }
            fmt.Printf("Preview saved: %s\n", previewOut)
            if open {
                _ = exec.Command("xdg-open", previewOut).Start()
            }
            return nil
        },
    }
    cmdPreview.Flags().StringVar(&jsonFile, "json-file", "", "Path to JSON file (ComputeRequest)")
    cmdPreview.Flags().StringVar(&previewOut, "out", filepath.Join("/home/manuel/tmp", "preview.png"), "Output image path")
    cmdPreview.Flags().StringVar(&server, "server", "http://localhost:8088", "Server base URL")
    cmdPreview.Flags().BoolVar(&open, "open", false, "Open image after saving")
    root.AddCommand(cmdPreview)

    // render
    cmdRender := &cobra.Command{
        Use:   "render",
        Short: "POST /api/v1/render with a JSON request file and print paths",
        RunE: func(cmd *cobra.Command, args []string) error {
            if jsonFile == "" {
                return fmt.Errorf("--json-file is required")
            }
            b, err := os.ReadFile(jsonFile)
            if err != nil {
                return err
            }
            respBytes, err := httpPostRaw(fmt.Sprintf("%s/api/v1/render", server), b)
            if err != nil {
                return err
            }
            var out map[string]any
            _ = json.Unmarshal(respBytes, &out)
            pretty, _ := json.MarshalIndent(out, "", "  ")
            fmt.Println(string(pretty))
            return nil
        },
    }
    cmdRender.Flags().StringVar(&jsonFile, "json-file", "", "Path to JSON file (ComputeRequest)")
    cmdRender.Flags().StringVar(&server, "server", "http://localhost:8088", "Server base URL")
    root.AddCommand(cmdRender)

    // yaml (build simple yaml)
    cmdYAML := &cobra.Command{
        Use:   "yaml",
        Short: "POST /api/v1/yaml with a JSON request file and print YAML",
        RunE: func(cmd *cobra.Command, args []string) error {
            if jsonFile == "" {
                return fmt.Errorf("--json-file is required")
            }
            b, err := os.ReadFile(jsonFile)
            if err != nil {
                return err
            }
            respBytes, err := httpPostRaw(fmt.Sprintf("%s/api/v1/yaml", server), b)
            if err != nil {
                return err
            }
            fmt.Println(string(respBytes))
            return nil
        },
    }
    cmdYAML.Flags().StringVar(&jsonFile, "json-file", "", "Path to JSON file (ComputeRequest)")
    cmdYAML.Flags().StringVar(&server, "server", "http://localhost:8088", "Server base URL")
    root.AddCommand(cmdYAML)

    // Add all the new API commands
    if err := api.AddAllAPICommands(root); err != nil {
        return nil, fmt.Errorf("failed to add API commands: %w", err)
    }

    return root, nil
}

func httpPostJSON(url string, v any) ([]byte, error) {
    b, err := json.Marshal(v)
    if err != nil {
        return nil, err
    }
    return httpPostRaw(url, b)
}

func httpPostRaw(url string, body []byte) ([]byte, error) {
    req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
    if err != nil {
        return nil, err
    }
    req.Header.Set("Content-Type", "application/json")
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    data, _ := io.ReadAll(resp.Body)
    if resp.StatusCode < 200 || resp.StatusCode >= 300 {
        return nil, fmt.Errorf("http %d: %s", resp.StatusCode, string(data))
    }
    return data, nil
}


