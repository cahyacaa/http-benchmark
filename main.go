package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	_ "io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptrace"
	"sync"
	"time"

	"github.com/montanaflynn/stats"
	"golang.org/x/net/http2"
)

const (
	TIMEOUT_SEC = 300
)

type Options struct {
	NumRequests int
	HttpVersion int
}

var http1Client = &http.Client{
	Timeout: time.Duration(TIMEOUT_SEC) * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxConnsPerHost:     100,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     90 * time.Second,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: false},
	},
}

var http2ClientV2 = &http.Client{
	Timeout: time.Duration(TIMEOUT_SEC) * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxConnsPerHost:     100,
		MaxIdleConnsPerHost: 100,

		IdleConnTimeout: 90 * time.Second,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: false, NextProtos: []string{"h2"}},
	},
}

var http2Client = &http.Client{
	Timeout: time.Duration(TIMEOUT_SEC) * time.Second,
	Transport: &http2.Transport{
		TLSClientConfig:    &tls.Config{InsecureSkipVerify: false},
		DisableCompression: true,
		AllowHTTP:          true,
		MaxReadFrameSize:   262144 * 4, // defaults to 16k
		CountError: func(errType string) {
			println(errType)
		},
		//StrictMaxConcurrentStreams: true,
		//DialTLSContext: func(ctx context.Context, n, a string, _ *tls.Config) (net.Conn, error) {
		//	var d net.Dialer
		//	return d.DialContext(ctx, n, a)
		//},
	},
}

type RequestResult struct {
	TTFB         int
	Status       string
	DurationMs   int64
	ResponseSize int
	RequestErr   error
}

func sendRequest(httpVersion int) RequestResult {
	start := time.Now()
	var ttfb int64
	var requestErr error
	var responseSize int
	var client *http.Client
	var testURL string

	if httpVersion == 1 {
		client = http1Client
		testURL = "https://http1---korlantas-approver-tlmp6dxpfq-et.a.run.app/test-5mb"
	} else if httpVersion == 2 {
		client = http2Client
		testURL = "https://korlantas-approver-tlmp6dxpfq-et.a.run.app/test-5mb"
	} else {
		client = http2ClientV2
	}

	// file sizes
	// 1 = 64k
	// 2 = 256K
	// 3 = 1MB
	// 4 = 5MB
	// 5 = 10MB
	// 6 = 20MB
	// 7 = 48MB
	//
	// tweak the min/max file number to change the min/max response size
	//min := 1
	//max := 100 // exclusive
	//fileNum := rand.Intn(max-min) + min
	//testURL := "http://localhost:4000/test-5mb"
	//testURL := "http://localhost:8000/test-5mb"
	//testURL := "http://localhost:4000/test"
	fmt.Println(testURL)
	req, err := http.NewRequest("GET", testURL, nil)

	if err != nil {
		panic(err)
	}

	trace := getHttpTrace()

	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
	resp, err := client.Do(req)
	if err != nil {
		requestErr = err
	}

	if resp != nil {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("error", err)
		}
		responseSize = len(body)
		resp.Body.Close()
	}

	duration := time.Since(start).Milliseconds()

	result := RequestResult{
		TTFB:         int(ttfb),
		Status:       resp.Status,
		DurationMs:   duration,
		ResponseSize: responseSize,
		RequestErr:   requestErr,
	}

	return result
}

func benchmark(numRequests int, httpVersion int) {
	var wg sync.WaitGroup
	var mutex sync.Mutex
	var failed int
	//var results []RequestResult
	//var ttfbList []float64
	//start := time.Now()

	count := float32(0)
	//numBytes := 0

	start := time.Now()
	for i := 0; i < numRequests; i++ {
		count++
		wg.Add(1)
		go func() {
			defer wg.Done()
			res := sendRequest(httpVersion)
			if res.RequestErr != nil {
				fmt.Println(res.RequestErr)
			}

			if res.Status != "200" {
				fmt.Println(res.Status)
				mutex.Lock()
				failed++
				mutex.Unlock()
			}
			//
			//results = append(results, result)
			//ttfbList = append(ttfbList, float64(result.TTFB))

			//progress := float64(count) / float64(numRequests) * 100.0
			//numBytes += result.ResponseSize
			//numMB := float32(numBytes) / float32(1048576)
			//
			//durationSec := int(time.Since(start).Seconds())
			//reqPerSec := float32(0)
			//bandwidth := float32(0)
			//if durationSec > 0 {
			//	bandwidth = numMB / float32(time.Since(start).Seconds())
			//	reqPerSec = count / float32(durationSec)
			//}
			//fmt.Printf("%.2f, %.2f%%, %.2f r/s, %.2f mb/s, protocol: %d ,%+v\n", count, progress, reqPerSec, bandwidth, httpVersion, result)
		}()
	}

	wg.Wait()
	fmt.Println(time.Since(start), failed)
	findAvg()
	//fmt.Println(
	//	"ttfb",
	//	calcPercentile(ttfbList, 50),
	//	calcPercentile(ttfbList, 90),
	//	calcPercentile(ttfbList, 95),
	//	calcPercentile(ttfbList, 99),
	//	time.Since(start),
	//)
}

func calcPercentile(values []float64, percent float64) float64 {
	result, _ := stats.Percentile(values, percent)
	return result
}

func parseOptions() *Options {
	var numLogs, httpVersion int

	flag.IntVar(&numLogs, "c", 300, "number of requests")
	flag.IntVar(&httpVersion, "http", 1, "HTTP version to use")
	flag.Parse()

	return &Options{
		NumRequests: numLogs,
		HttpVersion: httpVersion,
	}
}

//go run cmd/bench/main.go  -http 2

func main() {
	rand.NewSource(time.Now().UnixNano())

	opts := parseOptions()
	benchmark(opts.NumRequests, opts.HttpVersion)
}
