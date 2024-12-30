package connection

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type Configuration struct {
	Server   string `yaml:"server"`
	Port     int    `yaml:"port"`
	Nick     string `yaml:"nick"`
	User     string `yaml:"user"`
	Ssl      bool   `yaml:"ssl"`
	RealName string `yaml:"realName"`
}

// reads from yaml and populates the configuration struct
func NewConfiguration() (*Configuration, error) {
	yamlData, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Fatalf("Error reading yaml file: %s", err)
	}

	c := &Configuration{}
	err = yaml.Unmarshal(yamlData, &c)

	return c, err
}
