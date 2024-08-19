package config

import (
	"flag"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env    string `yaml:"env" env-default:"local"`
	Secret string `yaml:"secret" env-required:"true"`

	Server    ServerConfig    `yaml:"server"`
	Tarantool TarantoolConfig `yaml:"tarantool"`
}

type ServerConfig struct {
	Host    string        `yaml:"host" env-default:"localhost"`
	Port    int           `yaml:"port" env-required:"true"`
	Timeout time.Duration `yaml:"timeout" env-default:"10s"`
}

type TarantoolConfig struct {
	Host string `yaml:"host" env-default:"localhost"`
	Port int    `yaml:"port" env-required:"true"`

	User string `yaml:"user" env-required:"true"`
	Pass string `yaml:"pass" env-required:"true"`

	Timeout time.Duration `yaml:"timeout" env-default:"10s"`
}

func MustLoad() *Config {
	path := fetchConfigPath()
	if path == "" {
		panic("config path is empty")
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		panic("config file does not exists: " + err.Error())
	}

	cfg := &Config{}
	if err := cleanenv.ReadConfig(path, cfg); err != nil {
		panic("failed to read config: " + err.Error())
	}

	return cfg
}

func fetchConfigPath() string {
	var res string

	flag.StringVar(&res, "config", "", "config path")
	flag.Parse()

	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}

	return res
}
