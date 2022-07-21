package app

import "strings"

type Config struct {
	CVConfig     *CommonViewDeviceConfig
	SRConfig     *SatelliteReceiverConfig
	CounerConfig *MultipleCounterConfig
	CertPath     string
	RPCListener  string
	WSListener   string
	ESConf       *ElasticSearchConfig
}
type CommonViewDeviceConfig struct {
	SerialPath string
	DevID      string
}

type SatelliteReceiverConfig struct {
	GPSSerialPath string
	BDSerialPath  string
}

type MultipleCounterConfig struct {
	Endpoint  string
	Community string
}

var esIndexMapping = `{
	"settings" : {
		   "index" : {
			   "number_of_shards" : 2, 
			   "number_of_replicas" : 1
		   }
	   }
   }`

// ElasticSearchConfig elasticsearch cluster config
type ElasticSearchConfig struct {
	Endpoints string
	IndexName string
}

func (conf *ElasticSearchConfig) GetEndpoints() []string {
	return strings.Split(conf.Endpoints, ",")
}
