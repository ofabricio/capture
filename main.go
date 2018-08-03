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

type transport struct {
	http.RoundTripper
	itemInfoPath string
	maxItems     int
	currItemID   int
}

var captures Captures

var dashboardSocket socketio.Socket

func main() {
	args := ParseArgs()

	proxyHost := fmt.Sprintf("http://localhost:%s", args.proxyPort)
	dashboardPath := fmt.Sprintf("/%s/", args.dashboard)
	dashboardItemInfoPath := fmt.Sprintf("/%s/items/", args.dashboard)

	transp := &transport{
		RoundTripper: http.DefaultTransport,
		itemInfoPath: dashboardItemInfoPath,
		maxItems:     args.maxCaptures,
		currItemID:   0,
	}

	http.Handle("/", getProxyHandler(args.targetURL, transp))
	http.Handle("/socket.io/", getDashboardSocketHandler())
	http.Handle(dashboardPath, getDashboardHandler())
	http.Handle(dashboardItemInfoPath, getDashboardItemInfoHandler())

	fmt.Printf("\nListening on %s", proxyHost)
	fmt.Printf("\n             %s/%s\n\n", proxyHost, args.dashboard)

	fmt.Println(http.ListenAndServe(":"+args.proxyPort, nil))
}

func getDashboardSocketHandler() http.Handler {
	server, err := socketio.NewServer(nil)
	if err != nil {
		fmt.Println("socket server error", err)
	}
	server.On("connection", func(so socketio.Socket) {
		dashboardSocket = so
		dashboardSocket.Emit("captures", captures.MetadataOnly())
	})
	server.On("error", func(so socketio.Socket, err error) {
		fmt.Println("socket error", err)
	})
	return server
}

func getDashboardHandler() http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.Header().Add("Content-Type", "text/html")
		res.Write([]byte(dashboardHTML))
	})
}

func getDashboardItemInfoHandler() http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		id := req.URL.Path[strings.LastIndex(req.URL.Path, "/")+1:]
		i, _ := strconv.Atoi(id)
		json, _ := json.Marshal(captures[i])
		res.Header().Add("Content-Type", "application/json")
		res.Write([]byte(json))
	})
}

func getProxyHandler(url *url.URL, transp *transport) http.Handler {
	proxy := httputil.NewSingleHostReverseProxy(url)
	proxy.Transport = transp
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		request.Host = request.URL.Host
		proxy.ServeHTTP(response, request)
	})
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {

	reqDump, err := dumpRequest(req)
	if err != nil {
		return nil, err
	}

	res, err := t.RoundTripper.RoundTrip(req)
	if err != nil {
		return nil, fmt.Errorf("uh oh | %v | %s", err, req.URL)
	}

	resDump, err := dumpResponse(res)
	if err != nil {
		return nil, err
	}

	capture := Capture{
		ID:       t.NewItemID(),
		Path:     req.URL.Path,
		Method:   req.Method,
		Status:   res.StatusCode,
		InfoPath: t.itemInfoPath,
		Request:  string(reqDump),
		Response: string(resDump),
	}

	captures.Add(capture)
	captures.RemoveLastAfterReaching(t.maxItems)
	if dashboardSocket != nil {
		dashboardSocket.Emit("captures", captures.MetadataOnly())
	}
	return res, nil
}

func (t *transport) NewItemID() int {
	t.currItemID++
	return t.currItemID
}

func dumpRequest(req *http.Request) ([]byte, error) {
	return httputil.DumpRequest(req, true)
}

func dumpResponse(res *http.Response) ([]byte, error) {
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
