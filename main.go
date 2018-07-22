package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"strconv"
	"strings"

	"github.com/googollee/go-socket.io"
)

type Transport struct {
	http.RoundTripper
}

var captures Captures

var socket socketio.Socket

var host string
var dashboardPath string
var dashboardItemsPath string

func main() {
	parseArgs()

	proxy := httputil.NewSingleHostReverseProxy(args.url)
	proxy.Transport = Transport{http.DefaultTransport}

	dashboardPath = "/" + args.dashboard + "/"
	dashboardItemsPath = dashboardPath + "items/"

	http.Handle("/", getProxyHandler(proxy))
	http.Handle("/socket.io/", getSocketHandler())
	http.Handle(dashboardItemsPath, getCapturesHandler())
	http.Handle(dashboardPath, getDashboardHandler())

	host = "http://localhost:" + args.port

	fmt.Printf("\nListening on %s", host)
	fmt.Printf("\n             %s/%s\n\n", host, args.dashboard)

	http.ListenAndServe(":"+args.port, nil)
}

func getCapturesHandler() http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		id := strings.TrimPrefix(req.URL.Path, dashboardItemsPath)
		i, _ := strconv.Atoi(id)
		json, _ := json.Marshal(captures[i])
		res.Header().Add("Content-Type", "application/json")
		res.Write([]byte(json))
	})
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

	reqDump, err := DumpRequest(req)
	if err != nil {
		return nil, err
	}

	res, err := t.RoundTripper.RoundTrip(req)
	if err != nil {
		return nil, errors.New("uh oh | " + err.Error() + " | " + req.URL.String())
	}

	resDump, err := DumpResponse(res)
	if err != nil {
		return nil, err
	}

	capture := Capture{
		Path:     req.URL.Path,
		Method:   req.Method,
		Status:   res.StatusCode,
		Request:  string(reqDump),
		Response: string(resDump),
	}

	captures.Add(capture)
	emit()

	return res, nil
}

func DumpRequest(req *http.Request) ([]byte, error) {
	return httputil.DumpRequest(req, true)
}

func DumpResponse(res *http.Response) ([]byte, error) {
	var originalBody bytes.Buffer
	reader := io.TeeReader(res.Body, &originalBody)
	if res.Header.Get("Content-Encoding") == "gzip" {
		reader, _ = gzip.NewReader(reader)
	}
	res.Body = ioutil.NopCloser(reader)
	resDump, err := httputil.DumpResponse(res, true)
	res.Body = ioutil.NopCloser(&originalBody)
	return resDump, err
}

func emit() {
	if socket == nil {
		return
	}
	socket.Emit("captures", captures.ToReferences(host+dashboardItemsPath))
}
