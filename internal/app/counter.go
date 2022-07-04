package app

import (
	"fmt"

	"github.com/gosnmp/gosnmp"
	"github.com/sirupsen/logrus"
)

func (tsa *TimeSourceApp) _startCounter(errChan chan error) {
	go func() {
		for data := range tsa.mc.ReadData() {
			pdu1 := gosnmp.SnmpPDU{
				Name:  "1.3.6.1.4.1.326.2.5.1.4",
				Type:  gosnmp.Counter64,
				Value: data.NTSC,
			}
			pdu2 := gosnmp.SnmpPDU{
				Name:  "1.3.6.1.4.1.326.2.5.1.5",
				Type:  gosnmp.Counter64,
				Value: data.Clock,
			}
			pdu3 := gosnmp.SnmpPDU{
				Name:  "1.3.6.1.4.1.326.2.5.1.6",
				Type:  gosnmp.Counter64,
				Value: data.GPS,
			}
			pdu4 := gosnmp.SnmpPDU{
				Name:  "1.3.6.1.4.1.326.2.5.1.7",
				Type:  gosnmp.Counter64,
				Value: data.Beidou,
			}
			trap := gosnmp.SnmpTrap{
				Variables: []gosnmp.SnmpPDU{pdu1, pdu2, pdu3, pdu4},
			}
			if _, err := tsa.snmp.SendTrap(trap); err != nil {
				logrus.WithField("prefix", "app.counter").
					Errorf("failed to send snmp trap data: %v", err)
				continue
			}
			logrus.WithField("prefix", "app.counter").Tracef(
				"send snmp data success: %s", data.String())
		}
	}()
	go func() {
		var err error
		if err = tsa.snmp.Connect(); err != nil {
			errChan <- fmt.Errorf("failed to connect snmp trap server: %v", err)
			return
		}
		err = <-tsa.mc.Start()
		errChan <- err
	}()
}
