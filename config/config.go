package config

import "github.com/fstanis/screenresolution"

type Config struct {
	Screen ScreenConfig `yaml:"screen"`
	Time   TimeConfig   `yaml:"time"`
}

type ScreenConfig struct {
	Res *screenresolution.Resolution `yaml:"res"`
}

type TimeConfig struct {
	Now string `yaml:"now"`
}
