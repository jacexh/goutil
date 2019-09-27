package arg

import (
	"flag"
	"os"
	"reflect"
	"strconv"
)

var (
	bound = struct {
		flagValueMapper map[string]reflect.Value
		flagEnvMapper   map[string]string
		flagPtrMapper   map[string]reflect.Value
	}{
		flagValueMapper: make(map[string]reflect.Value),
		flagEnvMapper:   make(map[string]string),
		flagPtrMapper:   make(map[string]reflect.Value),
	}

	mode = 1 // 0: flag优先；1 ENV优先
)

func EnvFirst(b bool) {
	if b {
		mode = 1
	} else {
		mode = 0
	}
}

func BindStringVar(arg *string, name, env, value, usage string) {
	flag.StringVar(arg, name, value, usage)

	bound.flagValueMapper[name] = reflect.ValueOf(value)
	bound.flagEnvMapper[name] = env
	bound.flagPtrMapper[name] = reflect.ValueOf(arg)
}

func BindIntVar(arg *int, name, env string, value int, usage string) {
	flag.IntVar(arg, name, value, usage)

	bound.flagValueMapper[name] = reflect.ValueOf(value)
	bound.flagEnvMapper[name] = env
	bound.flagPtrMapper[name] = reflect.ValueOf(arg)
}

func BindBoolVar(arg *bool, name, env string, value bool, usage string) {
	flag.BoolVar(arg, name, value, usage)

	bound.flagValueMapper[name] = reflect.ValueOf(value)
	bound.flagEnvMapper[name] = env
	bound.flagPtrMapper[name] = reflect.ValueOf(arg)
}

func BindFloatVar(arg *float64, name, env string, value float64, usage string) {
	flag.Float64Var(arg, name, value, usage)

	bound.flagValueMapper[name] = reflect.ValueOf(value)
	bound.flagEnvMapper[name] = env
	bound.flagPtrMapper[name] = reflect.ValueOf(arg)
}

func parseArg(argName string) {
	var readFromFlag bool
	var readFromEnv bool

	dv := bound.flagValueMapper[argName]
	ptr := bound.flagPtrMapper[argName]

	if ptr.Elem() != dv { // 与默认值一致，会认为未传入值，只能如此
		readFromFlag = true
	}

	ev := os.Getenv(bound.flagEnvMapper[argName])
	if ev != "" {
		readFromEnv = true
	}

	if (readFromFlag && readFromEnv && mode == 1) || (readFromEnv && !readFromFlag) { // 同时传入，根据优先级赋值
		switch dv.Kind() {
		case reflect.Int: // env读取值默认为string类型，需要转换
			ei, err := strconv.Atoi(ev)
			if err != nil {
				panic(err)
			}
			ptr.Elem().Set(reflect.ValueOf(ei))

		case reflect.String:
			ptr.Elem().Set(reflect.ValueOf(ev))

		case reflect.Bool:
			eb, err := strconv.ParseBool(ev)
			if err != nil {
				panic(err)
			}
			ptr.Elem().Set(reflect.ValueOf(eb))

		case reflect.Float64:
			ef, err := strconv.ParseFloat(ev, 10)
			if err != nil {
				panic(err)
			}
			ptr.Elem().Set(reflect.ValueOf(ef))
		}
	}
}

func Parse() {
	flag.Parse()

	for key := range bound.flagEnvMapper {
		parseArg(key)
	}
}
