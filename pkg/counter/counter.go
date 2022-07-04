package counter

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	DEFAULT_BUF_SIZE = 2048
)

type PPSData struct {
	NTSC   uint64
	Clock  uint64
	GPS    uint64
	Beidou uint64
}

func (pd *PPSData) String() string {
	return fmt.Sprintf("NTSC: %d Clock: %d GPS: %d Beidou: %d",
		pd.NTSC, pd.Clock, pd.GPS, pd.Beidou)
}

type MultipleCounter struct {
	addr    string
	lis     *net.UDPConn
	ppsData chan PPSData
}

func NewMultipleCounter(addr string) (*MultipleCounter, error) {
	if addr == "" {
		return nil, fmt.Errorf("counter addr not define")
	}
	return &MultipleCounter{
		addr:    addr,
		ppsData: make(chan PPSData),
	}, nil
}

func (c *MultipleCounter) Start() chan error {
	errChan := make(chan error, 1)
	addr, err := net.ResolveUDPAddr("udp", c.addr)
	if err != nil {
		errChan <- fmt.Errorf("failed to resolve udp addr: %v", err)
		return errChan
	}
	laddr, _ := net.ResolveUDPAddr("udp", "0.0.0.0:45454")
	c.lis, err = net.DialUDP("udp", laddr, addr)
	if err != nil {
		errChan <- fmt.Errorf(
			"failed to dial udp to address [%s]: %v", c.addr, err)
		return errChan
	}
	go func() {
		if _, err = c.lis.Write([]byte("hello")); err != nil {
			errChan <- fmt.Errorf(
				"failed to write data to address [%s]: %v", c.addr, err)
			return
		}
		logrus.WithField("prefix", "counter").
			Infof("start counter data receiver: %s", c.addr)
		buf := make([]byte, DEFAULT_BUF_SIZE)
		for {
			c.lis.SetDeadline(time.Now().Add(time.Second * 3))
			rn, err := c.lis.Read(buf)
			if err != nil {
				logrus.WithField("prefix", "counter").
					Errorf("failed to read udp data: %v", err)
				c.lis.Close()
				return
			}
			data := buf[:rn]
			dataStr := strings.TrimSpace(string(data))

			datas := strings.Split(dataStr, ",")
			datas = datas[1:]
			ntscPPS, err := strconv.ParseFloat(datas[0], 64)
			if err != nil {
				ntscPPS = 0
			}
			clockPPS, err := strconv.ParseFloat(datas[1], 64)
			if err != nil {
				clockPPS = 0
			}
			gpsPPS, err := strconv.ParseFloat(datas[2], 64)
			if err != nil {
				gpsPPS = 0
			}
			bdPPS, err := strconv.ParseFloat(datas[3], 64)
			if err != nil {
				bdPPS = 0
			}
			c.ppsData <- PPSData{
				NTSC:   uint64(ntscPPS * 1000),
				Clock:  uint64(clockPPS * 1000),
				GPS:    uint64(gpsPPS * 1000),
				Beidou: uint64(bdPPS * 1000),
			}
		}
	}()
	return errChan
}

func (c *MultipleCounter) ReadData() chan PPSData {
	return c.ppsData
}
