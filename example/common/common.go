package common

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"

	"git.garena.com/duanzy/motto/motto"
	"git.garena.com/lixh/goorm"
	"github.com/gogo/protobuf/proto"
)

/* Custom configuration struct */

func NewConfiguration(filepath string) *Configuration {
	return &Configuration{
		filepath: filepath,
	}
}

type Configuration struct {
	Settings

	filepath string
}

func (c *Configuration) Motto() *motto.Settings {
	return c.Settings.Motto
}

func (c *Configuration) Load() (err error) {
	content, err := ioutil.ReadFile(c.filepath)

	if err != nil {
		return
	}

	xml.Unmarshal(content, &c.Settings)

	fmt.Println("Application settings loaded")

	return
}

type Settings struct {
	Motto *motto.Settings

	LogLevel int
	Country  string
}

func Cfg(app motto.Application) *Configuration {
	return app.Settings().(*Configuration)
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
