package cv

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/jacobsa/go-serial/serial"
	"github.com/sirupsen/logrus"
)

const (
	// DefaultCVHeaderSize default serial port data header size
	DefaultCVHeaderSize = 7
	// DefaultSatelliteDataSize default satellite data size
	DefaultSatelliteDataSize = 24
)

type Device struct {
	serialPath string
	opts       serial.OpenOptions
	msgChan    chan serialData
	sp         io.ReadWriteCloser
}

type serialData struct {
	Data []string
	Raw  []byte
}

func NewDevice(path string) (*Device, error) {
	return &Device{
		serialPath: path,
		msgChan:    make(chan serialData),
		opts: serial.OpenOptions{
			PortName:        path,
			BaudRate:        9600,
			DataBits:        8,
			StopBits:        1,
			MinimumReadSize: 1,
		},
	}, nil
}

func (d *Device) Open() chan error {
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
			line, err := reader.ReadString('.')
			if err != nil {
				errChan <- fmt.Errorf("failed to read data: %v", err)
				return
			}
			line = strings.TrimSpace(line)
			data := line[1 : len(line)-1]
			srcData := serialData{
				Data: strings.Split(data, ","),
				Raw:  []byte(line),
			}
			d.msgChan <- srcData
		}
	}()
	return errChan
}

// ReadMsg read cv data message
func (d *Device) ReadMsg(errChan chan error, fn func([]byte, []string) error) {
	go func() {
		logrus.WithField("prefix", "cv.device.read").
			Infof("start read CV data with serial port [%s]", d.serialPath)
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

// Close close common view device
func (d *Device) Close() error {
	if d.sp == nil {
		return nil
	}
	return d.sp.Close()
}
