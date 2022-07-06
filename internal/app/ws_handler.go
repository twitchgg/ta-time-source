package app

import (
	"net/http"

	"github.com/sirupsen/logrus"
	"ntsc.ac.cn/ta-time-source/pkg/ws"
)

func (tsa *TimeSourceApp) wsGPSTimerHandler(w http.ResponseWriter, r *http.Request) {
	tsa.timerHandler(ws.GPSTimeType, w, r)
}

func (tsa *TimeSourceApp) wsBDTimerHandler(w http.ResponseWriter, r *http.Request) {
	tsa.timerHandler(ws.BDTimeType, w, r)
}

func (tsa *TimeSourceApp) wsBJTimerHandler(w http.ResponseWriter, r *http.Request) {
	tsa.timerHandler(ws.BDTimeType, w, r)
}

func (tsa *TimeSourceApp) timerHandler(t ws.TimeType, w http.ResponseWriter, r *http.Request) {
	session, err := tsa.wssSM.Start(t, w, r, ws.SessionTypeTime)
	if err != nil {
		logrus.WithError(err).Errorf("create client [%s] session failed")
		return
	}
	session.WaitForClose()
}
