package internal

import "time"

type Config struct {
	LogLevel     string        `yaml:"LogLevel"`
	ClientId     string        `yaml:"ClientId"`
	ClientSecret string        `yaml:"ClientSecret"`
	Token        string        `yaml:"Token"`
	RefreshToken string        `yaml:"RefreshToken"`
	TokenExpiry  string        `yaml:"TokenExpiry"`
	Sources      []Source      `yaml:"Sources"`
	TestMode     bool          `yaml:"TestMode"`
	Rotate180    bool          `yaml:"Rotate180"`
	TimeWindow   time.Duration `yaml:"TimeWindow"`
}

type Source struct {
	StationName string   `yaml:"StationName"`
	ModuleNames []string `yaml:"ModuleNames"`
}
