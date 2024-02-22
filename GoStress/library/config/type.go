package config

type ConfMeta struct {
	ServiceName string `mapstructure:"service-name"`
}

type Server struct {
	ConfMeta
	Secret  []byte
	Version string
	Name    string
}

type Service struct {
	ConfMeta
	Name     string
	AddrList []string
	LB       bool `mapstructure:"load-balance"`
}

type MySQLConf struct {
	ConfMeta
	MySQL struct {
		Addr     string
		Database string
		Username string
		Password string
		Charset  string
	}
}

type EsConf struct {
	ConfMeta
	ES struct {
		Host     string
		Port     string
		UserName string
		Password string
	}
}

type RedisConf struct {
	ConfMeta
	Redis struct {
		Addr     string
		Password string
	}
}

type RabbitMQConf struct {
	ConfMeta
	RabbitMQ struct {
		Addr     string
		Username string
		Password string
	}
}

type InfluxDBConf struct {
	ConfMeta
	InfluxDB struct {
		Addr  string
		Org   string
		Token string
	}
}
