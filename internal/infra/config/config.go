package config

type Config struct {
	DB             DB                   `yaml:"database"`
	RestServer     RestServer           `yaml:"rest_server"`
	CreateUserCons CreateUserConsumer   `yaml:"create_user_consumer"`
	UserCreatedPub UserCreatedPublisher `yaml:"user_created_publisher"`
}

type RestServer struct {
	Addr string `yaml:"address"`
}

type DB struct {
	MasterURL string `yaml:"master_url"`
	SlaveURL  string `yaml:"slave_url"`
}

type CreateUserConsumer struct {
	Brokers []string `yaml:"brokers"`
	Topic   string   `yaml:"topic"`
	GroupID string   `yaml:"group_id"`
}

type UserCreatedPublisher struct {
	Brokers []string `yaml:"brokers"`
	Topic   string   `yaml:"topic"`
}
