package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/ipfs/boxo/files"
	iface "github.com/ipfs/kubo/core/coreiface"
)

func UploadCompleteAPI(ctx context.Context, stateDir string, api iface.CoreAPI) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fileID := r.PathValue("fileID")
		filePath := filepath.Join(stateDir, fileID)

		_, err := os.Stat(filePath)
		if os.IsNotExist(err) {
			slog.Error("Could not find upload", "file ID", fileID, "error", err)
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		} else if err != nil {
			slog.Error("Failed checking if file exists", "file path", filePath, "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		file, err := os.Open(filePath)
		if err != nil {
			slog.Error("Failed reading file", "file path", filePath, "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		defer file.Close()

		ipfsFile := files.NewReaderFile(file)
		cidFile, err := api.Unixfs().Add(ctx, ipfsFile)
		if err != nil {
			slog.Error("Failed adding file to IPFS instance", "file path", filePath, "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		err = api.Pin().Add(ctx, cidFile)
		if err != nil {
			slog.Error("Failed pinning file to IPFS instance", "file path", filePath, "cid", cidFile.RootCid(), "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.Write([]byte(cidFile.RootCid().String()))
	}
}
