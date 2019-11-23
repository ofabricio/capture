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
	"path"
	"path/filepath"
	"plugin"
	"sort"
	"strings"
	"time"
)

// StatusInternalProxyError is any unknown proxy error
const StatusInternalProxyError = 999

func main() {
	config := ReadConfig()

	fmt.Printf("\nListening on http://localhost:%s", config.ProxyPort)
	fmt.Printf("\nDashboard on http://localhost:%s", config.DashboardPort)
	fmt.Println()

	srv := NewCaptureService(config.MaxCaptures)
	hdr := NewRecorderHandler(srv, NewPluginHandler(NewProxyHandler(config.TargetURL)))

	go func() {
		fmt.Println(http.ListenAndServe(":"+config.DashboardPort, NewDashboardHandler(hdr, srv, config)))
		os.Exit(1)
	}()
	fmt.Println(http.ListenAndServe(":"+config.ProxyPort, hdr))
}

func NewDashboardHandler(h http.HandlerFunc, srv *CaptureService, config Config) http.Handler {
	router := http.NewServeMux()
	router.HandleFunc("/", NewDashboardHTMLHandler(config))
	router.HandleFunc("/conn/", NewDashboardConnHandler(srv))
	router.HandleFunc("/info/", NewDashboardInfoHandler(srv))
	router.HandleFunc("/clear/", NewDashboardClearHandler(srv))
	router.HandleFunc("/retry/", NewDashboardRetryHandler(srv, h))
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
		id := path.Base(req.URL.Path)
		capture := srv.Find(id)

		// creates a new request based on the current one
		r, _ := http.NewRequest(capture.Req.Method, capture.Req.Url, bytes.NewReader(capture.Req.Body))
		r.Header = capture.Req.Header

		next.ServeHTTP(rw, r)
	}
}

// NewDashboardInfoHandler returns the full capture info
func NewDashboardInfoHandler(srv *CaptureService) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		id := path.Base(req.URL.Path)
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
			fmt.Printf("Loading plugin '%s'\n", file.Name())
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

// NewRecorderHandler records all the traffic data
func NewRecorderHandler(srv *CaptureService, next http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {

		// Save req body for later.

		reqBody := &bytes.Buffer{}
		r.Body = ioutil.NopCloser(io.TeeReader(r.Body, reqBody))

		rec := httptest.NewRecorder()

		// Record Roundtrip.

		start := time.Now()

		next.ServeHTTP(rec, r)

		elapsed := time.Since(start).Truncate(time.Millisecond) / time.Millisecond

		resBody := rec.Body.Bytes()

		// Respond to client with recorded response.

		for k, v := range rec.Header() {
			rw.Header()[k] = v
		}
		rw.WriteHeader(rec.Code)
		rw.Write(resBody)

		// Save req and res data.

		req := Req{
			Proto:  r.Proto,
			Method: r.Method,
			Url:    r.URL.String(),
			Path:   r.URL.Path,
			Header: r.Header,
			Body:   reqBody.Bytes(),
		}
		res := Res{
			Proto:  rec.Result().Proto,
			Status: rec.Result().Status,
			Code:   rec.Code,
			Header: rec.Header(),
			Body:   resBody,
		}
		srv.Insert(Capture{Req: req, Res: res, Elapsed: elapsed})
	}
}

// NewProxyHandler is the reverse proxy handler
func NewProxyHandler(URL string) http.HandlerFunc {
	url, _ := url.Parse(URL)
	proxy := httputil.NewSingleHostReverseProxy(url)
	proxy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, err error) {
		fmt.Printf("Uh oh | %v | %s %s\n", err, req.Method, req.URL)
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

func dump(c *Capture) CaptureInfo {
	req := c.Req
	res := c.Res
	return CaptureInfo{
		Request:  dumpContent(req.Header, req.Body, "%s %s %s\n\n", req.Method, req.Path, req.Proto),
		Response: dumpContent(res.Header, res.Body, "%s %s\n\n", res.Proto, res.Status),
		Curl:     dumpCurl(req),
	}
}

func dumpContent(header http.Header, body []byte, format string, args ...interface{}) string {
	b := strings.Builder{}
	fmt.Fprintf(&b, format, args...)
	dumpHeader(&b, header)
	b.WriteString("\n")
	dumpBody(&b, header, body)
	return b.String()
}

func dumpHeader(dst *strings.Builder, header http.Header) {
	var headers []string
	for k, v := range header {
		headers = append(headers, fmt.Sprintf("%s: %s\n", k, strings.Join(v, " ")))
	}
	sort.Strings(headers)
	for _, v := range headers {
		dst.WriteString(v)
	}
}

func dumpBody(dst *strings.Builder, header http.Header, body []byte) {
	reqBody := body
	if header.Get("Content-Encoding") == "gzip" {
		reader, _ := gzip.NewReader(bytes.NewReader(body))
		reqBody, _ = ioutil.ReadAll(reader)
	}
	dst.Write(reqBody)
}

func dumpCurl(req Req) string {
	var b strings.Builder
	// build cmd
	fmt.Fprintf(&b, "curl -X %s %s", req.Method, req.Url)
	// build headers
	for k, v := range req.Header {
		fmt.Fprintf(&b, " \\\n  -H '%s: %s'", k, strings.Join(v, " "))
	}
	// build body
	if len(req.Body) > 0 {
		fmt.Fprintf(&b, " \\\n  -d '%s'", req.Body)
	}
	return b.String()
}
