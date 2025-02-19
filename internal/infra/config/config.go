package config

type Config struct {
	DB             DB                   `yaml:"database"`
	RestServer     RestServer           `yaml:"rest_server"`
	UserCreatedPub UserCreatedPublisher `yaml:"user_created_publisher"`
}

type RestServer struct {
	Addr string `yaml:"address"`
}

type DB struct {
	MasterURL string `yaml:"master_url"`
	SlaveURL  string `yaml:"slave_url"`
}

type UserCreatedPublisher struct {
	Addrs []string `yaml:"addresses"`
	Topic string   `yaml:"topic"`
}
