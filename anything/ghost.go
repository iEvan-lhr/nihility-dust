package anything

import "fmt"

func ErrorExit(err error) {
	if err != nil {
		//panic(err.Error())
		fmt.Println(err.Error())
	}
}
