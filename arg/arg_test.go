package arg

import (
	"flag"
	"fmt"
	"log"
	"os"
	"testing"
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
