package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/uwufile/uwufile/connectivity"
	"github.com/uwufile/uwufile/internal/handlers"
)

func main() {
	addr := flag.String("listen", "127.0.0.1:8080", "specify the LISTEN address")
	flag.Parse()

	stateDir := "./state"
	maxUploadSize := 2 * 1024 * 1024 * 1024

	// TODO: should bootstrap to known host
	_, node, err := connectivity.MakeNode(context.Background())
	if err != nil {
		slog.Error("Failed setting up peer discovery", "error", err)
		os.Exit(1)
	}

	router := http.NewServeMux()

	router.Handle("POST /", handlers.UploadAPI(stateDir, int64(maxUploadSize)))
	router.Handle("PUT /{fileID}", handlers.UploadAPI(stateDir, int64(maxUploadSize)))
	router.Handle("POST /{fileID}/complete", handlers.UploadCompleteAPI(stateDir))
	router.Handle("GET /{fileID}", handlers.DownloadAPI(stateDir, node))

	s := &http.Server{
		Addr:           *addr,
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   0,
		MaxHeaderBytes: 1 << 20,
	}

	slog.Info("HTTP server listening on " + *addr)
	err = s.ListenAndServe()
	if err != nil {
		slog.Error("Error running HTTP server", "error", err)
		os.Exit(1)
	}
}
