package arg

import (
	"flag"
	"os"
	"reflect"
	"strconv"
	"time"
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

	register(arg, name, env, value)
}

func BindIntVar(arg *int, name, env string, value int, usage string) {
	flag.IntVar(arg, name, value, usage)

	register(arg, name, env, value)
}

func BindBoolVar(arg *bool, name, env string, value bool, usage string) {
	flag.BoolVar(arg, name, value, usage)

	register(arg, name, env, value)
}

func BindFloat64Var(arg *float64, name, env string, value float64, usage string) {
	flag.Float64Var(arg, name, value, usage)

	register(arg, name, env, value)
}

func BindDurationVar(arg *time.Duration, name, env string, value time.Duration, usage string) {
	flag.DurationVar(arg, name, value, usage)

	register(arg, name, env, value)
}

func BindInt64Var(arg *int64, name, env string, value int64, usage string) {
	flag.Int64Var(arg, name, value, usage)

	register(arg, name, env, value)
}

func BindUintVar(arg *uint, name, env string, value uint, usage string) {
	flag.UintVar(arg, name, value, usage)
	register(arg, name, env, value)
}

func BindUint64Var(arg *uint64, name, env string, value uint64, usage string) {
	flag.Uint64Var(arg, name, value, usage)
	register(arg, name, env, value)

}

func register(arg interface{}, name, env string, value interface{}) {
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
			ef, err := strconv.ParseFloat(ev, 64)
			if err != nil {
				panic(err)
			}
			ptr.Elem().Set(reflect.ValueOf(ef))

		case reflect.Int64:
			if _, ok := ptr.Elem().Interface().(time.Duration); ok {
				duration, err := time.ParseDuration(ev)
				if err != nil {
					panic(err)
				}
				ptr.Elem().Set(reflect.ValueOf(duration))
				return
			}
			ei, err := strconv.ParseInt(ev, 10, 64)
			if err != nil {
				panic(err)
			}
			ptr.Elem().Set(reflect.ValueOf(ei))

		case reflect.Uint:
			eu, err := strconv.ParseUint(ev, 10, 32)
			if err != nil {
				panic(err)
			}
			ptr.Elem().Set(reflect.ValueOf(uint(eu)))

		case reflect.Uint64:
			eu, err := strconv.ParseUint(ev, 10, 64)
			if err != nil {
				panic(err)
			}
			ptr.Elem().Set(reflect.ValueOf(eu))
		}
	}
}

func Parse() {
	flag.Parse()

	for key := range bound.flagEnvMapper {
		parseArg(key)
	}
}

func Parsed() bool {
	return flag.Parsed()
}
