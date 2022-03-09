package dust

import (
	"fmt"
	"github.com/antchfx/htmlquery"
	"github.com/iEvan-lhr/nihility-dust/anything"
	"golang.org/x/net/html"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strings"
)

/*
如果这个世界上有什么东西是最小的，
	那他就不能使用任何东西
If there is the smallest thing in this world,
	he can’t use anything.
*/

func (d Dust) Empty() {
	//TODO implement me
}

//Dust 灰尘 即使是最小的分子 团结起来也能创造一切
type Dust struct {
	imgMap map[int]string
}

func (d *Dust) SliceString(mission chan *anything.Mission, data []any) {
	fmt.Println(data)
	mission <- &anything.Mission{Name: anything.ExitFunction}
}

func (d *Dust) CheckString(mission chan *anything.Mission, data []any) {
	fmt.Println(data)
	mission <- &anything.Mission{Name: anything.ExitFunction}
}

func (d *Dust) PersistenceUrl(mission chan *anything.Mission, data []any) {
	msg, needUrl := data[0].(string), data[1].(string)
	if msg != "" {
		doc, err := htmlquery.Parse(strings.NewReader(msg))
		anything.ErrorExit(err)
		find := htmlquery.Find(doc, "//a")
		img := htmlquery.Find(doc, "//img")

		//可解耦方法
		mission <- &anything.Mission{Name: "FindAllIMG", Pursuit: []any{img}}

		m := make([]string, 0)
		for i := range find {
			if val := htmlquery.SelectAttr(find[i], "href"); strings.Index(val, needUrl) != -1 {
				if strings.Index(val, "http") == -1 {
					m = append(m, "https:"+val)
				} else {
					m = append(m, val)
				}
			}
		}
	}
	mission <- &anything.Mission{Name: anything.ExitFunction}
}

func (d *Dust) FindAllIMG(mission chan *anything.Mission, a []any) {
	urls := a[0].([]*html.Node)
	for i := range urls {
		if val := htmlquery.SelectAttr(urls[i], "src"); d.checkIsImg(val) {
			if strings.Index(val, "http") == -1 {
				d.saveIMG("https:" + val)
			} else {
				d.saveIMG(val)
			}
		}
	}
	mission <- &anything.Mission{Name: anything.ExitFunction}
}

func (d *Dust) checkIsImg(str string) bool {
	if strings.Index(str, "jpg") != -1 {
		return true
	}
	if strings.Index(str, "png") != -1 {
		return true
	}
	if strings.Index(str, "jpeg") != -1 {
		return true
	}
	if strings.Index(str, "gif") != -1 {
		return true
	}

	return false
}

func (d *Dust) saveIMG(url string) {
	resp, err := http.Get(url)
	defer func() {
		err = resp.Body.Close()
		if err != nil {
			return
		}
	}()
	if err != nil {
		fmt.Println(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	name := d.cutStringByName(url)
	if _, ok := d.imgMap[d.getOnlyInt(name)]; !ok {
		if len(d.imgMap) == 0 {
			d.imgMap = map[int]string{d.getOnlyInt(name): "ok"}
		} else {
			d.imgMap[d.getOnlyInt(name)] = "ok"
		}
		_ = ioutil.WriteFile("downloadRES/"+name, body, 0755)
	}
}

func (d *Dust) cutStringByName(s string) string {
	is := strings.Split(s, "/")
	if len(is) > 0 {
		return is[len(is)-1]
	} else {
		return s
	}
}

func (d *Dust) getOnlyInt(s string) int {
	i := 0
	for k := range s {
		i += int(s[k])
	}
	return i
}

func (d *Dust) StartMission(mission chan *anything.Mission, a []any) {
	open, err := os.Open(a[0].(string))
	anything.ErrorExit(err)
	all, err := ioutil.ReadAll(open)
	anything.ErrorExit(err)
	mission <- &anything.Mission{Name: "PersistenceUrl", Pursuit: []any{string(all), "KKKKKKKKKKKKK"}}
	mission <- &anything.Mission{Name: anything.ExitFunction}
}

func (d *Dust) CheckIsBig(mission chan *anything.Mission, a []any) {
	x, y := rand.Intn(20), rand.Intn(20)
	mission <- &anything.Mission{Name: "CountXY", Pursuit: []any{x, y}}
}

func (d *Dust) AllNumber(mission chan *anything.Mission, a []any) {
	if a[0].(int) == 23 {
		mission <- &anything.Mission{Name: anything.ExitFunction, Pursuit: []any{23}}
	} else {
		mission <- &anything.Mission{Name: anything.DC, Pursuit: []any{a[0].(int)}}
		mission <- &anything.Mission{Name: "CheckIsBig", Pursuit: []any{}}
	}

}
