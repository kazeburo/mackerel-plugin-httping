package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"net"
	"net/http"
	"os"
	"runtime"
	"sort"
	"time"

	flags "github.com/jessevdk/go-flags"
	"go.mercari.io/go-dnscache"
	"go.uber.org/zap"
)

// Version by Makefile
var Version string

type cmdOpts struct {
	URL              string `long:"url" description:"URL to ping" required:"true"`
	Timeout          int    `long:"timeout" default:"5000" description:"timeout millisec per ping"`
	Interval         int    `long:"interval" default:"10" description:"sleep millisec after every ping"`
	Count            int    `long:"count" default:"10" description:"Count Sending ping"`
	KeyPrefix        string `long:"key-prefix" description:"Metric key prefix" required:"true"`
	DisableKeepalive bool   `long:"disable-keepalive" description:"disable keepalive"`
	Version          bool   `short:"v" long:"version" description:"Show version"`
}

func round(f float64) int64 {
	return int64(math.Round(f)) - 1
}

func createReq(opts cmdOpts) (*http.Request, error) {
	return http.NewRequest("GET", opts.URL, nil)
}

func doRequest(req *http.Request, client http.Client) (time.Duration, error) {
	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	oneByte := make([]byte, 1)
	_, err = resp.Body.Read(oneByte)
	if err != nil && err != io.EOF {
		return 0, err
	}
	// Use Start Transfer timing
	elapsed := time.Since(start)
	defer func() {
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
	}()
	return elapsed, nil
}

func getStats(opts cmdOpts) error {
	resolver, _ := dnscache.New(600*time.Second, time.Millisecond*time.Duration(opts.Timeout), zap.NewNop())

	dialer := (&net.Dialer{
		Timeout:   time.Millisecond * time.Duration(opts.Timeout),
		KeepAlive: 30 * time.Second,
	}).DialContext
	client := http.Client{
		Transport: &http.Transport{
			DialContext:           dnscache.DialFunc(resolver, dialer),
			IdleConnTimeout:       30 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			ResponseHeaderTimeout: time.Millisecond * time.Duration(opts.Timeout),
			DisableKeepAlives:     opts.DisableKeepalive,
		},
	}
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	rand.Seed(time.Now().UTC().UnixNano())

	var rtts sort.Float64Slice
	var t float64
	s := float64(0)
	e := float64(0)

	// preflight
	preReq, err := createReq(opts)
	if err != nil {
		errorNow := uint64(time.Now().Unix())
		fmt.Printf("httping.%s_rtt_count.success\t%f\t%d\n", opts.KeyPrefix, 0.0, errorNow)
		fmt.Printf("httping.%s_rtt_count.error\t%f\t%d\n", opts.KeyPrefix, float64(opts.Count), errorNow)
		return err
	}
	_, err = doRequest(preReq, client)
	if err != nil {
		log.Printf("error in preflight: %v", err)
	}

	for i := 0; i < opts.Count; i++ {
		time.Sleep(time.Millisecond * time.Duration(opts.Interval))
		req, _ := createReq(opts)
		elapsed, err := doRequest(req, client)
		if err != nil {
			log.Printf("%v", err)
			e++
			continue
		}

		rttMilliSec := float64(elapsed.Nanoseconds()) / 1000.0 / 1000.0
		rtts = append(rtts, rttMilliSec)
		t += rttMilliSec
		s++
	}

	sort.Sort(rtts)
	now := uint64(time.Now().Unix())
	fmt.Printf("httping.%s_rtt_count.success\t%f\t%d\n", opts.KeyPrefix, s, now)
	fmt.Printf("httping.%s_rtt_count.error\t%f\t%d\n", opts.KeyPrefix, e, now)
	if s > 0 {
		fmt.Printf("httping.%s_rtt_ms.max\t%f\t%d\n", opts.KeyPrefix, rtts[round(s)], now)
		fmt.Printf("httping.%s_rtt_ms.min\t%f\t%d\n", opts.KeyPrefix, rtts[0], now)
		fmt.Printf("httping.%s_rtt_ms.average\t%f\t%d\n", opts.KeyPrefix, t/s, now)
		fmt.Printf("httping.%s_rtt_ms.90_percentile\t%f\t%d\n", opts.KeyPrefix, rtts[round(s*0.90)], now)
	}

	return nil
}

func main() {
	os.Exit(_main())
}

func _main() int {
	opts := cmdOpts{}
	psr := flags.NewParser(&opts, flags.Default)
	_, err := psr.Parse()
	if err != nil {
		return 1
	}
	if opts.Version {
		fmt.Printf(`%s %s
Compiler: %s %s
`,
			os.Args[0],
			Version,
			runtime.Compiler,
			runtime.Version())
		return 0
	}

	err = getStats(opts)
	if err != nil {
		log.Printf("%v", err)
		return 1
	}
	return 0
}
