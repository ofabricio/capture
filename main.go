package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"strconv"
	"strings"

	"github.com/googollee/go-socket.io"
)

type transport struct {
	http.RoundTripper
}

var captureID = 0
var captures Captures

var socket socketio.Socket

var host string
var dashboardPath string
var dashboardItemPath string

var args Args

func main() {
	args = args.Parse()

	dashboardPath = fmt.Sprintf("/%s/", args.dashboard)
	dashboardItemPath = fmt.Sprintf("/%s/items/", args.dashboard)

	http.Handle("/", getProxyHandler())
	http.Handle("/socket.io/", getSocketHandler())
	http.Handle(dashboardItemPath, getDashboardItemHandler())
	http.Handle(dashboardPath, getDashboardHandler())

	host = fmt.Sprintf("http://localhost:%s", args.port)

	fmt.Printf("\nListening on %s", host)
	fmt.Printf("\n             %s/%s\n\n", host, args.dashboard)

	if err := http.ListenAndServe(":"+args.port, nil); err != nil {
		fmt.Println(err)
	}
}

func getProxyHandler() http.Handler {
	proxy := httputil.NewSingleHostReverseProxy(args.url)
	proxy.Transport = transport{http.DefaultTransport}
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		request.Host = request.URL.Host
		proxy.ServeHTTP(response, request)
	})
}

func getSocketHandler() http.Handler {
	server, err := socketio.NewServer(nil)
	if err != nil {
		panic(err)
	}
	server.On("connection", func(so socketio.Socket) {
		socket = so
		emit()
	})
	server.On("error", func(so socketio.Socket, err error) {
		fmt.Println("socket error:", err)
	})
	return server
}

func getDashboardItemHandler() http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		id := strings.TrimPrefix(req.URL.Path, dashboardItemPath)
		i, _ := strconv.Atoi(id)
		json, _ := json.Marshal(captures[i])
		res.Header().Add("Content-Type", "application/json")
		res.Write([]byte(json))
	})
}

func getDashboardHandler() http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.Header().Add("Content-Type", "text/html")
		res.Write([]byte(dashboardHTML))
	})
}

func (t transport) RoundTrip(req *http.Request) (*http.Response, error) {

	reqDump, err := DumpRequest(req)
	if err != nil {
		return nil, err
	}

	res, err := t.RoundTripper.RoundTrip(req)
	if err != nil {
		return nil, fmt.Errorf("uh oh | %v | %s", err, req.URL)
	}

	resDump, err := DumpResponse(res)
	if err != nil {
		return nil, err
	}

	capture := Capture{
		ID:       captureID,
		Path:     req.URL.Path,
		Method:   req.Method,
		Status:   res.StatusCode,
		Request:  string(reqDump),
		Response: string(resDump),
	}
	captureID++

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
	socket.Emit("captures", captures.ToReferences(host+dashboardItemPath))
}
