package app

import (
	"fmt"

	"github.com/gosnmp/gosnmp"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"ntsc.ac.cn/ta-registry/pkg/pb"
	rpc "ntsc.ac.cn/ta-registry/pkg/rpc"
	"ntsc.ac.cn/ta-time-source/pkg/counter"
	"ntsc.ac.cn/ta-time-source/pkg/cv"
	"ntsc.ac.cn/ta-time-source/pkg/sreceiver"
	"ntsc.ac.cn/ta-time-source/pkg/ws"
)

type TimeSourceApp struct {
	conf           *Config
	cv             *cv.Device
	gpReceiver     *sreceiver.SatelliteReceiver
	gbReceiver     *sreceiver.SatelliteReceiver
	mc             *counter.MultipleCounter
	rpcServer      *rpc.Server
	userCVSessions []*userCVSession
	snmp           *gosnmp.GoSNMP

	wss   *ws.WebsocketService
	wssSM *ws.SessionManager
}

func NewTimeSourceApp(conf *Config) (*TimeSourceApp, error) {
	cvDev, err := cv.NewDevice(conf.CVConfig.SerialPath)
	if err != nil {
		return nil, err
	}
	gpDev, err := sreceiver.NewDevice(conf.SRConfig.GPSSerialPath)
	if err != nil {
		return nil, err
	}
	gbDev, err := sreceiver.NewDevice(conf.SRConfig.BDSerialPath)
	if err != nil {
		return nil, err
	}
	mc, err := counter.NewMultipleCounter(conf.CounerConfig.Endpoint)
	if err != nil {
		return nil, err
	}
	snmpConf := gosnmp.Default
	snmpConf.Community = "1234qwer"
	snmpConf.Target = "127.0.0.1"
	snmpConf.Port = 1169
	app := TimeSourceApp{
		conf:           conf,
		cv:             cvDev,
		gpReceiver:     gpDev,
		gbReceiver:     gbDev,
		mc:             mc,
		snmp:           snmpConf,
		userCVSessions: make([]*userCVSession, 0),
	}
	rpcConf, err := rpc.GenServerRPCConfig(conf.CertPath, conf.RPCListener)
	if err != nil {
		return nil, err
	}
	rpcServ, err := rpc.NewServer(rpcConf, []grpc.ServerOption{
		grpc.StreamInterceptor(
			rpc.StreamServerInterceptor(rpc.CertCheckFunc)),
		grpc.UnaryInterceptor(
			rpc.UnaryServerInterceptor(rpc.CertCheckFunc)),
	}, func(g *grpc.Server) {
		pb.RegisterCommonViewDataServiceServer(g, &app)
	})
	if err != nil {
		return nil, fmt.Errorf("create grpc server failed: %v", err)
	}
	app.rpcServer = rpcServ
	wss, err := ws.NewWebsocketService(&ws.Config{
		BindAddr: conf.WSListener,
	})
	app.wssSM = wss.SessionManager
	if err != nil {
		return nil, err
	}
	app.wss = wss
	return &app, nil
}

func (tsa *TimeSourceApp) Start() chan error {
	errChan := make(chan error, 1)
	go func() {
		err := <-tsa.rpcServer.Start()
		errChan <- err
	}()
	go func() {
		if err := <-tsa.cv.Open(); err != nil {
			errChan <- err
			return
		}
	}()
	go func() {
		tsa.cv.ReadMsg(errChan, func(raw []byte, data []string) error {
			logrus.WithField("prefix", "service.cv").
				Debugf("read common view data size [%d],session size [%d]", len(raw), len(tsa.userCVSessions))
			for _, session := range tsa.userCVSessions {
				session.dataChan <- raw
			}
			return nil
		})
	}()
	tsa._startCounter(errChan)
	tsa._startSReceiver(errChan)
	go func() {
		err := <-tsa.wss.Start([]*ws.WSHandler{
			{Path: "/time/gps", Handler: tsa.wsGPSTimerHandler},
			{Path: "/time/bd", Handler: tsa.wsBDTimerHandler},
			{Path: "/time/bj", Handler: tsa.wsBJTimerHandler},
		})
		errChan <- err
	}()
	go tsa._startPushTime(tsa.gpReceiver.TimeChan(), ws.GPSTimeType)
	go tsa._startPushTime(tsa.gbReceiver.TimeChan(), ws.BDTimeType)
	return errChan
}

func (tsa *TimeSourceApp) _startPushTime(timeChan chan string, tt ws.TimeType) {
	for secondTime := range timeChan {
		for _, se := range tsa.wssSM.FindByTimeType(tt, ws.SessionTypeTime) {
			if se.Error() != nil || se.IsClose() {
				logrus.WithField("prefix", "service.ws").
					Infof("close session [%s]: %v", se.Addr(), se.Error())
				tsa.wssSM.Remove(se.Addr())
				continue
			}
			se.PushTime(secondTime)
		}
	}
}
