package config

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"os"
)

type Configuration struct {
	Security Security `toml:"security"`
	Mongo    Mongo    `toml:"mongo"`
	S3       S3       `toml:"s3"`
}

type Security struct {
	JWT string `toml:"jwt"`
}

type Mongo struct {
	URI string `toml:"uri"`
}

type S3 struct {
	Key      string `toml:"key"`
	Secret   string `toml:"secret"`
	Token    string `toml:"token"`
	Endpoint string `toml:"endpoint"`
	Region   string `toml:"region"`
}

func Get() *Configuration {
	f := "config.toml"

	if _, err := os.Stat(f); err != nil {
		f = "example.config.toml"
		fmt.Println("Couldn't find a config.toml file in root directory, using the defaults...")
	}

	var conf Configuration
	_, err := toml.DecodeFile(f, &conf)

	if err != nil {
		panic("Failed to decode config file: " + err.Error())
	}

	return &conf
}
