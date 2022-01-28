package anything

import "fmt"

func ErrorExit(err error) {
	if err != nil {
		fmt.Println(err.Error())
	}
}

// Everything 一切都源自于根
// Everything comes from the root
type Everything interface {
}
