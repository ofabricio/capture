package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"
)

//go:embed dashboard.html
var dashHTML []byte

func main() {

	proxyURL := flag.String("url", "https://jsonplaceholder.typicode.com", "Required. Set the url to proxy")
	proxPort := flag.String("port", "9000", "Set the proxy port")
	dashPort := flag.String("dashboard", "9001", "Set the dashboard port")
	flag.Parse()

	URL, err := url.Parse(*proxyURL)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	proxy, mux := NewProxy(URL, make(chan Capture, 8))

	fmt.Printf("\nListening on http://localhost:%s", *proxPort)
	fmt.Printf("\nDashboard on http://localhost:%s", *dashPort)
	fmt.Printf("\n\n")

	go http.ListenAndServe(":"+*dashPort, mux)
	fmt.Println(http.ListenAndServe(":"+*proxPort, http.HandlerFunc(proxy)))
}

func NewProxy(URL *url.URL, captures chan Capture) (http.HandlerFunc, *http.ServeMux) {

	proxy := func(w http.ResponseWriter, r *http.Request) {

		r.Host = URL.Host
		r.URL.Host = URL.Host
		r.URL.Scheme = URL.Scheme

		var reqBody bytes.Buffer
		r.Body = io.NopCloser(io.TeeReader(r.Body, &reqBody))

		res := httptest.NewRecorder()

		start := time.Now()
		httputil.NewSingleHostReverseProxy(URL).ServeHTTP(res, r)
		elapsed := time.Since(start).Truncate(time.Millisecond) / time.Millisecond

		for k, v := range res.Header() {
			w.Header()[k] = v
		}
		w.WriteHeader(res.Code)
		w.Write(res.Body.Bytes())

		r.Body = io.NopCloser(bytes.NewReader(reqBody.Bytes()))

		if res.Header().Get("Content-Encoding") == "gzip" {
			rd, _ := gzip.NewReader(res.Body)
			dt, _ := io.ReadAll(rd)
			rd.Close()
			res.Body = bytes.NewBuffer(dt)
		}

		dumpReq, _ := httputil.DumpRequest(r, true)
		dumpRes, _ := httputil.DumpResponse(res.Result(), true)

		captures <- Capture{
			Verb:    r.Method,
			Path:    r.URL.Path,
			Status:  res.Result().Status,
			Group:   res.Code / 100,
			Req:     string(dumpReq),
			Res:     string(dumpRes),
			Elapsed: elapsed,
			Curl:    curl(r.Method, r.URL.String(), r.Header, reqBody.Bytes()),
		}
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/html")
		w.Write(dashHTML)
	})

	mux.HandleFunc("/retry", func(w http.ResponseWriter, r *http.Request) {
		rr, err := http.ReadRequest(bufio.NewReader(strings.NewReader(r.FormValue("req"))))
		if err != nil {
			fmt.Println("invalid request format:", err)
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		proxy(httptest.NewRecorder(), rr)
		w.WriteHeader(200)
	})

	mux.HandleFunc("/captures", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		for {
			select {
			case c := <-captures:
				jsn, _ := json.Marshal(c)
				fmt.Fprintf(w, "event: captures\ndata: %s\n\n", jsn)
				w.(http.Flusher).Flush()
			case <-r.Context().Done():
				return
			}
		}
	})

	return proxy, mux
}

func curl(method, url string, head http.Header, body []byte) string {
	var b strings.Builder
	// Build cmd.
	fmt.Fprintf(&b, "curl -X %s %s", method, url)
	// Build head.
	for k, v := range head {
		fmt.Fprintf(&b, " \\\n  -H '%s: %s'", k, strings.Join(v, " "))
	}
	// Build body.
	if len(body) > 0 {
		fmt.Fprintf(&b, " \\\n  -d '%s'", body)
	}
	return b.String()
}

type Capture struct {
	Verb    string
	Path    string
	Status  string
	Group   int
	Req     string
	Res     string
	Curl    string
	Elapsed time.Duration
}
