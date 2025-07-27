package handlers

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/ipfs/go-cid"
	mh "github.com/multiformats/go-multihash"
)

func UploadCompleteAPI(stateDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fileID := r.PathValue("fileID")
		filePath := filepath.Join(stateDir, fileID)

		c, err := calculateCIDv1(filePath)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to calculate CID: %v", err), http.StatusInternalServerError)
			return
		}

		cidStr := c.String()
		newPath := filepath.Join(stateDir, cidStr)

		if err := os.Rename(filePath, newPath); err != nil {
			http.Error(w, fmt.Sprintf("rename failed: %v", err), http.StatusInternalServerError)
			return
		}

		w.Write(c.Bytes())
	}
}

func calculateCIDv1(filePath string) (cid.Cid, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return cid.Cid{}, err
	}
	defer f.Close()

	hasher := sha256.New()
	buf := make([]byte, 256*1024)
	for {
		n, err := f.Read(buf)
		if n > 0 {
			hasher.Write(buf[:n])
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return cid.Cid{}, err
		}
	}

	digest := hasher.Sum(nil)
	multiHash, err := mh.Encode(digest, mh.SHA2_256)
	if err != nil {
		return cid.Cid{}, err
	}

	c := cid.NewCidV1(cid.Raw, multiHash)
	return c, nil
}
