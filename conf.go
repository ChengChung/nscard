package main

import (
	"flag"
	"os"

	"github.com/chengchung/nscard/db"
	"github.com/chengchung/nscard/task"
	"github.com/chengchung/nscard/utils"
	"gopkg.in/yaml.v3"
)

var conf_path = flag.String("conf", "./conf/conf.yaml", "Path to the configuration file")

var cfg *Config

type Config struct {
	TaskConfigs task.TaskConfigs  `yaml:"tasks"`
	DBConfig    db.DBConfig       `yaml:"db"`
	AdminConfig utils.AdminConfig `yaml:"admin"`
	HttpConfig  HttpConfig        `yaml:"http"`
}

type HttpConfig struct {
	ListenAddr string `yaml:"listen_addr"`
}

func ParseConfig() {
	file, err := os.Open(*conf_path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	err = yaml.NewDecoder(file).Decode(&cfg)
	if err != nil {
		panic(err)
	}
}
