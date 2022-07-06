package sreceiver

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/jacobsa/go-serial/serial"
	"github.com/sirupsen/logrus"
)

type serialData struct {
	Data []string
	Raw  []byte
}

type SatelliteReceiver struct {
	serialPath string
	opts       serial.OpenOptions
	msgChan    chan serialData
	sp         io.ReadWriteCloser
	timeChan   chan string
	prefix     string
}

func NewDevice(path string) (*SatelliteReceiver, error) {
	return &SatelliteReceiver{
		serialPath: path,
		msgChan:    make(chan serialData),
		timeChan:   make(chan string),
		opts: serial.OpenOptions{
			PortName:        path,
			BaudRate:        115200,
			DataBits:        8,
			StopBits:        1,
			MinimumReadSize: 1,
		},
	}, nil
}

func (d *SatelliteReceiver) Open(prefix string) chan error {
	d.prefix = prefix
	errChan := make(chan error, 1)
	var err error
	if d.sp, err = serial.Open(d.opts); err != nil {
		errChan <- fmt.Errorf(
			"failed to open device [%s]: %v", d.serialPath, err)
		return errChan
	}
	go func() {
		reader := bufio.NewReader(d.sp)
		for {
			raw, _, err := reader.ReadLine()
			if err != nil {
				errChan <- fmt.Errorf("failed to read data: %v", err)
				return
			}

			line := strings.TrimSpace(string(raw))
			if !strings.HasPrefix(line, prefix) {
				continue
			}
			logrus.WithField("prefix", "cv.device.open").
				Tracef("read [%s | %s] data: %s", d.serialPath, prefix, line)
			srcData := serialData{
				Data: strings.Split(line, ","),
				Raw:  []byte(line),
			}
			d.msgChan <- srcData
		}
	}()
	return errChan
}

// ReadMsg read cv data message
func (d *SatelliteReceiver) ReadMsg(errChan chan error, fn func([]byte, []string) error) {
	go func() {
		logrus.WithField("prefix", "cv.device.read").
			Infof("start read data with serial port [%s]", d.serialPath)
		for m := range d.msgChan {
			if fn == nil {
				continue
			}
			if err := fn(m.Raw, m.Data); err != nil {
				logrus.WithError(err).Errorf("cv data process failed")
			}
		}
	}()
}

// ReadMsg read cv data message
func (d *SatelliteReceiver) ReadTime(errChan chan error) {
	go func() {
		logrus.WithField("prefix", "cv.device.read").
			Infof("start read data with serial port [%s]", d.serialPath)
		for m := range d.msgChan {
			if len(m.Data) != 14 {
				logrus.WithField("prefix", "cv.device.read").
					Warnf("bad data len [%d] with serial port [%s]", len(m.Data), d.serialPath)
				continue
			}
			timeDataStr := m.Data[1]
			utcTimeData := fmt.Sprintf("%0.2v:%s:%s+00:00", timeDataStr[0:2], timeDataStr[2:4], timeDataStr[4:6])
			format := "15:04:05Z07:00"
			t1, err := time.Parse(format, utcTimeData)
			if err != nil {
				logrus.WithField("prefix", "cv.device.read").
					Warnf("failed to parse time data: %v", err)
				continue
			}
			t1Data := fmt.Sprintf("%0.2d:%0.2d:%0.2d", t1.Hour(), t1.Minute(), t1.Second())
			logrus.WithField("prefix", "cv.device.read").
				Tracef("dev [%s] time data: %s", d.prefix, t1Data)
		}
	}()
}

// Close close common view device
func (d *SatelliteReceiver) Close() error {
	if d.sp == nil {
		return nil
	}
	return d.sp.Close()
}
