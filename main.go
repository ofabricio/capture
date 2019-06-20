package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"plugin"
	"strings"
	"time"

	"github.com/ofabricio/curl"
)

// StatusInternalProxyError is any unknown proxy error
const StatusInternalProxyError = 999

func main() {
	config := ReadConfig()

	proxyURL := fmt.Sprintf("http://localhost:%s", config.ProxyPort)

	fmt.Printf("\nListening on %s", proxyURL)
	fmt.Printf("\n             %s%s\n\n", proxyURL, config.DashboardPath)

	fmt.Println(http.ListenAndServe(":"+config.ProxyPort, NewCaptureHandler(config)))
}

func NewCaptureHandler(config Config) http.Handler {

	srv := NewCaptureService(config.MaxCaptures)

	handler := NewRecorderHandler(srv, NewPluginHandler(NewProxyHandler(config.TargetURL)))

	router := http.NewServeMux()
	router.HandleFunc(config.DashboardPath, NewDashboardHTMLHandler(config))
	router.HandleFunc(config.DashboardConnPath, NewDashboardConnHandler(srv))
	router.HandleFunc(config.DashboardInfoPath, NewDashboardInfoHandler(srv))
	router.HandleFunc(config.DashboardClearPath, NewDashboardClearHandler(srv))
	router.HandleFunc(config.DashboardRetryPath, NewDashboardRetryHandler(srv, handler))
	router.HandleFunc("/", handler)
	return router
}

// NewDashboardConnHandler opens an event stream connection with the dashboard
// so that it is notified everytime a new capture arrives
func NewDashboardConnHandler(srv *CaptureService) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		if _, ok := rw.(http.Flusher); !ok {
			fmt.Printf("streaming not supported at %s\n", req.URL)
			http.Error(rw, "streaming not supported", http.StatusInternalServerError)
			return
		}
		rw.Header().Set("Content-Type", "text/event-stream")
		rw.Header().Set("Cache-Control", "no-cache")
		for {
			jsn, _ := json.Marshal(srv.DashboardItems())
			fmt.Fprintf(rw, "event: captures\ndata: %s\n\n", jsn)
			rw.(http.Flusher).Flush()

			select {
			case <-srv.Updated():
			case <-req.Context().Done():
				return
			}
		}
	}
}

// NewDashboardClearHandler clears all the captures
func NewDashboardClearHandler(srv *CaptureService) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		srv.RemoveAll()
		rw.WriteHeader(http.StatusOK)
	}
}

// NewDashboardHTMLHandler returns the dashboard html page
func NewDashboardHTMLHandler(config Config) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Add("Content-Type", "text/html")
		t, err := template.New("dashboard template").Delims("<<", ">>").Parse(dashboardHTML)
		if err != nil {
			msg := fmt.Sprintf("could not parse dashboard html template: %v", err)
			fmt.Println(msg)
			http.Error(rw, msg, http.StatusInternalServerError)
			return
		}
		t.Execute(rw, config)
	}
}

// NewDashboardRetryHandler retries a request
func NewDashboardRetryHandler(srv *CaptureService, next http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		id := req.URL.Path[strings.LastIndex(req.URL.Path, "/")+1:]
		capture := srv.Find(id)
		var reqBody []byte
		capture.Req.Body, reqBody = drain(capture.Req.Body)
		r, _ := http.NewRequest(capture.Req.Method, capture.Req.URL.String(), bytes.NewReader(reqBody))
		r.Header = capture.Req.Header
		next.ServeHTTP(rw, r)
	}
}

// NewDashboardInfoHandler returns the full capture info
func NewDashboardInfoHandler(srv *CaptureService) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		id := req.URL.Path[strings.LastIndex(req.URL.Path, "/")+1:]
		capture := srv.Find(id)
		rw.Header().Add("Content-Type", "application/json")
		json.NewEncoder(rw).Encode(dump(capture))
	}
}

// NewPluginHandler loads plugin files in the current directory. They are loaded sorted by filename.
func NewPluginHandler(next http.HandlerFunc) http.HandlerFunc {
	ex, err := os.Executable()
	if err != nil {
		fmt.Println("error: could not get executable:", err)
		return next
	}
	exPath := filepath.Dir(ex)
	files, err := ioutil.ReadDir(exPath)
	if err != nil {
		fmt.Println("error: could not read directory:", err)
		return next
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if strings.HasSuffix(file.Name(), ".so") {
			fmt.Printf("loading plugin '%s'\n", file.Name())
			p, err := plugin.Open(exPath + "/" + file.Name())
			if err != nil {
				fmt.Println("error: could not open plugin:", err)
				os.Exit(1)
			}
			fn, err := p.Lookup("Handler")
			if err != nil {
				fmt.Println("error: could not find plugin Handler function:", err)
				os.Exit(1)
			}
			pluginHandler, ok := fn.(func(http.HandlerFunc) http.HandlerFunc)
			if !ok {
				fmt.Println("error: plugin Handler function should be 'func(http.HandlerFunc) http.HandlerFunc'")
				os.Exit(1)
			}
			next = pluginHandler(next)
		}
	}
	return next
}

