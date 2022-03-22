package anything

import "log"

func ErrorExit(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func ErrorDontExit(err error) {
	if err != nil {
		log.Println(err.Error())
	}
}
