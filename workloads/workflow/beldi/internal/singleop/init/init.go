package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/eniac/Beldi/internal/utils"
	"github.com/eniac/Beldi/pkg/beldilib"
)

var table = "singleop"
var baseline = false

var nKeys = 10000
var valueSize = 256 // bytes
var value string

func init() {
	if nk, err := strconv.Atoi(os.Getenv("NUM_KEYS")); err == nil {
		nKeys = nk
	} else {
		panic("invalid NUM_KEYS")
	}
	if beldilib.TYPE == "BASELINE" {
		// table = "b" + table
		baseline = true
	}
	if vs, err := strconv.Atoi(os.Getenv("VALUE_SIZE")); err == nil {
		valueSize = vs
	} else {
		panic("invalid VALUE_SIZE")
	}
	value = utils.RandomString(valueSize)
}

func clean() {
	if baseline {
		beldilib.DeleteTable("b" + table)
		beldilib.WaitUntilDeleted("b" + table)
	} else {
		beldilib.DeleteLambdaTables(table)
		beldilib.WaitUntilDeleted(table)
		beldilib.WaitUntilDeleted(fmt.Sprintf("%s-log", table))
		beldilib.WaitUntilDeleted(fmt.Sprintf("%s-collector", table))
	}
}

func create() {
	if baseline {
		beldilib.CreateBaselineTable("b" + table)
		time.Sleep(10 * time.Second)
		beldilib.WaitUntilActive("b" + table)
	} else {
		for _, lambda := range []string{table, "nop"} {
			beldilib.CreateLambdaTables(lambda)
			time.Sleep(10 * time.Second)
			beldilib.WaitUntilActive(lambda)
			beldilib.WaitUntilActive(fmt.Sprintf("%s-log", lambda))
			beldilib.WaitUntilActive(fmt.Sprintf("%s-collector", lambda))
		}
	}
}

func populate() {
	for i := 0; i < nKeys; i++ {
		beldilib.Populate(table, strconv.Itoa(i), value, baseline)
	}
}

func main() {
	option := os.Args[1]
	if option == "clean" {
		clean()
	} else if option == "create" {
		create()
	} else if option == "populate" {
		populate()
	} else {
		panic("unkown option: " + option)
	}
}
