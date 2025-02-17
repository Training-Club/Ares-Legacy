package config

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"os"
)

type Configuration struct {
	Ares  Ares  `toml:"ares"`
	Gin   Gin   `toml:"gin"`
	Auth  Auth  `toml:"auth"`
	Mongo Mongo `toml:"mongo"`
	Redis Redis `toml:"redis"`
	S3    S3    `toml:"s3"`
}

type Ares struct {
	CreateAdminAccount bool `toml:"createAdminAccount"`
}

type Gin struct {
	Mode string `toml:"mode"`
	Port string `toml:"port"`
}

type Auth struct {
	AccessTokenPublicKey string `toml:"accessTokenPubKey"`
	AccessTokenTTL       int    `toml:"accessTokenTTL"`

	RefreshTokenPublicKey string `toml:"refreshTokenPubKey"`
	RefreshTokenTTL       int    `toml:"refreshTokenTTL"`
}

type Mongo struct {
	URI string `toml:"uri"`
}

type Redis struct {
	Address  string `toml:"address"`
	Password string `toml:"password"`
}

type S3 struct {
	Key      string `toml:"key"`
	Secret   string `toml:"secret"`
	Token    string `toml:"token"`
	Endpoint string `toml:"endpoint"`
	Region   string `toml:"region"`
	Bucket   string `toml:"bucket"`
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
