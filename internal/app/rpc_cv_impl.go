package app

import (
	"strings"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"ntsc.ac.cn/tas/tas-commons/pkg/pb"
	rpc "ntsc.ac.cn/tas/tas-commons/pkg/rpc"
)

type userCVSession struct {
	machineID string
	dataChan  chan []byte
}

func (tsa *TimeSourceApp) findUserCVSession(machineID string) *userCVSession {
	for _, v := range tsa.userCVSessions {
		if v.machineID == machineID {
			return v
		}
	}
	return nil
}

func (tsa *TimeSourceApp) PushStationData(req *pb.PushRequest,
	stream pb.CommonViewDataService_PushStationDataServer) error {
	logrus.WithField("prefix", "app.cv").Infof(
		"connect device [%s]", req.MachineID)
	if err := rpc.CheckMachineID(stream.Context(), req.MachineID); err != nil {
		return err
	}
	session := tsa.findUserCVSession(req.MachineID)
	if session == nil {
		session = &userCVSession{
			machineID: req.MachineID,
			dataChan:  make(chan []byte),
		}
		tsa.userCVSessions = append(tsa.userCVSessions, session)
		logrus.WithField("prefix", "app.cv").Infof(
			"create session by machine-id [%s]", req.MachineID)
	}
	for data := range session.dataChan {
		if err := stream.Send(&pb.CommonViewRawData{Data: data}); err != nil {
			return rpc.GenerateError(codes.Canceled, err)
		}
	}
	return nil
}

func (tsa *TimeSourceApp) PullStationData(
	stream pb.CommonViewDataService_PullStationDataServer) error {
	var machineID string
	var err error
	for {
		if machineID == "" {
			if machineID, err = rpc.GetMachineID(stream.Context()); err != nil {
				return rpc.GenerateArgumentError("machine id")
			}
		}
		rpcData, err := stream.Recv()
		if err != nil {
			if strings.Contains(err.Error(), "EOF") {
				logrus.WithField("prefix", "app.cv").Infof(
					"session by machine-id [%s] closed", machineID)
				return nil
			}
			return rpc.GenerateError(codes.Canceled, err)
		}
		line := string(rpcData.Data)
		if len(line) < 10 {
			logrus.WithField("prefix", "service.cv").
				Warnf("failed to parse user common view data [%s]", line)
			continue
		}
		data := line[1 : len(line)-1]
		lineData := strings.Split(data, ",")
		if err := tsa.process(rpcData.Mode.String(), machineID, rpcData.Data, lineData); err != nil {
			logrus.WithField("prefix", "service.cv").
				Errorf("failed to save user common view data [%v]", err)
		}
	}
}
