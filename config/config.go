package config

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/spf13/pflag"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Database struct {
	Driver            string
	DbName            string
	Id                string
	Password          string
	IpAddress         string
	Port              int
	MaxOpenConnection int
	MaxIdleConnection int
}

type LogConfig struct {
	Enable   bool
	StdOut   bool
	Level    string
	Filename string
	MaxSize  int
	MaxAge   int
}

type Echo struct {
	Port int
}

type Api struct {
	HandleTimeoutMS      int
	SellerUploadFilePath string
}

type Config struct {
	Log       LogConfig `toml:"log"`
	Database  Database  `toml:"database"`
	Echo      Echo      `toml:"echo"`
	Api       Api       `toml:"api"`
	LogConfig lumberjack.Logger
}

var conf *Config

func Get() *Config {
	if conf != nil {
		return conf
	}
	conf = &Config{}
	var configPath string
	pflag.StringVarP(&configPath, "config", "c", "config.toml", "config file path")
	pflag.Parse()

	if _, err := toml.DecodeFile(configPath, &conf); err != nil {
		panic(fmt.Sprintf("failed to decode file: %s", err.Error()))
	}

	conf.LogConfig = lumberjack.Logger{
		Filename: conf.Log.Filename,
		MaxSize:  conf.Log.MaxSize,
		MaxAge:   conf.Log.MaxAge,
		Compress: true,
	}
	fmt.Printf("configuration: %+v\n", conf)
	return conf
}
