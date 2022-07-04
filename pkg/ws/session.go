package ws

import (
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

type TimeStatus uint8

const (
	TimeStatusOK TimeStatus = iota
	TimeStatusFailed
)

func (t TimeStatus) String() string {
	switch t {
	case TimeStatusOK:
		return "ok"
	case TimeStatusFailed:
		return "failed"
	default:
		return "unknow"
	}
}

type SessionType int

const (
	SessionTypeTime SessionType = iota
	SessionTypeTimeStatus
)

type TimeType int
type Session struct {
	st         SessionType
	addr       string
	wsConn     *websocket.Conn
	timeChan   chan string
	timeStatus chan TimeStatus
	tt         TimeType
	err        error
	isClose    bool
}

func (s *Session) WaitForClose() {
	go func() {
		for {
			mt, _, err := s.wsConn.ReadMessage()
			if err != nil || mt == websocket.CloseMessage {
				s.Close()
				return
			}
		}
	}()
	switch s.st {
	case SessionTypeTime:
		for t := range s.timeChan {
			if s.isClose {
				return
			}
			if err := s.wsConn.WriteMessage(websocket.TextMessage, []byte(t)); err != nil {
				s.err = err
				logrus.WithField("prefix", "session.wait_for_close").Warnf("client failed: %s", err.Error())
				s.Close()
				return
			}
		}
	case SessionTypeTimeStatus:
		for t := range s.timeStatus {
			if s.isClose {
				return
			}
			if err := s.wsConn.WriteMessage(websocket.TextMessage, []byte(t.String())); err != nil {
				s.err = err
				logrus.WithField("prefix", "session.wait_for_close").Warnf("client failed: %s", err.Error())
				s.Close()
				return
			}
		}
	}

}

func (s *Session) Close() error {
	s.isClose = true
	return s.wsConn.Close()
}

func (s *Session) PushTime(t string) {
	s.timeChan <- t
}

func (s *Session) PushTimeStatus(status TimeStatus) {
	s.timeStatus <- status
}

type SessionManager struct {
	sessions []*Session
	lock     sync.Mutex
}

func (m *SessionManager) Start(tt TimeType, w http.ResponseWriter, r *http.Request, st SessionType) (*Session, error) {
	addr := r.RemoteAddr
	s := m.Find(addr)
	if s == nil {
		wsConn, err := defaultWSUpgrader.Upgrade(w, r, nil)
		if err != nil {
			logrus.WithError(err).Errorf("upgrader [%s] failed", r.RemoteAddr)
			if wsConn != nil {
				wsConn.Close()
			}
			return nil, err
		}
		s = &Session{
			addr:   addr,
			wsConn: wsConn,
			// timeChan: make(chan string),
			tt: tt,
			st: st,
		}
		if st == SessionTypeTime {
			s.timeChan = make(chan string)
		}
		if st == SessionTypeTimeStatus {
			s.timeStatus = make(chan TimeStatus)
		}
		m.sessions = append(m.sessions, s)
	}
	return s, nil
}

func (m *SessionManager) FindByTimeType(tt TimeType, st SessionType) []*Session {
	var ss []*Session
	for _, s := range m.sessions {
		if s.tt == tt && s.st == st {
			ss = append(ss, s)
		}
	}
	return ss
}

func (m *SessionManager) Remove(addr string) {
	m.lock.Lock()
	defer m.lock.Unlock()
	for i, s := range m.sessions {
		if s.addr == addr {
			m.sessions = append(m.sessions[:i], m.sessions[i+1:]...)
		}
	}
}

func (m *SessionManager) Find(addr string) *Session {
	for _, se := range m.sessions {
		if se.addr == addr {
			return se
		}
	}
	return nil
}
