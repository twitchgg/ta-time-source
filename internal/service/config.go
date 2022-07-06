package service

type Config struct {
	CertPath        string
	ServerName      string
	ServiceEndpoint string
	CVDataListener  string
}

func (c *Config) Check() error {
	return nil
}
