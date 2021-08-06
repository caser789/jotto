package common

import (
	"fmt"

	"github.com/caser789/jotto/jotto"
)

func Boot(app interface{}) {
	a := app.(motto.Application)

	fmt.Println("booted", a)
}
