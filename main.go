package main

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"

	"github.com/googollee/go-socket.io"
)

var captures Captures
var socket socketio.Socket

type Transport struct {
	http.RoundTripper
}

func main() {
	targetURL, proxyPort, dashboard, maxCaptures := parseFlags()
	captures.max = maxCaptures

	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	proxy.Transport = Transport{http.DefaultTransport}

	http.Handle("/", getProxyHandler(proxy))
	http.Handle("/socket.io/", getSocketHandler())
	http.Handle("/"+dashboard+"/", getDashboardHandler())

	fmt.Printf("\nListening on http://localhost:%s", proxyPort)
	fmt.Printf("\n             http://localhost:%s/%s\n\n", proxyPort, dashboard)

	http.ListenAndServe(":"+proxyPort, nil)
}

func getProxyHandler(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		request.Host = request.URL.Host
		handler.ServeHTTP(response, request)
	})
}

func getSocketHandler() http.Handler {
	server, err := socketio.NewServer(nil)
	if err != nil {
		log.Fatal(err)
	}
	server.On("connection", func(so socketio.Socket) {
		socket = so
		emit()
	})
	server.On("error", func(so socketio.Socket, err error) {
		log.Println("socket error:", err)
	})
	return server
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

	capture := Capture{req.URL.Path, req.Method, res.StatusCode,
		string(reqDump),
		string(resDump),
	}

	captures.Add(capture)
	emit()

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

func emit() {
	if socket == nil {
		return
	}
	socket.Emit("captures", captures.items)
}
