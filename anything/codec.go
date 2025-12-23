package anything

import (
	"fmt"
	"reflect"
)

// Interpreter (解释器)
// 负责将 []any 绑定到具体的 Go 结构体上
type Interpreter struct {
	data     []any
	pipeline chan *Mission // 注入的管道，用于自动填充 chan *Mission 类型的字段
}

func NewInterpreter(data []any, pipeline chan *Mission) *Interpreter {
	return &Interpreter{
		data:     data,
		pipeline: pipeline,
	}
}

// Bind 将数据绑定到目标结构体指针 ptr
func (i *Interpreter) Bind(ptr any) error {
	val := reflect.ValueOf(ptr)
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("interpreter: target must be a pointer to a struct, got %T", ptr)
	}

	elem := val.Elem()
	typ := elem.Type()

	for j := 0; j < elem.NumField(); j++ {
		fieldVal := elem.Field(j)
		fieldType := typ.Field(j)

		// 1. 自动注入 Pipeline (chan *Mission)
		// 只要字段类型是 chan *Mission，就自动注入，无需标签
		if fieldType.Type == reflect.TypeOf((chan *Mission)(nil)) {
			if fieldVal.CanSet() {
				fieldVal.Set(reflect.ValueOf(i.pipeline))
			}
			continue
		}

		// 2. 解析 dust 标签 (dust:"0")
		tag := fieldType.Tag.Get("dust")
		if tag == "" {
			continue // 没有标签的字段跳过
		}

		var index int
		// 简单的解析逻辑，后续可扩展支持默认值等
		if _, err := fmt.Sscanf(tag, "%d", &index); err != nil {
			continue // 标签格式错误，跳过
		}

		// 3. 数据填充
		if index >= 0 && index < len(i.data) {
			srcVal := reflect.ValueOf(i.data[index])

			// 检查源数据是否有效
			if !srcVal.IsValid() {
				continue
			}

			// 类型兼容性检查与赋值
			if srcVal.Type().AssignableTo(fieldVal.Type()) {
				fieldVal.Set(srcVal)
			} else if srcVal.Type().ConvertibleTo(fieldVal.Type()) {
				// 尝试弱类型转换 (例如 int -> int64)
				fieldVal.Set(srcVal.Convert(fieldVal.Type()))
			} else {
				// 记录类型不匹配警告，或者选择报错
				fmt.Printf("[Interpreter] Warning: Field '%s' expects %v but got %v (index %d)\n",
					fieldType.Name, fieldVal.Type(), srcVal.Type(), index)
			}
		}
	}
	return nil
}

// Assembler (组装器)
// 负责将任意结果集打包回 []any
type Assembler struct{}

func NewAssembler() *Assembler {
	return &Assembler{}
}

// Pack 将反射返回值列表转换为 []any
func (a *Assembler) Pack(results []reflect.Value) []any {
	output := make([]any, 0, len(results))
	for _, v := range results {
		if v.IsValid() && v.CanInterface() {
			output = append(output, v.Interface())
		} else {
			output = append(output, nil)
		}
	}
	return output
}
