package dust

import (
	"bufio"
	"os"
	"reflect"
	"strings"
)

// Scanner 扫描器
type Scanner struct {
	scan map[string]reflect.Value
}


//Dust 灰尘 最小的一份子
type Dust struct {
	Describe string
}

var Dusts string
var sc Scanner

//初始化------------------------------
func init() {
	sc= Scanner{
		scan: make(map[string]reflect.Value, 0),
	}
	valueDust = reflect.ValueOf(&Dust{})
	sc.scan["文件读取"] = valueDust.MethodByName("ReadFile")
	LoadAllDust(sc)
}

// ReadFile 初始化添加的第一个方法 读取文件数据
func (d *Dust) ReadFile(file interface{}) interface{} {
	open, _ := os.Open(file.(string))
	newScanner := bufio.NewScanner(open)
	var s []string
	for newScanner.Scan() {
		s = append(s, newScanner.Text())
	}
	return s
}

var valueDust reflect.Value

// LoadAllDust 读取在本地文件中初始化的方法 建议不要加入太多方法 以免造成程序启动速度过慢
// 本地文件的添加模式为 [调用名称]:[方法名]
func LoadAllDust(scanner Scanner) {
	for _, v := range Building("文件读取","dust.txt").([]string) {
		des:=strings.Split(v,":")[1]
		dow:=strings.Split(v,":")[0]
		scanner.scan[dow]=valueDust.MethodByName(des)
	}
}
// Building 通过Building调用请求 name是用到的方法描述 values是参数列表
func   Building(name string,values ...interface{}) interface{} {
	var vas []reflect.Value
	for _, value := range values {
		vas=append(vas,reflect.ValueOf(value))
	}
	return sc.scan[name].Call(vas)[0].Interface()
}

// AddDust 手动添加方法 主要方法接受器要是 dust包下Dust
func AddDust(name,fun string) {
	if _,ok:=sc.scan[name];ok{
		println(name,":方法已经存在，请检查！")
	}else {
		sc.scan[name]=valueDust.MethodByName(fun)
		println("添加方法",name,"成功！")
	}
}