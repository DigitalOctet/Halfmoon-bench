package main

import (
	"log"
	"os"
	"strconv"
	"time"

	"math/rand"

	"github.com/eniac/Beldi/internal/utils"
	"github.com/eniac/Beldi/pkg/cayonlib"

	"cs.utexas.edu/zjia/faas"
)

const table = "rw"

var nKeys = 10000
var valueSize = 256 // bytes
var value string

var nOps float64
var readRatio float64
var nReads int
var sleepDuration = 5 * time.Millisecond

func init() {
	if nk, err := strconv.Atoi(os.Getenv("NUM_KEYS")); err == nil {
		nKeys = nk
	} else {
		panic("invalid NUM_KEYS")
	}
	if vs, err := strconv.Atoi(os.Getenv("VALUE_SIZE")); err == nil {
		valueSize = vs
	} else {
		panic("invalid VALUE_SIZE")
	}
	if ops, err := strconv.ParseFloat(os.Getenv("NUM_OPS"), 64); err == nil {
		nOps = ops
	} else {
		panic("invalid NUM_OPS")
	}
	rr, err := strconv.ParseFloat(os.Getenv("READ_RATIO"), 64)
	if err != nil || rr < 0 || rr > 1 {
		panic("invalid READ_RATIO")
	} else {
		readRatio = rr
	}
	nReads = int(nOps * readRatio)
	log.Printf("[INFO] nKeys=%d, valueSize=%d, nOps=%d, readRatio=%.2f, nReads=%d", nKeys, valueSize, int(nOps), readRatio, nReads)

	value = utils.RandomString(valueSize)
	rand.Seed(time.Now().UnixNano())
}

func Handler(env *cayonlib.Env) interface{} {
	// nReads := int(nOps * readRatio)
	for i := 0; i < nReads; i++ {
		cayonlib.Read(env, table, strconv.Itoa(rand.Intn(nKeys)))
		time.Sleep(sleepDuration)
	}
	for i := 0; i < int(nOps)-nReads; i++ {
		cayonlib.Write(env, table, strconv.Itoa(rand.Intn(nKeys)), map[string]interface{}{
			"V": value,
		})
		time.Sleep(sleepDuration)
	}
	return nil
}

func main() {
	faas.Serve(cayonlib.CreateFuncHandlerFactory(Handler))
}
