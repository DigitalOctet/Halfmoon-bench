package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"time"

	"cs.utexas.edu/zjia/faas-retwis/utils"

	"github.com/montanaflynn/stats"
)

var FLAGS_faas_gateway string
var FLAGS_fn_prefix string

// var FLAGS_num_users int
var FLAGS_concurrency int
var FLAGS_duration int
var FLAGS_prewarm int

// var FLAGS_percentages string

// var FLAGS_bodylen int
var FLAGS_rand_seed int
var FLAGS_zipf_skew float64
var FLAGS_keyspace int
var FLAGS_read_keys int
var FLAGS_rw_ratio float64
var rng *rand.Zipf

func init() {
	flag.StringVar(&FLAGS_faas_gateway, "faas_gateway", "127.0.0.1:8081", "")
	flag.StringVar(&FLAGS_fn_prefix, "fn_prefix", "", "")
	// flag.IntVar(&FLAGS_num_users, "num_users", 1000, "")
	flag.IntVar(&FLAGS_concurrency, "concurrency", 1, "")
	flag.IntVar(&FLAGS_duration, "duration", 10, "")
	// flag.StringVar(&FLAGS_percentages, "percentages", "50,50", "write,read")
	// flag.IntVar(&FLAGS_bodylen, "bodylen", 64, "")
	flag.IntVar(&FLAGS_rand_seed, "rand_seed", 23333, "")
	flag.Float64Var(&FLAGS_zipf_skew, "zipf_skew", 1.1, "")
	flag.IntVar(&FLAGS_read_keys, "read_keys", 8, "")
	flag.Float64Var(&FLAGS_rw_ratio, "rw_ratio", 0.125, "")
	flag.IntVar(&FLAGS_keyspace, "keyspace", 10000, "")
	flag.IntVar(&FLAGS_prewarm, "prewarm", 5, "")
	// rand.Seed(int64(FLAGS_rand_seed))
}

// func parsePercentages(s string) ([]int, error) {
// 	parts := strings.Split(s, ",")
// 	if len(parts) != 2 {
// 		return nil, fmt.Errorf("Need exactly four parts splitted by comma")
// 	}
// 	results := make([]int, 2)
// 	for i, part := range parts {
// 		if parsed, err := strconv.Atoi(part); err != nil {
// 			return nil, fmt.Errorf("Failed to parse %d-th part", i)
// 		} else {
// 			results[i] = parsed
// 		}
// 	}
// 	for i := 1; i < len(results); i++ {
// 		results[i] += results[i-1]
// 	}
// 	if results[len(results)-1] != 100 {
// 		return nil, fmt.Errorf("Sum of all parts is not 100")
// 	}
// 	return results, nil
// }

func buildTestRequest() utils.JSONValue {
	// tags := make([]uint64, 0, FLAGS_read_keys)
	tags := make(map[uint64]struct{})
	nReads := FLAGS_read_keys
	nWrites := int(float64(FLAGS_read_keys) * FLAGS_rw_ratio)
	if nWrites == 0 {
		nWrites = 1
	}
	n := nReads
	if nWrites > nReads {
		n = nWrites
	}
	for i := 0; i < n; i++ {
		for {
			tag := rng.Uint64() + 2
			if _, ok := tags[tag]; !ok {
				tags[tag] = struct{}{}
				break
			}
		}
	}
	keys := []uint64{}
	for k := range tags {
		keys = append(keys, k)
	}

	return utils.JSONValue{"readset": keys[:nReads], "writeset": keys[:nWrites]}
}

const kTxnConflitMsg = "txn aborted"

func printFnResult(fnName string, duration time.Duration, results []*utils.FaasCall) {
	total := 0
	succeeded := 0
	txnConflit := 0
	latencies := make([]float64, 0, 128)
	for _, result := range results {
		if result.FnName == FLAGS_fn_prefix+fnName {
			total++
			if result.Result.Success {
				succeeded++
			} else if result.Result.Message == kTxnConflitMsg {
				txnConflit++
			}
			if result.Result.StatusCode == 200 {
				d := result.Result.Duration
				latencies = append(latencies, float64(d.Microseconds()))
			}
		}
	}
	if total == 0 {
		return
	}
	failed := total - succeeded - txnConflit
	fmt.Printf("[%s]\n", fnName)
	fmt.Printf("Throughput: %.1f requests per sec\n", float64(total)/duration.Seconds())
	if txnConflit > 0 {
		ratio := float64(txnConflit) / float64(txnConflit+succeeded)
		fmt.Printf("Transaction conflits: %d (%.2f%%)\n", txnConflit, ratio*100.0)
	}
	if failed > 0 {
		ratio := float64(failed) / float64(total)
		fmt.Printf("Failed transactions: %d (%.2f%%)\n", failed, ratio*100.0)
	}
	if len(latencies) > 0 {
		median, _ := stats.Median(latencies)
		p99, _ := stats.Percentile(latencies, 99.0)
		fmt.Printf("Latency: median = %.3fms, tail (p99) = %.3fms\n", median/1000.0, p99/1000.0)
	}
}

func main() {
	flag.Parse()

	src := rand.NewSource(int64(FLAGS_rand_seed))
	rnd := rand.New(src)
	rng = rand.NewZipf(rnd, FLAGS_zipf_skew, 1, uint64(FLAGS_keyspace)-1)

	// percentages, err := parsePercentages(FLAGS_percentages)
	// if err != nil {
	// 	log.Fatalf("[FATAL] Invalid \"percentages\" flag: %v", err)
	// }
	client := utils.NewFaasClient(FLAGS_faas_gateway, FLAGS_concurrency)
	log.Printf("[INFO] Start running for %d seconds with concurrency of %d", FLAGS_duration, FLAGS_concurrency)
	startTime := time.Now()
	for {
		elapsed := time.Since(startTime)
		if elapsed > time.Duration(FLAGS_duration)*time.Second {
			break
		}
		preWarm := elapsed < time.Duration(FLAGS_prewarm)*time.Second
		client.AddJsonFnCall(FLAGS_fn_prefix+"RWTxn", buildTestRequest(), preWarm)
		// k := rand.Intn(100)
		// if k < percentages[0] {
		// 	client.AddJsonFnCall(FLAGS_fn_prefix+"TestWrite", buildTestRequest())
		// } else if k < percentages[1] {
		// 	client.AddJsonFnCall(FLAGS_fn_prefix+"TestRead", buildTestRequest())
		// }
	}
	elapsed := time.Since(startTime) - time.Duration(FLAGS_prewarm)*time.Second
	results := client.WaitForResults()
	fmt.Printf("Benchmark runs for %v, %.1f request per sec\n", elapsed, float64(len(results))/elapsed.Seconds())
	printFnResult("RWTxn", elapsed, results)
	// printFnResult("TestWrite", elapsed, results)
	// printFnResult("TestRead", elapsed, results)
}
