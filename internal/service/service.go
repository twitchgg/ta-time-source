package service

import (
	"context"
	"fmt"
	"net"

	"github.com/denisbrodbeck/machineid"
	"github.com/sirupsen/logrus"
	"ntsc.ac.cn/ta-registry/pkg/pb"
	"ntsc.ac.cn/ta-registry/pkg/rpc"
	"ntsc.ac.cn/ta-time-source/pkg/ws"
)

type DataService struct {
	conf             *Config
	cvServiceClient  pb.CommonViewDataServiceClient
	cvMainDataClient pb.CommonViewDataService_PushMainStationDataClient
	machineID        string
	tcpSessions      []*tcpSession
	wss              *ws.WebsocketService
}

func NewDataService(conf *Config) (*DataService, error) {
	if conf == nil {
		return nil, fmt.Errorf("rpc server config is not define")
	}
	machineID, err := machineid.ID()
	if err != nil {
		return nil, fmt.Errorf("generate machine id failed: %v", err)
	}
	if err := conf.Check(); err != nil {
		return nil, fmt.Errorf("check config failed: %v", err)
	}
	tlsConf, err := rpc.GetTlsConfig(machineID, conf.CertPath, conf.ServerName)
	if err != nil {
		return nil, fmt.Errorf("generate tls config failed: %v", err)
	}
	conn, err := rpc.DialRPCConn(&rpc.DialOptions{
		RemoteAddr: conf.ServiceEndpoint,
		TLSConfig:  tlsConf,
	})
	if err != nil {
		return nil, fmt.Errorf(
			"dial management grpc connection failed: %v", err)
	}
	wss, err := ws.NewWebsocketService(&ws.Config{
		BindAddr: conf.WSListener,
	})
	if err != nil {
		return nil, err
	}
	return &DataService{
		conf:            conf,
		machineID:       machineID,
		cvServiceClient: pb.NewCommonViewDataServiceClient(conn),
		tcpSessions:     make([]*tcpSession, 0),
		wss:             wss,
	}, nil
}

func (ds *DataService) Start() chan error {
	errChan := make(chan error, 1)
	var err error
	if ds.cvMainDataClient, err = ds.cvServiceClient.PushMainStationData(
		context.Background(), &pb.PushRequest{
			MachineID: ds.machineID,
		}); err != nil {
		errChan <- fmt.Errorf(
			"failed to create common view data service client: %v", err)
		return errChan
	}

	go func() {
		for {
			reply, err := ds.cvMainDataClient.Recv()
			if err != nil {
				logrus.WithField("prefix", "service.rpc").
					Errorf("failed to receive common view data: %v", err)
			}
			for _, session := range ds.tcpSessions {
				if _, err := session.conn.Write(reply.Data); err != nil {
					logrus.WithField("prefix", "service.rpc").
						Errorf("faile to write common view data: %v", err)
				}
			}
		}
	}()

	func() {
		laddr, err := net.ResolveTCPAddr("tcp", ds.conf.CVDataListener)
		if err != nil {
			errChan <- fmt.Errorf("failed to resolve tcp addr: %v", err)
			return
		}
		lis, err := net.ListenTCP("tcp", laddr)
		if err != nil {
			errChan <- fmt.Errorf("failed to create cv listener: %v", err)
			return
		}
		for {
			conn, err := lis.AcceptTCP()
			if err != nil {
				logrus.WithField("prefix", "service.rpc").
					Errorf("failed to accept client: %v", err)
				continue
			}
			tcpSession := &tcpSession{
				conn: conn,
			}
			go func() {
				if err := tcpSession.receive(); err != nil {
					logrus.WithField("prefix", "service.rpc").
						Errorf("%v", err)
					tcpSession.conn.Close()
				}
			}()
			ds.tcpSessions = append(ds.tcpSessions, tcpSession)
			logrus.WithField("prefix", "service.rpc").Infof("create tcp client: %s", conn.RemoteAddr())
		}
	}()
	go func() {
		err := <-ds.wss.Start()
		errChan <- err
	}()
	return errChan
}
