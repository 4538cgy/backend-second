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
	Level    string
	Filename string
	MaxSize  int
	MaxAge   int
}

type Config struct {
	Log       LogConfig `toml:"log"`
	LogConfig lumberjack.Logger
	Database  Database `toml:"database"`
}

func Get() *Config {
	var conf Config
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
	return &conf
}
