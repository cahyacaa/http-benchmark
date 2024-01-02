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

func sendRequest(testURL string, client *http.Client) RequestResult {
	start := time.Now()
	var ttfb int64
	var requestErr error
	var responseSize int

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
	var (
		wg        sync.WaitGroup
		mutex     sync.Mutex
		failedReq int
		client    *http.Client
		testURL   string
		totalMB   float32
	)

	count := 0

	if httpVersion == 1 {
		client = http1Client
		testURL = "https://http1---korlantas-approver-tlmp6dxpfq-et.a.run.app/test-1mb"
	} else if httpVersion == 2 {
		client = http2Client
		testURL = "https://http2---korlantas-approver-tlmp6dxpfq-et.a.run.app/test-1mb"
	}

	fmt.Printf("Test will be executed on URL %s\n", testURL)
	fmt.Printf("Protocol %v\n", httpVersion)

	start := time.Now()
	for i := 0; i < numRequests; i++ {
		count++
		wg.Add(1)
		go func() {
			defer wg.Done()
			res := sendRequest(testURL, client)
			if res.RequestErr != nil {
				fmt.Println(res.RequestErr)
			}

			if res.Status != "200 OK" {
				mutex.Lock()
				failedReq++
				mutex.Unlock()
			} else {
				mutex.Lock()
				totalMB += float32(res.ResponseSize) / float32(1048576)
				mutex.Unlock()
			}

		}()
	}

	wg.Wait()
	fmt.Println(totalMB)
	finishTime := time.Since(start)
	fmt.Printf("Time to finish all request %v, and total failed request %v\n", finishTime.Seconds(), failedReq)
	fmt.Printf("Request success / s : %v\n", float64(numRequests-failedReq)/finishTime.Seconds())
	fmt.Printf("Bandwidth : %v\n", float64(totalMB)/finishTime.Seconds())
	findAvg()
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
