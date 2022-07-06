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
	conf           *Config
	SessionManager *SessionManager
}

func NewWebsocketService(conf *Config) (*WebsocketService, error) {
	sm := SessionManager{
		sessions: make([]*Session, 0),
	}
	return &WebsocketService{
		conf:           conf,
		SessionManager: &sm,
	}, nil
}

func (wss *WebsocketService) Start(handlers []*WSHandler) chan error {
	errChan := make(chan error, 1)
	for _, handler := range handlers {
		http.HandleFunc(handler.Path, handler.Handler)
	}
	go func() {
		logrus.WithField("prefix", "ws").
			Infof("start http ws server with: %s", wss.conf.BindAddr)
		if err := http.ListenAndServe(wss.conf.BindAddr, nil); err != nil {
			errChan <- err
		}
	}()
	return errChan
}
