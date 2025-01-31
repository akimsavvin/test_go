package config

type Config struct {
	DB         DB         `yaml:"database"`
	RestServer RestServer `yaml:"rest_server"`
}

type RestServer struct {
	Addr string `yaml:"address"`
}

type DB struct {
	MasterURL string `yaml:"master_url"`
	SlaveURL  string `yaml:"slave_url"`
}
