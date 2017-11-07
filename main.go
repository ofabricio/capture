package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
)

var captures []Capture
var maxCaptures int

func main() {
	args := parseArgs()
	maxCaptures = args.maxCaptures

	URL, _ := url.Parse(args.url)
	proxy := httputil.NewSingleHostReverseProxy(URL)
	http.Handle("/", getProxyHandler(proxy))
	http.Handle("/socket.io/", getSocketHandler())
	http.Handle("/"+args.dashboard+"/", getDashboardHandler())
	http.ListenAndServe(":"+args.port, nil)
}

func getProxyHandler(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		request.Host = request.URL.Host

		var reqBody bytes.Buffer
		request.Body = ioutil.NopCloser(io.TeeReader(request.Body, &reqBody))

		responseWrapper := NewResponseWrapper(response)
		handler.ServeHTTP(responseWrapper, request)

		saveRequestAndResponse(request, &reqBody, responseWrapper)
	})
}

func saveRequestAndResponse(request *http.Request, reqBody io.Reader, response *ResponseWrapper) {
	capture := Capture{}
	capture.Write(request, reqBody, response)

	captures = append([]Capture{capture}, captures...)
	if len(captures) > maxCaptures {
		captures = captures[:len(captures)-1]
	}
	emit(captures)
}
