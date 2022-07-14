package app

import (
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

func (tsa *TimeSourceApp) PushMainStationData(req *pb.PushRequest,
	stream pb.CommonViewDataService_PushMainStationDataServer) error {
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
