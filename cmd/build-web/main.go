package main

import (
    "context"
    "errors"
    "fmt"
    "log"
    "os"
    "path/filepath"
    "strings"

    "dagger.io/dagger"
)

func findRepoRoot(start string) (string, error) {
    dir := start
    for i := 0; i < 5; i++ {
        if _, err := os.Stat(filepath.Join(dir, "web", "package.json")); err == nil {
            return dir, nil
        }
        nd := filepath.Dir(dir)
        if nd == dir {
            break
        }
        dir = nd
    }
    return "", errors.New("could not locate repo root (web/package.json not found)")
}

func main() {
    pnpmVersion := os.Getenv("WEB_PNPM_VERSION")
    if pnpmVersion == "" {
        pnpmVersion = "10.15.0"
    }

    ctx := context.Background()
    client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
    if err != nil {
        log.Fatalf("connect dagger: %v", err)
    }
    defer func() { _ = client.Close() }()

    wd, err := os.Getwd()
    if err != nil {
        log.Fatalf("getwd: %v", err)
    }

    repoRoot, err := findRepoRoot(wd)
    if err != nil {
        log.Fatalf("%v (start=%s)", err, wd)
    }
    webPath := filepath.Join(repoRoot, "web")
    outPath := filepath.Join(wd, "dist")

    webDir := client.Host().Directory(webPath)

    baseImage := "node:22"
    if bi := os.Getenv("WEB_BUILDER_IMAGE"); bi != "" {
        valid := false
        if strings.Contains(bi, "@") {
            valid = true
        } else if idx := strings.LastIndex(bi, ":"); idx > 0 && idx < len(bi)-1 {
            valid = true
        }
        if valid {
            baseImage = bi
        } else {
            log.Printf("Ignoring invalid WEB_BUILDER_IMAGE=%q; fallback to %s", bi, baseImage)
        }
    }

    base := client.Container().From(baseImage)
    if strings.HasPrefix(baseImage, "ghcr.io/") {
        user := os.Getenv("REGISTRY_USER")
        if user == "" {
            user = os.Getenv("GHCR_USERNAME")
        }
        token := os.Getenv("REGISTRY_TOKEN")
        if token == "" {
            token = os.Getenv("GHCR_TOKEN")
        }
        if user != "" && token != "" {
            sec := client.SetSecret("ghcr_token", token)
            base = base.WithRegistryAuth("ghcr.io", user, sec)
        }
    }

    pnpmCacheDir := os.Getenv("PNPM_CACHE_DIR")
    var ctr *dagger.Container
    if pnpmCacheDir != "" {
        if !filepath.IsAbs(pnpmCacheDir) {
            abs, err := filepath.Abs(pnpmCacheDir)
            if err != nil {
                log.Fatalf("resolve PNPM_CACHE_DIR: %v", err)
            }
            pnpmCacheDir = abs
        }
        hostCache := client.Host().Directory(pnpmCacheDir)
        ctr = base.
            WithMountedDirectory("/src", webDir).
            WithWorkdir("/src").
            WithEnvVariable("PNPM_HOME", "/pnpm").
            WithMountedDirectory("/pnpm/store", hostCache)
    } else {
        pnpmCache := client.CacheVolume("pnpm-store")
        ctr = base.
            WithMountedDirectory("/src", webDir).
            WithWorkdir("/src").
            WithEnvVariable("PNPM_HOME", "/pnpm").
            WithMountedCache("/pnpm/store", pnpmCache)
    }

    if os.Getenv("WEB_BUILDER_IMAGE") == "" || !strings.Contains(os.Getenv("WEB_BUILDER_IMAGE"), ":") {
        ctr = ctr.WithExec([]string{"sh", "-lc", fmt.Sprintf("corepack enable && corepack prepare pnpm@%s --activate", pnpmVersion)})
    }
    ctr = ctr.
        WithExec([]string{"sh", "-lc", "pnpm --version"}).
        WithExec([]string{"sh", "-lc", "pnpm install --store-dir /pnpm/store --reporter=append-only"}).
        WithExec([]string{"sh", "-lc", "pnpm build"})

    dist := ctr.Directory("/src/dist")
    if _, err := dist.Export(ctx, outPath); err != nil {
        log.Fatalf("export dist: %v", err)
    }
    log.Printf("exported web dist to %s", outPath)
}

