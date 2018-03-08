package locker

import (
	"fmt"
)

var debug = false

func l(where string, msg interface{}) {
	if debug {
		fmt.Printf("# %s - %v\n", where, msg)
	}
}
