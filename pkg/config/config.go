package config

type Config struct {
	App      App      `mapstructure:"app"`
	Mail     Mail     `mapstructure:"mail"`
	JWT      JWT      `mapstracture:"jwt"`
	Database Database `mapstracture:"database"`
}

type App struct {
	Addr                string `mapstructure:"addr"`
	LoginRemoteIPMode   bool   `mapstructure:"lremoteipmode"`
	RefreshRemoteIPMode bool   `mapstructure:"rremoteipmode"`
	Domain              string `mapstructure:"domain"`
}

type Mail struct {
	From string `mapstructure:"from"`
	Pass string `mapstructure:"pass"`
}

type JWT struct {
	AccessSecretKey  string `mapstracture:"accesssec"`
	RefreshSecretKey string `mapstracture:"refreshsec"`
}

type Database struct {
	Host     string `mapstracture:"host"`
	User     string `mapstracture:"user"`
	Password string `mapstracture:"password"`
	DBName   string `mapstracture:"dbname"`
	Port     uint   `mapstracture:"port"`
	SSLMode  string `mapstracture:"sslmode"`
}
