package handlers

import (
	"net/http"

	iface "github.com/ipfs/kubo/core/coreiface"
)

func DownloadAPI(stateDir string, api iface.CoreAPI) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}
