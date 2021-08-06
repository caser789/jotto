package common

import (
	"encoding/xml"
	"io/ioutil"

	"git.garena.com/caser789/jotto/jotto"
)

type Config struct {
	MottoSettings *jotto.Settings

	LogLevel int
	Country  string
}

func (c *Config) Motto() *jotto.Settings {
	return c.MottoSettings
}

func LoadCfg(file string) (cfg *Config) {
	content, err := ioutil.ReadFile(file)

	if err != nil {
		return
	}

	cfg = &Config{}
	xml.Unmarshal(content, cfg)

	return
}

func Cfg(app jotto.Application) *Config {
	return app.Settings().(*Config)
}
