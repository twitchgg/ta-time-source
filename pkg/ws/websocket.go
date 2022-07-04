package ws

import (
	"net/http"

	"github.com/sirupsen/logrus"
)

type WSHandlerFunc func(http.ResponseWriter, *http.Request)
type WSHandler struct {
	Path    string
	Handler WSHandlerFunc
}
type WebsocketService struct {
	conf *Config
}

func NewWebsocketService(conf *Config) (*WebsocketService, error) {
	return &WebsocketService{
		conf: conf,
	}, nil
}

func (wss *WebsocketService) Start(handlers ...WSHandler) chan error {
	errChan := make(chan error, 1)
	for _, handler := range handlers {
		http.HandleFunc(handler.Path, handler.Handler)
	}
	go func() {
		logrus.WithField("prefix", "ws").
			Debugf("start http ws server with: %s", wss.conf.BindAddr)
		if err := http.ListenAndServe(wss.conf.BindAddr, nil); err != nil {
			errChan <- err
		}
	}()
	return errChan
}
