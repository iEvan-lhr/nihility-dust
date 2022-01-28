package main

import (
	"fmt"
	"github.com/iEvan-lhr/nihility-dust/anything"
	"github.com/iEvan-lhr/nihility-dust/dust"
	"io/ioutil"
	"os"
	"time"
)

func main() {
	start := time.Now()
	w := dust.Wind{}
	w.Register(&dust.Dust{})
	w.Init()
	open, err := os.Open("dust/test.html")
	anything.ErrorExit(err)
	all, err := ioutil.ReadAll(open)
	anything.ErrorExit(err)
	w.Schedule("PersistenceUrl", string(all), "KKKKKKKKKK")
	_, ok := w.A["PersistenceUrl"]
	for !ok {
		_, ok = w.A["PersistenceUrl"]
	}
	fmt.Println(time.Now().Sub(start))
}
