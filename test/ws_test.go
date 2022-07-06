package test

import (
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

func TestWSClient(t *testing.T) {
	done := make(chan struct{})
	go wsClient("ws://10.25.133.17:8788/time/bd", "北斗时间")
	go wsClient("ws://10.25.133.17:8788/time/bj", "北京时间")
	go wsClient("ws://10.25.133.17:8788/time/gps", "GPS时间")
	<-done
}

func wsClient(url, prefix string) {
	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		panic(err)
	}
	defer c.Close()
	time.Sleep(time.Second)
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			logrus.WithField("prefix", "test").Fatal(err)
		}
		logrus.WithField("prefix", "test").
			Tracef("[%s] recv: %s", prefix, string(message))
	}
}
