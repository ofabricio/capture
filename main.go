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
	"net/url"
	"strconv"
	"strings"

	"github.com/googollee/go-socket.io"
)

var captures Captures

var dashboardSocket socketio.Socket

func main() {
	config := ReadConfig()
	startCapture(config)
}

func startCapture(config Config) {
	http.Handle("/", proxyHandler(config))
	http.Handle("/socket.io/", dashboardSocketHandler(config))
	http.Handle(config.DashboardPath, dashboardHandler())
	http.Handle(config.DashboardClearPath, dashboardClearHandler())
	http.Handle(config.DashboardItemInfoPath, dashboardItemInfoHandler())

	captureHost := fmt.Sprintf("http://localhost:%s", config.ProxyPort)

	fmt.Printf("\nListening on %s", captureHost)
	fmt.Printf("\n             %s/%s\n\n", captureHost, config.Dashboard)

	fmt.Println(http.ListenAndServe(":"+config.ProxyPort, nil))
}

func dashboardSocketHandler(config Config) http.Handler {
	server, err := socketio.NewServer(nil)
	if err != nil {
		fmt.Printf("socket server error: %v\n", err)
	}
	server.On("connection", func(so socketio.Socket) {
		dashboardSocket = so
		dashboardSocket.Emit("config", config)
		emitToDashboard(captures)
	})
	server.On("error", func(so socketio.Socket, err error) {
		fmt.Printf("socket error: %v\n", err)
	})
	return server
}

func dashboardClearHandler() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		captures = nil
		emitToDashboard(captures)
		rw.WriteHeader(http.StatusOK)
	})
}

func dashboardHandler() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Add("Content-Type", "text/html")
		fmt.Fprint(rw, dashboardHTML)
	})
}

func dashboardItemInfoHandler() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		idStr := req.URL.Path[strings.LastIndex(req.URL.Path, "/")+1:]
		idInt, _ := strconv.Atoi(idStr)
		for _, c := range captures {
			if c.ID == idInt {
				rw.Header().Add("Content-Type", "application/json")
				json.NewEncoder(rw).Encode(c)
				break
			}
		}
	})
}

func proxyHandler(config Config) http.Handler {
	url, _ := url.Parse(config.TargetURL)
	captureID := 0
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		req.Host = url.Host
		req.URL.Host = url.Host
		req.URL.Scheme = url.Scheme

		reqDump, err := dumpRequest(req)
		if err != nil {
			fmt.Printf("could not dump request: %v\n", err)
		}

		proxy := httputil.NewSingleHostReverseProxy(url)
		proxy.ModifyResponse = func(res *http.Response) error {
			resDump, err := dumpResponse(res)
			if err != nil {
				return fmt.Errorf("could not dump response: %v", err)
			}
			captureID++
			capture := Capture{
				ID:       captureID,
				Path:     req.URL.Path,
				Method:   req.Method,
				Status:   res.StatusCode,
				Request:  string(reqDump),
				Response: string(resDump),
			}
			captures.Add(capture)
			captures.RemoveLastAfterReaching(config.MaxCaptures)
			emitToDashboard(captures)
			return nil
		}
		proxy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, err error) {
			fmt.Printf("uh oh | %v | %s\n", err, req.URL)
		}
		proxy.ServeHTTP(rw, req)
	})
}

func dumpRequest(req *http.Request) ([]byte, error) {
	if req.Header.Get("Content-Encoding") == "gzip" {
		var originalBody bytes.Buffer
		tee := io.TeeReader(req.Body, &originalBody)
		reader, _ := gzip.NewReader(tee)
		req.Body = ioutil.NopCloser(reader)
		reqDump, err := httputil.DumpRequest(req, true)
		req.Body = ioutil.NopCloser(&originalBody)
		return reqDump, err
	}
	return httputil.DumpRequest(req, true)
}

func dumpResponse(res *http.Response) ([]byte, error) {
	if res.Header.Get("Content-Encoding") == "gzip" {
		var originalBody bytes.Buffer
		tee := io.TeeReader(res.Body, &originalBody)
		reader, _ := gzip.NewReader(tee)
		res.Body = ioutil.NopCloser(reader)
		resDump, err := httputil.DumpResponse(res, true)
		res.Body = ioutil.NopCloser(&originalBody)
		return resDump, err
	}
	return httputil.DumpResponse(res, true)
}

func emitToDashboard(captures Captures) {
	if dashboardSocket != nil {
		dashboardSocket.Emit("captures", captures.MetadataOnly())
	}
}
