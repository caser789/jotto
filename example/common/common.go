package common

import (
	"encoding/xml"
	"io/ioutil"

	"git.garena.com/caser789/jotto/jotto"
	"git.garena.com/lixh/goorm"
	"github.com/gogo/protobuf/proto"
)

/* Custom configuration struct */

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

/* Custom context struct */

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

/* Custom processor */

func NewProcessor(message, reply proto.Message, handler motto.ProcessorHandler, middlewares []motto.Middleware, orm *OrmSetting) *Processor {

	return &Processor{
		BaseProcessor: *motto.NewProcessor(message, reply, handler, middlewares).(*motto.BaseProcessor),
		Orm:           orm,
	}
}

type Processor struct {
	motto.BaseProcessor
	Orm *OrmSetting
}

type OrmSetting struct {
	Database string
	Flag     int
}
