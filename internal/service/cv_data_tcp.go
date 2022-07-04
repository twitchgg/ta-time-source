package service

import (
	"fmt"
	"net"
)

type tcpSession struct {
	conn *net.TCPConn
}

func (ts *tcpSession) receive() error {
	buf := make([]byte, 2048)
	for {
		n, err := ts.conn.Read(buf)
		if err != nil {
			return fmt.Errorf("failed to read tcp data: %v", err)
		}
		data := buf[:n]
		fmt.Println("====", string(data))
	}
}
