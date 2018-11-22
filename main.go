package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/googollee/go-socket.io"
)

var dashboardSocket socketio.Socket

func main() {
	config := ReadConfig()
	startCapture(config)
}

func startCapture(config Config) {

	repo := NewCapturesRepository(config.MaxCaptures)

	http.Handle("/", NewRecorder(repo, NewProxyHandler(config.TargetURL)))
	http.Handle(config.DashboardPath, NewDashboardHtmlHandler())
	http.Handle(config.DashboardClearPath, NewDashboardClearHandler(repo))
	http.Handle(config.DashboardItemInfoPath, NewDashboardItemInfoHandler(repo))
	http.Handle(config.DashboardConnPath, NewDashboardSocketHandler(repo, config))

	captureHost := fmt.Sprintf("http://localhost:%s", config.ProxyPort)

	fmt.Printf("\nListening on %s", captureHost)
	fmt.Printf("\n             %s/%s\n\n", captureHost, config.Dashboard)

	fmt.Println(http.ListenAndServe(":"+config.ProxyPort, nil))
}

func NewDashboardSocketHandler(repo CaptureRepository, config Config) http.Handler {
	server, err := socketio.NewServer(nil)
	if err != nil {
		fmt.Printf("socket server error: %v\n", err)
	}
	server.On("connection", func(so socketio.Socket) {
		dashboardSocket = so
		dashboardSocket.Emit("config", config)
		emitToDashboard(repo.FindAll())
	})
	server.On("error", func(so socketio.Socket, err error) {
		fmt.Printf("socket error: %v\n", err)
	})
	return server
}

func NewDashboardClearHandler(repo CaptureRepository) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		repo.RemoveAll()
		emitToDashboard(nil)
		rw.WriteHeader(http.StatusOK)
	})
}

func NewDashboardHtmlHandler() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Add("Content-Type", "text/html")
		fmt.Fprint(rw, dashboardHTML)
	})
}

func NewDashboardItemInfoHandler(repo CaptureRepository) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		id := req.URL.Path[strings.LastIndex(req.URL.Path, "/")+1:]
		capture := repo.Find(id)
		if capture == nil {
			http.Error(rw, "Item Not Found", http.StatusNotFound)
			return
		}
		rw.Header().Add("Content-Type", "application/json")
		json.NewEncoder(rw).Encode(capture)
	})
}

func NewRecorder(repo CaptureRepository, next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		reqDump, err := dumpRequest(req)
		if err != nil {
			fmt.Printf("could not dump request: %v\n", err)
		}

		rec := httptest.NewRecorder()

		next.ServeHTTP(rec, req)

		for k, v := range rec.HeaderMap {
			rw.Header()[k] = v
		}
		rw.WriteHeader(rec.Code)
		rw.Write(rec.Body.Bytes())

		res := rec.Result()
		resDump, err := dumpResponse(res)
		if err != nil {
			fmt.Printf("could not dump response: %v\n", err)
		}
		capture := Capture{
			Path:     req.URL.Path,
			Method:   req.Method,
			Status:   res.StatusCode,
			Request:  string(reqDump),
			Response: string(resDump),
		}
		repo.Insert(capture)
		emitToDashboard(repo.FindAll())
	})
}

func NewProxyHandler(URL string) http.Handler {
	url, _ := url.Parse(URL)
	proxy := httputil.NewSingleHostReverseProxy(url)
	proxy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, err error) {
		fmt.Printf("uh oh | %v | %s %s\n", err, req.Method, req.URL)
	}
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		req.Host = url.Host
		req.URL.Host = url.Host
		req.URL.Scheme = url.Scheme
		proxy.ServeHTTP(rw, req)
	})
}

func dumpRequest(req *http.Request) ([]byte, error) {
	if req.Header.Get("Content-Encoding") == "gzip" {
		var reqBody []byte
		req.Body, reqBody = drain(req.Body)
		reader, _ := gzip.NewReader(bytes.NewReader(reqBody))
		req.Body = ioutil.NopCloser(reader)
		reqDump, err := httputil.DumpRequest(req, true)
		req.Body = ioutil.NopCloser(bytes.NewReader(reqBody))
		return reqDump, err
	}
	return httputil.DumpRequest(req, true)
}

func dumpResponse(res *http.Response) ([]byte, error) {
	if res.Header.Get("Content-Encoding") == "gzip" {
		var resBody []byte
		res.Body, resBody = drain(res.Body)
		reader, _ := gzip.NewReader(bytes.NewReader(resBody))
		res.Body = ioutil.NopCloser(reader)
		resDump, err := httputil.DumpResponse(res, true)
		res.Body = ioutil.NopCloser(bytes.NewReader(resBody))
		return resDump, err
	}
	return httputil.DumpResponse(res, true)
}

func drain(b io.ReadCloser) (io.ReadCloser, []byte) {
	all, _ := ioutil.ReadAll(b)
	b.Close()
	return ioutil.NopCloser(bytes.NewReader(all)), all
}

func emitToDashboard(captures []Capture) {
	if dashboardSocket == nil {
		return
	}
	metadatas := make([]CaptureMetadata, len(captures))
	for i, capture := range captures {
		metadatas[i] = capture.Metadata()
	}
	dashboardSocket.Emit("captures", metadatas)
}
