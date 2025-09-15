package main

import (
    "flag"
    "log"
    "net/http"
    "os"
    "path/filepath"
)

func main() {
    var (
        root = flag.String("root", "./dist", "path to built web assets (dist)")
        addr = flag.String("addr", ":8080", "listen address")
    )
    flag.Parse()

    mux := http.NewServeMux()

    mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        _, _ = w.Write([]byte(`{"ok":true}`))
    })

    // serve static web files
    abs, err := filepath.Abs(*root)
    if err != nil {
        log.Fatalf("resolve root: %v", err)
    }
    if _, err := os.Stat(abs); err != nil {
        log.Printf("warning: web dist not found at %s", abs)
    }
    mux.Handle("/", http.FileServer(http.Dir(abs)))

    log.Printf("serving on %s (web from %s)", *addr, abs)
    if err := http.ListenAndServe(*addr, mux); err != nil {
        log.Fatal(err)
    }
}


