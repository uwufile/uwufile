package handlers

import (
	"context"
	"io"
	"log/slog"
	"net/http"

	"github.com/ipfs/boxo/files"
	ipfspath "github.com/ipfs/boxo/path"
	iface "github.com/ipfs/kubo/core/coreiface"
)

func DownloadAPI(ctx context.Context, stateDir string, api iface.CoreAPI) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fileID := r.PathValue("fileID")

		path, err := ipfspath.NewPath("/ipfs/" + fileID)
		if err != nil {
			slog.Error("Failed to parse CID path", "cid", fileID, "error", err)
			http.Error(w, "Invalid CID", http.StatusBadRequest)
			return
		}

		nodeFile, err := api.Unixfs().Get(ctx, path)
		if err != nil {
			slog.Error("Failed to get file from IPFS", "cid", fileID, "error", err)
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}

		file, ok := nodeFile.(files.File)
		if !ok {
			slog.Error("CID does not point to a regular file", "cid", fileID)
			http.Error(w, "CID does not point to a file", http.StatusUnsupportedMediaType)
			return
		}

		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", "attachment; filename=\""+fileID+"\"")

		_, err = io.Copy(w, file)
		if err != nil {
			slog.Error("Failed to write file to response", "cid", fileID, "error", err)
			http.Error(w, "Failed to stream file", http.StatusInternalServerError)
			return
		}
	}
}
