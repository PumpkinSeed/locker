package locker

import (
	"fmt"
)

var debug = true

func l(where string, msg interface{}) {
	if debug {
		fmt.Printf("%s - %v\n", where, msg)
	}
}
