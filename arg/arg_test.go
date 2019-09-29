package arg

import (
	"flag"
	"fmt"
	"log"
	"os"
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	var val string
	BindStringVar(&val, "value", "TestParse", "foobar", "test")
	Parse()

	if val != "foobar" {
		t.FailNow()
	}
}

func TestReadFromEnv(t *testing.T) {
	var arg string
	BindStringVar(&arg, "arg", "TestReadFromEnv", "foobar", "test")
	err := os.Setenv("TestReadFromEnv", "peace_and_love")

	Parse()
	if err != nil {
		log.Fatal(err)
	}

	if arg != "peace_and_love" {
		log.Println("value = " + arg)
		t.FailNow()
	}
}

func TestReadFromFlag(t *testing.T) {
	var arg string
	BindStringVar(&arg, "arg", "TestReadFromFlag", "foobar", "test")
	flag.Set("arg", "peace")
	Parse()

	fmt.Println(arg)
}

func TestReadFromFlagAndEnv(t *testing.T) {
	var arg int
	BindIntVar(&arg, "arg", "TestReadFromFlagAndEnv", 111, "test")
	flag.Set("arg", "fromFlag")
	os.Setenv("TestReadFromFlagAndEnv", "222")
	Parse()

	if arg != 222 {
		t.Fatal(arg)
	}
}

func TestReadFromFlagAndEnv_Bool(t *testing.T) {
	var arg bool
	BindBoolVar(&arg, "arg", "TestReadFromFlagAndEnv", false, "test")
	flag.Set("arg", "false")
	os.Setenv("TestReadFromFlagAndEnv", "true")
	Parse()

	if !arg {
		t.Fatal(arg)
	}
}

func TestReadFromFlagAndEnv_Float64(t *testing.T) {
	var arg float64
	BindFloat64Var(&arg, "arg", "TestReadFromFlagAndEnv", 3.2, "test")
	flag.Set("arg", "3.3")
	os.Setenv("TestReadFromFlagAndEnv", "3.4")
	Parse()

	if arg != 3.4 {
		t.Fatal(arg)
	}
}

func TestReadFromFlagAndEnv_Float64_Mode0(t *testing.T) {
	var arg float64
	BindFloat64Var(&arg, "arg", "TestReadFromFlagAndEnv", 3.2, "test")
	flag.Set("arg", "3.3")
	os.Setenv("TestReadFromFlagAndEnv", "3.4")
	EnvFirst(false)
	Parse()

	if arg != 3.3 {
		t.Fatal(arg)
	}
}

func TestReadFromFlagAndEnv_Duration(t *testing.T) {
	var arg time.Duration
	BindDurationVar(&arg, "arg", "TestReadFromFlagAndEnv", 1*time.Minute, "test")
	flag.Set("arg", "2")
	os.Setenv("TestReadFromFlagAndEnv", "3h")
	Parse()

	fmt.Println(arg.String())
}

func TestReadFromFlagAndEnv_Int64(t *testing.T) {
	var arg int64
	BindInt64Var(&arg, "arg", "TestReadFromFlagAndEnv", 0, "test")
	flag.Set("arg", "1")
	os.Setenv("TestReadFromFlagAndEnv", "2")
	Parse()

	fmt.Println(arg)
}

func TestConvertFlagName(t *testing.T) {
	SetEnvPrefix("foo")
	p := convertFlagNameToEnvName("project_id", "_")
	fmt.Println(p)
}
