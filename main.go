package main

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
)

type Capture map[string]interface{}

var captures []Capture
var maxCaptures int

type Transport struct {
	http.RoundTripper
}

func main() {
	targetURL, proxyPort, dashboard, maxCaptrs := parseFlags()
	maxCaptures = maxCaptrs

	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	proxy.Transport = Transport{http.DefaultTransport}

	http.Handle("/", getProxyHandler(proxy))
	http.Handle("/socket.io/", getSocketHandler())
	http.Handle("/"+dashboard+"/", getDashboardHandler())

	fmt.Printf("\nListening on http://localhost:%s\n\n", proxyPort)

	http.ListenAndServe(":"+proxyPort, nil)
}

func getProxyHandler(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		request.Host = request.URL.Host
		handler.ServeHTTP(response, request)
	})
}

func (t Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	reqDump, err := httputil.DumpRequest(req, true)
	if err != nil {
		return nil, err
	}

	res, err := t.RoundTripper.RoundTrip(req)
	if err != nil {
		return nil, errors.New(err.Error() + ": " + req.URL.String())
	}

	resDump, err := DumpResponse(res)
	if err != nil {
		return nil, err
	}

	capture := Capture{
		"url":      req.URL.Path,
		"method":   req.Method,
		"status":   res.StatusCode,
		"request":  string(reqDump),
		"response": string(resDump),
	}

	save(capture)

	return res, nil
}

func DumpResponse(res *http.Response) ([]byte, error) {
	var originalBody bytes.Buffer
	res.Body = ioutil.NopCloser(io.TeeReader(res.Body, &originalBody))
	if res.Header.Get("Content-Encoding") == "gzip" {
		res.Body, _ = gzip.NewReader(res.Body)
	}
	resDump, err := httputil.DumpResponse(res, true)
	res.Body = ioutil.NopCloser(&originalBody)
	return resDump, err
}

func save(capture Capture) {
	captures = append([]Capture{capture}, captures...)
	if len(captures) > maxCaptures {
		captures = captures[:len(captures)-1]
	}
	emit(captures)
}
