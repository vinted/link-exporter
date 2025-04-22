package config

import (
	"encoding/json"
	"os"
)

type Cfg struct {
	MonitorPort       string
	HttpListenAddress string
	Interfaces        []string
	HeartBeatInterval string
	HeartBeatTimeout  string
	LogLevel          string
}

var (
	Config *Cfg
)

func Init(configPath string) error {
	content := []byte(`{}`)
	_, err := os.Stat(configPath)
	if !os.IsNotExist(err) {
		content, err = os.ReadFile(configPath)
		if err != nil {
			return err
		}
	}

	if len(content) == 0 {
		content = []byte(`{}`)
	}

	err = json.Unmarshal(content, &Config)
	if err != nil {
		return err
	}
	return nil
}
