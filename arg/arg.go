package arg

import (
	"flag"
	"os"
	"reflect"
	"strconv"
	"strings"
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

	envPrefix = ""

	// DefaultSplitFlag
	DefaultSplitFlag = "-"
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

func BindStringVarWithSmartEnvName(arg *string, name, value, usage string) string {
	env := convertFlagNameToEnvName(name, DefaultSplitFlag)
	BindStringVar(arg, name, env, value, usage)
	return env
}

func BindIntVar(arg *int, name, env string, value int, usage string) {
	flag.IntVar(arg, name, value, usage)

	register(arg, name, env, value)
}

func BindIntVarWithSmartEnvName(arg *int, name string, value int, usage string) string {
	env := convertFlagNameToEnvName(name, DefaultSplitFlag)
	BindIntVar(arg, name, env, value, usage)
	return env
}

func BindBoolVar(arg *bool, name, env string, value bool, usage string) {
	flag.BoolVar(arg, name, value, usage)

	register(arg, name, env, value)
}

func BindBoolVarWithSmartEnvName(arg *bool, name string, value bool, usage string) string {
	env := convertFlagNameToEnvName(name, DefaultSplitFlag)
	BindBoolVar(arg, name, env, value, usage)
	return env
}

func BindFloat64Var(arg *float64, name, env string, value float64, usage string) {
	flag.Float64Var(arg, name, value, usage)

	register(arg, name, env, value)
}

func BindFloat64VarWithSmartEnvName(arg *float64, name string, value float64, usage string) string {
	env := convertFlagNameToEnvName(name, DefaultSplitFlag)
	BindFloat64Var(arg, name, env, value, usage)
	return env
}

func BindDurationVar(arg *time.Duration, name, env string, value time.Duration, usage string) {
	flag.DurationVar(arg, name, value, usage)

	register(arg, name, env, value)
}

func BindDurationVarWithSmartEnvName(arg *time.Duration, name string, value time.Duration, usage string) string {
	env := convertFlagNameToEnvName(name, DefaultSplitFlag)
	BindDurationVar(arg, name, env, value, usage)
	return env
}

func BindInt64Var(arg *int64, name, env string, value int64, usage string) {
	flag.Int64Var(arg, name, value, usage)

	register(arg, name, env, value)
}

func BindInt64VarWithSmartEnvName(arg *int64, name string, value int64, usage string) string {
	env := convertFlagNameToEnvName(name, DefaultSplitFlag)
	BindInt64Var(arg, name, env, value, usage)
	return env
}

func BindUintVar(arg *uint, name, env string, value uint, usage string) {
	flag.UintVar(arg, name, value, usage)
	register(arg, name, env, value)
}

func BindUintVarWithSmartEnvName(arg *uint, name string, value uint, usage string) string {
	env := convertFlagNameToEnvName(name, DefaultSplitFlag)
	BindUintVar(arg, name, env, value, usage)
	return env
}

func BindUint64Var(arg *uint64, name, env string, value uint64, usage string) {
	flag.Uint64Var(arg, name, value, usage)
	register(arg, name, env, value)

}

func BindUint64VarWithSmartEnvName(arg *uint64, name string, value uint64, usage string) string {
	env := convertFlagNameToEnvName(name, DefaultSplitFlag)
	BindUint64Var(arg, name, env, value, usage)
	return env
}

func register(arg interface{}, name, env string, value interface{}) {
	if env == "" {
		panic("disallow blank environment variable name")
	}
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

	var ev string
	ev, readFromEnv = os.LookupEnv(bound.flagEnvMapper[argName])

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

func SetEnvPrefix(pre string) {
	envPrefix = pre
}

func convertFlagNameToEnvName(f, sep string) string {
	var subs []string
	if envPrefix != "" {
		subs = append(subs, strings.ToUpper(envPrefix))
	}
	for _, sub := range strings.Split(f, sep) {
		subs = append(subs, strings.ToUpper(sub))
	}
	return strings.Join(subs, "_")
}
