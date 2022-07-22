package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	"ntsc.ac.cn/ta-time-source/pkg/cv"
)

// CVEsEntity cv es entity
type CVEsEntity struct {
	Mode      string    `json:"@Mode"`
	Tnano     time.Time `json:"@tnano"`
	Producer  string    `json:"@producer"`
	Utcdate   time.Time `json:"@utcdate"`
	Locationx int       `json:"locationx"`
	Locationy int       `json:"locationy"`
	Locationz int       `json:"locationz"`
	Sysdelay  int       `json:"sysdelay"`
	Refdelay  int       `json:"refdelay"`
	Sat       string    `json:"sat"`
	Cl        string    `json:"cl"`
	Mjd       int       `json:"mjd"`
	Sttime    time.Time `json:"@sttime"`
	Trkl      int       `json:"trkl"`
	Elv       int       `json:"elv"`
	Azth      int       `json:"azth"`
	Refsv     int       `json:"refsv"`
	Srsv      int       `json:"srsv"`
	Refsys    int       `json:"refsys"`
	Srsys     int       `json:"srsys"`
	Dsg       int       `json:"dsg"`
	Ioe       int       `json:"ioe"`
	Mdtr      int       `json:"mdtr"`
	Smdt      int       `json:"smdt"`
	Mdio      int       `json:"mdio"`
	Smdi      int       `json:"smdi"`
	Msio      int       `json:"msio"`
	Smsi      int       `json:"smsi"`
	Isg       int       `json:"isg"`
	Fr        int       `json:"fr"`
	Hc        int       `json:"hc"`
	Frc       string    `json:"frc"`
	Ck        string    `json:"ck"`
}

func genESData(raw [][]interface{}, mode string) []*CVEsEntity {
	var entities []*CVEsEntity
	for _, data := range raw {
		e := CVEsEntity{
			Mode:      mode,
			Tnano:     data[0].(time.Time),
			Producer:  data[1].(string),
			Utcdate:   data[2].(time.Time),
			Locationx: data[3].(int),
			Locationy: data[4].(int),
			Locationz: data[5].(int),
			Sysdelay:  data[6].(int),
			Refdelay:  data[7].(int),
			Sat:       data[8].(string),
			Cl:        data[9].(string),
			Mjd:       data[10].(int),
			Sttime:    data[11].(time.Time),
			Trkl:      data[12].(int),
			Elv:       data[13].(int),
			Azth:      data[14].(int),
			Refsv:     data[15].(int),
			Srsv:      data[16].(int),
			Refsys:    data[17].(int),
			Srsys:     data[18].(int),
			Dsg:       data[19].(int),
			Ioe:       data[20].(int),
			Mdtr:      data[21].(int),
			Smdt:      data[22].(int),
			Mdio:      data[23].(int),
			Smdi:      data[24].(int),
			Msio:      data[25].(int),
			Smsi:      data[26].(int),
			Isg:       data[27].(int),
			Fr:        data[28].(int),
			Hc:        data[29].(int),
			Frc:       data[30].(string),
			Ck:        data[31].(string),
		}
		entities = append(entities, &e)
	}
	return entities
}

func (tsa *TimeSourceApp) insertCVData(data []*CVEsEntity) error {
	for _, e := range data {
		jsonData, err := json.Marshal(e)
		if err != nil {
			return err
		}
		resp, err := tsa.esCluster.Index(
			tsa.conf.ESConf.IndexName, bytes.NewReader(jsonData))
		if err != nil {
			return fmt.Errorf("push data failed: %s", err.Error())
		}
		defer resp.Body.Close()
		if resp.StatusCode != 201 {
			fmt.Println(resp)
			return fmt.Errorf("push data failed: %s", resp.Status())
		}
	}
	return nil
}

func (tsa *TimeSourceApp) process(mode string, machineID string, raw []byte, data []string) error {
	if len(data) == 0 {
		return fmt.Errorf("no common view data")
	}
	if len(data) < cv.DefaultCVHeaderSize {
		return fmt.Errorf("common view data length [%d]", len(data))
	}
	num := data[6]
	satelliteNum, err := strconv.Atoi(num)
	if err != nil {
		return fmt.Errorf("query common view satellite num failed [%s]", err.Error())
	}
	if satelliteNum == 0 {
		if len(data) != cv.DefaultCVHeaderSize {
			return fmt.Errorf(
				"statellite number is 0,query common view data length failed [%d]", len(data))
		}
	}
	dataLen := len(data) - cv.DefaultCVHeaderSize
	expectedLen := satelliteNum * cv.DefaultSatelliteDataSize
	if dataLen != expectedLen {
		return fmt.Errorf(
			"statellite number is [%d],expected payload length [%d],actual length [%d]",
			satelliteNum, expectedLen, dataLen)
	}
	tnano := time.Now()
	layout := "20060102T15:04:05Z07:00"
	utcDate, err := time.Parse(layout, data[0]+"T00:00:00Z")
	if err != nil {
		return err
	}
	header := data[1:6]
	var headerData []interface{}
	for _, hv := range header {
		av, err := strconv.Atoi(hv)
		if err != nil {
			return err
		}
		headerData = append(headerData, av)
	}
	idx := cv.DefaultCVHeaderSize
	var writeDate string
	var rawData [][]interface{}
	for i := 0; i < satelliteNum; i++ {
		asData := data[idx : idx+cv.DefaultSatelliteDataSize]
		idx = idx + cv.DefaultSatelliteDataSize
		var adata []interface{}
		adata = append(adata, tnano)
		if machineID == "" {
			adata = append(adata, tsa.conf.CVConfig.DevID)
		} else {
			adata = append(adata, machineID)
		}
		adata = append(adata, utcDate)
		adata = append(adata, headerData...)
		for ai, av := range asData {
			if ai == 3 {
				slayout := "20060102T150405Z07:00"
				ds := fmt.Sprintf("%sT%sZ", data[0], av)
				sdate, err := time.Parse(slayout, ds)
				if err != nil {
					return err
				}
				adata = append(adata, sdate)
				if writeDate == "" {
					writeDate = sdate.Local().Format(time.RFC3339)
				}
				continue
			}
			if ai == 0 || ai == 1 || ai == 22 || ai == 23 {
				adata = append(adata, av)
			} else {
				iv, err := strconv.Atoi(av)
				if err != nil {
					return err
				}
				adata = append(adata, iv)
			}
		}
		rawData = append(rawData, adata)
	}
	entities := genESData(rawData, mode)
	if err := tsa.insertCVData(entities); err != nil {
		return err
	}
	logrus.WithField("prefix", "cv.service.handler").
		Debugf("machine [%s] [%d] satellite cv data success,time [%s]",
			machineID, satelliteNum, writeDate)
	return nil
}
