package main

import (
	"fmt"
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

type Config struct {
	sync.RWMutex

	AppID     string
	SecureKey string

	Hosts []string `yaml:"hosts"`
}

var config *Config

func init() {
	if config != nil {
		return
	}
	config = new(Config)
	config.Lock()
	defer config.Unlock()

	if v := os.Getenv(LabelGodaddyAppID); v != "" {
		config.Lock()
		defer config.Unlock()
		config.AppID = v
	}

	if v := os.Getenv(LabelGodaddySK); v != "" {
		config.SecureKey = v
	}

	fmt.Println("loaded config from " + os.Getenv(LabelEnv) + ".yaml")
	content, err := os.ReadFile("config/" + os.Getenv(LabelEnv) + ".yaml")
	if err != nil {
		panic(err)
	}
	if err := yaml.Unmarshal(content, config); err != nil {
		panic(err)
	}
}
