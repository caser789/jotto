package common

import (
	"encoding/xml"
	"io/ioutil"

	"git.garena.com/caser789/jotto/jotto"
	"git.garena.com/lixh/goorm"
)

type Config struct {
	MottoSettings *motto.Settings

	LogLevel int
	Country  string
}

func (c *Config) Motto() *motto.Settings {
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

func Cfg(app motto.Application) *Config {
	return app.Settings().(*Config)
}

type Context struct {
	MottoCtx *motto.BaseContext
	Orm      *goorm.Orm
}

func (ctx *Context) Motto() *motto.BaseContext {
	return ctx.MottoCtx
}

func Ctx(ctx motto.Context) *Context {
	return ctx.(*Context)
}
