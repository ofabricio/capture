package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Test the reverse proxy handler
func TestProxyHandler(t *testing.T) {
	// given
	tt := []TestCase{
		GetRequest(),
		PostRequest(),
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			service := httptest.NewServer(http.HandlerFunc(tc.service))
			capture := httptest.NewServer(NewProxyHandler(service.URL))

			// when
			resp := tc.request(capture.URL)

			// then
			tc.test(t, resp)

			resp.Body.Close()
			capture.Close()
			service.Close()
		})
	}
}

type TestCase struct {
	name    string
	request func(string) *http.Response
	service func(http.ResponseWriter, *http.Request)
	test    func(*testing.T, *http.Response)
}

func GetRequest() TestCase {
	msg := "hello"
	return TestCase{
		name: "GetRequest",
		request: func(url string) *http.Response {
			res, _ := http.Get(url)
			return res
		},
		service: func(rw http.ResponseWriter, req *http.Request) {
			fmt.Fprint(rw, string(msg))
		},
		test: func(t *testing.T, res *http.Response) {
			body, _ := ioutil.ReadAll(res.Body)
			if string(body) != msg {
				t.Error("Wrong Body Response")
			}
		},
	}
}

func PostRequest() TestCase {
	msg := "hello"
	return TestCase{
		name: "PostRequest",
		request: func(url string) *http.Response {
			res, _ := http.Post(url, "text/plain", strings.NewReader(msg))
			return res
		},
		service: func(rw http.ResponseWriter, req *http.Request) {
			io.Copy(rw, req.Body)
		},
		test: func(t *testing.T, res *http.Response) {
			body, _ := ioutil.ReadAll(res.Body)
			if string(body) != msg {
				t.Error("Wrong Body Response")
			}
		},
	}
}

func TestDashboardRedirect(t *testing.T) {

	// Given.
	req, _ := http.NewRequest(http.MethodGet, "/something/", nil)
	rec := httptest.NewRecorder()

	// When.
	NewDashboardHTMLHandler().ServeHTTP(rec, req)

	// Then.
	if rec.Code != http.StatusTemporaryRedirect {
		t.Errorf("Wrong response code: got %d, want %d", rec.Code, http.StatusTemporaryRedirect)
	}
	if loc := rec.Header().Get("Location"); loc != "/" {
		t.Errorf("Wrong redirect path: got '%s', want '/'", loc)
	}
}

func Example_dump() {
	c := &Capture{
		Req: Req{
			Proto:  "HTTP/1.1",
			Url:    "http://localhost/hello",
			Path:   "/hello",
			Method: "GET",
			Header: map[string][]string{"Content-Encoding": {"none"}},
			Body:   []byte(`hello`),
		},
		Res: Res{
			Proto:  "HTTP/1.1",
			Header: map[string][]string{"Content-Encoding": {"gzip"}},
			Body:   gzipStr("gziped hello"),
			Status: "200 OK",
		},
	}
	got := dump(c)

	fmt.Println(got.Request)
	fmt.Println(got.Response)
	fmt.Println(got.Curl)

	// Output:
	// GET /hello HTTP/1.1
	//
	// Content-Encoding: none
	//
	// hello
	// HTTP/1.1 200 OK
	//
	// Content-Encoding: gzip
	//
	// gziped hello
	// curl -X GET http://localhost/hello \
	//   -H 'Content-Encoding: none' \
	//   -d 'hello'
}

func gzipStr(str string) []byte {
	var buff bytes.Buffer
	g := gzip.NewWriter(&buff)
	io.WriteString(g, str)
	g.Close()
	return buff.Bytes()
}
