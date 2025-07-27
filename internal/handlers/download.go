package handlers

import (
	"net/http"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

func DownloadAPI(stateDir string, node *pubsub.PubSub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}
