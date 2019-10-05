package main

import (
	"io/ioutil"
	"os"
	"sync"
	"time"

	toml "github.com/pelletier/go-toml"
	"github.com/woufrous/sunpos"
)

type XY struct {
	X int
	Y int
}

type Config struct {
	Location                 sunpos.Location
	TimeZoneStandardMaridian float64
	TimeAccel                time.Duration

	Skydome      SourceSkyDome
	Color        SourceColor
	FlipPanels   bool
	FlipExcludes []int
	RefreshRate  time.Duration

	RemapPanels map[string]XY
}

var (
	config *Config
	m      sync.Mutex
)

func (c *Config) GetSunCoordinates() *sunpos.SunCoordinates {
	now := time.Now()
	return sunpos.Sunpos(now, c.Location)
}

func getConfig() (*Config, error) {
	m.Lock()
	defer m.Unlock()
	if config != nil {
		return config, nil
	}

	configF, err := os.Open(configFileLocation)
	if err != nil {
		return nil, err
	}

	configR, err := ioutil.ReadAll(configF)
	if err != nil {
		return nil, err
	}

	config = &Config{}
	if err := toml.Unmarshal(configR, config); err != nil {
		return nil, err
	}

	return config, nil
}
