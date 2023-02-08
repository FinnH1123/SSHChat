package config

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v3"
)

type Configuration struct {
	Server   ServerConfiguration
	Database DatabaseConfiguration
}

type DatabaseConfiguration struct {
	Driver       string
	Dbname       string
	Username     string
	Password     string
	Host         string
	Port         string
	MaxLifetime  int
	MaxOpenConns int
	MaxIdleConns int
}

var Config *Configuration

type ServerConfiguration struct {
	Port string
	Host string
	Key  string
}

func Setup() {
	configuration := &Configuration{}

	yamlFile, err := ioutil.ReadFile("data/config.yml")
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, configuration)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	Config = configuration
}

// GetConfig helps you to get configuration data
func GetConfig() *Configuration {
	return Config
}
