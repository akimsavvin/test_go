package config

type Config struct {
	DB         DB         `yaml:"database"`
	RestServer RestServer `yaml:"rest_server"`
}

type RestServer struct {
	Addr string `yaml:"address"`
}

type DB struct {
	ConnStr string `yaml:"connection_string"`
}
