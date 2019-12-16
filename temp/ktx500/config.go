package ktx500

import (
	"github.com/fpawel/gofins/fins"
	"time"
)

type Config struct {
	MaxAttemptsRead int           `yaml:"max_attempts_read" comment:"число попыток получения ответа"`
	TimeoutMS       uint          `yaml:"timeout_ms" comment:"таймаут считывания, мс"`
	Pause           time.Duration `yaml:"pause" comment:"пауза опроса, с"`
	Server          FinsSettings  `yaml:"server" comment:"параметры ссервера omron fins"`
	Client          FinsSettings  `yaml:"client" comment:"параметры клиента omron fins"`
}

type FinsSettings struct {
	IP       string `yaml:"ip" comment:"tcp адрес"`
	Port     int    `yaml:"port" comment:"tcp порт"`
	Network  byte   `yaml:"network" comment:"fins network"`
	Node     byte   `yaml:"node" comment:"fins node"`
	FinsUnit byte   `yaml:"unit" comment:"fins unit"`
}

func (x FinsSettings) Address() fins.Address {
	return fins.NewAddress(x.IP, x.Port, x.Network, x.Node, x.FinsUnit)
}

func (x Config) newClient() (*fins.Client, error) {
	c, err := fins.NewClient(x.Client.Address(), x.Server.Address())
	if err != nil {
		return nil, err
	}
	c.SetTimeoutMs(x.TimeoutMS)
	return c, nil
}

func NewDefaultConfig() Config {
	return Config{
		MaxAttemptsRead: 20,
		Pause:           time.Second * 2,
		TimeoutMS:       1000,
		Server: FinsSettings{
			IP:   "192.168.250.1",
			Port: 9600,
			Node: 1,
		},
		Client: FinsSettings{
			IP:   "192.168.250.3",
			Port: 9600,
			Node: 254,
		},
	}
}
