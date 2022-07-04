package ws

import (
	"net/http"

	"github.com/gorilla/websocket"
)

var defaultWSUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