// NewRecorderHandler saves all the traffic data
func NewRecorderHandler(srv *CaptureService, next http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {

		// save req body for later
		var reqBody []byte
		req.Body, reqBody = drain(req.Body)

		rec := httptest.NewRecorder()

		start := time.Now()

		next.ServeHTTP(rec, req)

		elapsed := time.Since(start).Truncate(time.Millisecond) / time.Millisecond

		// respond
		for k, v := range rec.Header() {
			rw.Header()[k] = v
		}
		rw.WriteHeader(rec.Code)
		rw.Write(rec.Body.Bytes())

		// record req and res
		req.Body = ioutil.NopCloser(bytes.NewReader(reqBody))
		res := rec.Result()
		srv.Insert(Capture{Req: req, Res: res, Elapsed: elapsed})
	}
}

// NewProxyHandler is the reverse proxy handler
func NewProxyHandler(URL string) http.HandlerFunc {
	url, _ := url.Parse(URL)
	proxy := httputil.NewSingleHostReverseProxy(url)
	proxy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, err error) {
		fmt.Printf("uh oh | %v | %s %s\n", err, req.Method, req.URL)
		rw.WriteHeader(StatusInternalProxyError)
		fmt.Fprintf(rw, "%v", err)
	}
	return func(rw http.ResponseWriter, req *http.Request) {
		req.Host = url.Host
		req.URL.Host = url.Host
		req.URL.Scheme = url.Scheme
		proxy.ServeHTTP(rw, req)
	}
}

func dump(c *Capture) CaptureDump {
	reqDump, err := dumpRequest(c.Req)
	if err != nil {
		fmt.Printf("could not dump request: %v\n", err)
	}
	resDump, err := dumpResponse(c.Res)
	if err != nil {
		fmt.Printf("could not dump response: %v\n", err)
	}
	strcurl, err := curl.New(c.Req)
	if err != nil {
		fmt.Printf("could not convert request to curl: %v\n", err)
	}
	return CaptureDump{Request: string(reqDump), Response: string(resDump), Curl: strcurl}
}

func dumpRequest(req *http.Request) ([]byte, error) {
	if req.Header.Get("Content-Encoding") == "gzip" {
		return dumpGzipRequest(req)
	}
	return httputil.DumpRequest(req, true)
}

func dumpGzipRequest(req *http.Request) ([]byte, error) {
	var reqBody []byte
	req.Body, reqBody = drain(req.Body)
	reader, _ := gzip.NewReader(bytes.NewReader(reqBody))
	req.Body = ioutil.NopCloser(reader)
	reqDump, err := httputil.DumpRequest(req, true)
	req.Body = ioutil.NopCloser(bytes.NewReader(reqBody))
	return reqDump, err
}

func dumpResponse(res *http.Response) ([]byte, error) {
	if res.StatusCode == StatusInternalProxyError {
		return dumpInternalProxyError(res)
	}
	if res.Header.Get("Content-Encoding") == "gzip" {
		return dumpGzipResponse(res)
	}
	return httputil.DumpResponse(res, true)
}

// Dumps only the body when we have an proxy error.
// This body is set in NewProxyHandler() in proxy.ErrorHandler
func dumpInternalProxyError(res *http.Response) ([]byte, error) {
	var resBody []byte
	res.Body, resBody = drain(res.Body)
	return resBody, nil
}

func dumpGzipResponse(res *http.Response) ([]byte, error) {
	var resBody []byte
	res.Body, resBody = drain(res.Body)
	reader, _ := gzip.NewReader(bytes.NewReader(resBody))
	res.Body = ioutil.NopCloser(reader)
	resDump, err := httputil.DumpResponse(res, true)
	res.Body = ioutil.NopCloser(bytes.NewReader(resBody))
	return resDump, err
}

func drain(b io.ReadCloser) (io.ReadCloser, []byte) {
	all, _ := ioutil.ReadAll(b)
	b.Close()
	return ioutil.NopCloser(bytes.NewReader(all)), all
}
