package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/google/uuid"
)

type uploadResp struct {
	ID     string `json:"id"`
	Offset int64  `json:"offset"`
}

func UploadAPI(stateDir string, maxUploadSize int64) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
		err := r.ParseMultipartForm(maxUploadSize)
		if err != nil {
			slog.Warn("Chunk size too big!", "max", maxUploadSize)
			http.Error(w, http.StatusText(http.StatusRequestEntityTooLarge), http.StatusRequestEntityTooLarge)
			return
		}

		offsetStr := r.FormValue("offset")
		if offsetStr == "" {
			http.Error(w, "missing offset", http.StatusBadRequest)
			return
		}
		offset, err := strconv.ParseInt(offsetStr, 10, 64)
		if err != nil || offset < 0 {
			http.Error(w, "invalid offset", http.StatusBadRequest)
			return
		}

		filePart, _, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "missing file part", http.StatusBadRequest)
			return
		}
		defer filePart.Close()

		var (
			id       string
			fullPath string
			f        *os.File
		)

		if r.Method == http.MethodPost {
			if offset != 0 {
				http.Error(w, "offset must be 0 when creating new file", http.StatusBadRequest)
				return
			}

			id = uuid.New().String()
			fullPath = filepath.Join(stateDir, id)
			f, err = os.OpenFile(fullPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)

			if err != nil {
				slog.Error("failed to create file", "path", fullPath, "err", err)
				http.Error(w, "failed to create file", http.StatusInternalServerError)
				return
			}
		} else {
			id = r.PathValue("fileID")
			slog.Info("File ID given!", "ID", id)
			if _, err := uuid.Parse(id); err != nil {
				http.Error(w, "invalid fileID", http.StatusBadRequest)
				return
			}

			fullPath = filepath.Join(stateDir, id)
			f, err = os.OpenFile(fullPath, os.O_WRONLY, 0o644)
			if errors.Is(err, os.ErrNotExist) {
				http.Error(w, "upload not found", http.StatusNotFound)
				return
			}

			if err != nil {
				slog.Error("failed to open file", "path", fullPath, "err", err)
				http.Error(w, "failed to open file", http.StatusInternalServerError)
				return
			}
		}
		defer f.Close()

		if _, err := f.Seek(offset, io.SeekStart); err != nil {
			slog.Error("seek failed", "path", fullPath, "offset", offset, "err", err)
			http.Error(w, "seek failed", http.StatusInternalServerError)
			return
		}
		n, err := io.Copy(f, filePart)
		if err != nil {
			slog.Error("write failed", "path", fullPath, "err", err)
			http.Error(w, "write failed", http.StatusInternalServerError)
			return
		}

		resp := uploadResp{
			ID:     id,
			Offset: offset + n,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}
}
