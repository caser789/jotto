package common

import "git.garena.com/caser789/jotto/jotto"

type Config struct {
	Protocol string
	Address  string
	LogLevel int

	Country string
}

func Cfg(app motto.Application) *Config {
	cfg, ok := app.Get("cfg")

	if !ok {
		return nil
	}

	return cfg.(*Config)
}
