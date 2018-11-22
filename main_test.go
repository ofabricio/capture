package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"
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
			capture := httptest.NewServer(proxyHandler(service.URL))

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

func TestDumpRequest(t *testing.T) {
	msg := "hello"

	// given
	req, err := http.NewRequest(http.MethodPost, "http://localhost:9000/", strings.NewReader(msg))
	if err != nil {
		t.Errorf("Could not create request: %v", err)
	}

	// when
	body, err := dumpRequest(req)

	// then
	if err != nil {
		t.Errorf("Dump Request error: %v", err)
	}
	if !strings.Contains(string(body), msg) {
		t.Errorf("Dump Request is not '%s'", msg)
	}
}

func TestDumpRequestGzip(t *testing.T) {
	msg := "hello"

	// given
	req, err := http.NewRequest(http.MethodPost, "http://localhost:9000/", strings.NewReader(gzipStr(msg)))
	req.Header.Set("Content-Encoding", "gzip")
	if err != nil {
		t.Errorf("Could not create request: %v", err)
	}

	// when
	body, err := dumpRequest(req)

	// then
	if err != nil {
		t.Errorf("Dump Request Gzip error: %v", err)
	}
	if !strings.Contains(string(body), msg) {
		t.Errorf("Dump Request Gzip is not '%s'", msg)
	}
}

func TestDumpResponse(t *testing.T) {
	msg := "hello"

	// given
	res := &http.Response{Body: ioutil.NopCloser(strings.NewReader(msg))}

	// when
	body, err := dumpResponse(res)

	// then
	if err != nil {
		t.Errorf("Dump Response Error: %v", err)
	}
	if !strings.Contains(string(body), msg) {
		t.Errorf("Dump Response is not '%s'", msg)
	}
}

func TestDumpResponseGzip(t *testing.T) {
	msg := "hello"

	// given
	h := make(http.Header)
	h.Set("Content-Encoding", "gzip")
	res := &http.Response{Header: h, Body: ioutil.NopCloser(strings.NewReader(gzipStr(msg)))}

	// when
	body, err := dumpResponse(res)

	// then
	if err != nil {
		t.Errorf("Dump Response error: %v", err)
	}
	if !strings.Contains(string(body), msg) {
		t.Error("Not hello")
	}
}

func TestCaptureIDConcurrence(t *testing.T) {

	// This test bothers me

	// given

	interactions := 1000

	service := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
		rw.WriteHeader(http.StatusOK)
	}))
	repo := NewCapturesRepository(interactions)
	capture := httptest.NewServer(NewRecorder(repo, proxyHandler(service.URL)))
	defer service.Close()
	defer capture.Close()

	// when

	// Starts go routines so that captureID is incremented concurrently within proxyHandler()
	wg := &sync.WaitGroup{}
	wg.Add(interactions)
	for i := 0; i < interactions; i++ {
		go func() {
			_, err := http.Get(capture.URL)
			if err != nil {
				t.Fatalf("Request Failed: %v", err)
			}
			wg.Done()
		}()
	}
	wg.Wait()

	// then

	// Tests if captures IDs are sequential
	captures := repo.FindAll()
	if len(captures) == 0 {
		t.Fatalf("No captures found")
	}
	ids := make([]int, len(captures))
	for i := 0; i < len(captures); i++ {
		ids[i] = captures[i].ID
	}
	sort.Ints(ids)
	for i := 0; i < len(captures); i++ {
		if ids[i] != i+1 {
			t.Fatalf("Capture IDs are not sequential")
		}
	}
}

func gzipStr(str string) string {
	var buff bytes.Buffer
	g := gzip.NewWriter(&buff)
	io.WriteString(g, str)
	g.Close()
	return buff.String()
}
