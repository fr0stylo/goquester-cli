package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"runtime"
	"strings"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

type RequestParams struct {
	Type    string
	Count   int
	Flood   bool
	Body    string
	Urls    string
	Headers map[string]string
}

func parallelRequests(params *RequestParams) {
	finCh := make(chan time.Duration, 1000)
	for i := 0; i < params.Count; i++ {
		go pRequest(params, finCh)
	}

	for i := 0; i < params.Count; {
		select {
		case time := <-finCh:
			fmt.Println(time)
			i++
		}
	}
}

func pRequest(params *RequestParams, ch chan time.Duration) {
	ch <- sRequest(params.Type, params.Urls, params.Body, params.Headers)
}

func syncRequests(params *RequestParams) {
	for i := 0; i < params.Count; i++ {
		start := time.Now()
		body := params.Body

		if body == "" {
			body = RandStringRunes(rand.Int())
		}
		sRequest(params.Type, params.Urls, params.Body, params.Headers)

		elapsed := time.Since(start)

		fmt.Printf("Time: %s\n", elapsed)
	}
}

func sRequest(method string, url string, body string, headers map[string]string) time.Duration {
	req, err := http.NewRequest(method, url, strings.NewReader(body))
	req.Close = true
	req.Header.Set("Connection", "close")

	if err != nil {
		log.Fatal(err)
	}

	tr := &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 5 * time.Second,
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   time.Second * 10,
	}

	start := time.Now()

	res, err := client.Do(req)

	if err != nil {
		fmt.Println(err.Error())
	}

	if res != nil {
		res.Body.Close()
	}
	elapsed := time.Since(start)

	return elapsed
}

func makeRequests(params *RequestParams) {
	if params.Flood {
		parallelRequests(params)
	} else {
		syncRequests(params)
	}
}

func main() {
	runtime.GOMAXPROCS(10)
	http.DefaultClient.Timeout = time.Second * 10

	requestTypePtr := flag.String("method", http.MethodGet, "Pass HTTP request type")
	requestCountPtr := flag.Int("count", 1, "How many requests to do")
	floodPtr := flag.Bool("flood", false, "Fire requests in parallel")
	bodyPtr := flag.String("body", "", "Body to pass")
	urlsPtr := flag.String("url", "", "Url")

	flag.Parse()

	requestType := strings.ToLower(*requestTypePtr)

	switch requestType {
	case "put":
		requestType = http.MethodPut
		break
	case "get":
		requestType = http.MethodGet
		break
	case "post":
		requestType = http.MethodPost
		break
	case "delete":
		requestType = http.MethodDelete
		break
	case "options":
		requestType = http.MethodOptions
		break
	case "patch":
		requestType = http.MethodPatch
		break
	default:
		requestType = http.MethodGet
	}

	params := RequestParams{
		Body:  *bodyPtr,
		Count: *requestCountPtr,
		Flood: *floodPtr,
		Type:  requestType,
		Urls:  *urlsPtr,
	}

	timeStart := time.Now()

	makeRequests(&params)

	elapsed := time.Since(timeStart)
	fmt.Printf("Total elapsed: %s", elapsed)
}
