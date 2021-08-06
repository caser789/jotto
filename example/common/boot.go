package common

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"

	"github.com/caser789/jotto/jotto"
)

func Boot(app interface{}) {
	a := app.(motto.Application)

	content, err := ioutil.ReadFile("conf/conf.xml")

	if err != nil {
		return
	}

	cfg := &Config{}
	xml.Unmarshal(content, cfg)

	a.Set("cfg", cfg)
	fmt.Println("booted", a)
}
