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
	"plugin"
	"strings"

	"github.com/ofabricio/curl"
)

func main() {
	config := ReadConfig()
	startCapture(config)
}

func startCapture(config Config) {

	list := NewCaptureList(config.MaxCaptures)

	handler := NewPlugin(NewRecorder(list, NewProxyHandler(config.TargetURL)))

	http.Handle("/", handler)
	http.Handle(config.DashboardPath, NewDashboardHtmlHandler(config))
	http.Handle(config.DashboardConnPath, NewDashboardConnHandler(list))
	http.Handle(config.DashboardClearPath, NewDashboardClearHandler(list))
	http.Handle(config.DashboardRetryPath, NewDashboardRetryHandler(list, handler))
	http.Handle(config.DashboardItemInfoPath, NewDashboardItemInfoHandler(list))

	captureHost := fmt.Sprintf("http://localhost:%s", config.ProxyPort)

	fmt.Printf("\nListening on %s", captureHost)
	fmt.Printf("\n             %s/%s\n\n", captureHost, config.Dashboard)

	fmt.Println(http.ListenAndServe(":"+config.ProxyPort, nil))
}

func NewDashboardConnHandler(list *CaptureList) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if _, ok := rw.(http.Flusher); !ok {
			fmt.Printf("streaming not supported at %s\n", req.URL)
			http.Error(rw, "streaming not supported", http.StatusInternalServerError)
			return
		}
		rw.Header().Set("Content-Type", "text/event-stream")
		rw.Header().Set("Cache-Control", "no-cache")

		fmt.Fprintf(rw, "event: connected\ndata: %s\n\n", "clear")
		rw.(http.Flusher).Flush()

		for {
			jsn, _ := json.Marshal(list.ItemsAsMetadata())
			fmt.Fprintf(rw, "event: captures\ndata: %s\n\n", jsn)
			rw.(http.Flusher).Flush()

			select {
			case <-list.Updated:
			case <-req.Context().Done():
				return
			}
		}
	})
}

func NewDashboardClearHandler(list *CaptureList) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		list.RemoveAll()
		rw.WriteHeader(http.StatusOK)
	})
}

func NewDashboardHtmlHandler(config Config) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Add("Content-Type", "text/html")
		t, err := template.New("dashboard template").Delims("<<", ">>").Parse(dashboardHTML)
		if err != nil {
			msg := fmt.Sprintf("could not parse dashboard html template: %v", err)
			fmt.Println(msg)
			http.Error(rw, msg, http.StatusInternalServerError)
			return
		}
		t.Execute(rw, config)
	})
}

func NewDashboardRetryHandler(list *CaptureList, next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		id := req.URL.Path[strings.LastIndex(req.URL.Path, "/")+1:]
		capture := list.Find(id)
		if capture == nil {
			http.Error(rw, "Item Not Found", http.StatusNotFound)
			return
		}
		var reqBody []byte
		capture.Req.Body, reqBody = drain(capture.Req.Body)
		r, _ := http.NewRequest(capture.Req.Method, capture.Req.URL.String(), capture.Req.Body)
		r.Header = capture.Req.Header
		next.ServeHTTP(rw, r)
		capture.Req.Body = ioutil.NopCloser(bytes.NewReader(reqBody))
	})
}

func NewDashboardItemInfoHandler(list *CaptureList) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		id := req.URL.Path[strings.LastIndex(req.URL.Path, "/")+1:]
		capture := list.Find(id)
		if capture == nil {
			http.Error(rw, "Item Not Found", http.StatusNotFound)
			return
		}
		rw.Header().Add("Content-Type", "application/json")
		json.NewEncoder(rw).Encode(dump(capture))
	})
}

func NewPlugin(next http.Handler) http.Handler {
	p, err := plugin.Open("plugin.so")
	if err != nil {
		if strings.HasPrefix(err.Error(), "plugin.Open") {
			fmt.Printf("error: could not open plugin file 'plugin.so': %v\n", err)
		}
		return next
	}
	f, err := p.Lookup("Handler")
	if err != nil {
		fmt.Printf("error: could not find plugin Handler function %v\n", err)
		return next
	}
	pluginFn, ok := f.(func(http.Handler) http.Handler)
	if !ok {
		fmt.Println("error: plugin Handler function should be 'func(http.Handler) http.Handler'")
		return next
	}
	return pluginFn(next)
}

func NewRecorder(list *CaptureList, next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {

		// save req body for later
		var reqBody []byte
		req.Body, reqBody = drain(req.Body)

		rec := httptest.NewRecorder()

		next.ServeHTTP(rec, req)

		// respond
		for k, v := range rec.HeaderMap {
			rw.Header()[k] = v
		}
		rw.WriteHeader(rec.Code)
		rw.Write(rec.Body.Bytes())

		// record req and res
		req.Body = ioutil.NopCloser(bytes.NewReader(reqBody))
		res := rec.Result()
		list.Insert(Capture{Req: req, Res: res})
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
