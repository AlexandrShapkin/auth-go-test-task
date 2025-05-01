package config

type Config struct {
	JWT      JWT      `mapstracture:"jwt"`
	Database Database `mapstracture:"database"`
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
