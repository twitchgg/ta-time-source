package app

type Config struct {
	CVConfig     *CommonViewDeviceConfig
	SRConfig     *SatelliteReceiverConfig
	CounerConfig *MultipleCounterConfig
	CertPath     string
	RPCListener  string
	WSListener   string
}
type CommonViewDeviceConfig struct {
	SerialPath string
}

type SatelliteReceiverConfig struct {
	GPSSerialPath string
	BDSerialPath  string
}

type MultipleCounterConfig struct {
	Endpoint  string
	Community string
}
