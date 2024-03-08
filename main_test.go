package main

import (
	"compress/gzip"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
)

func Example() {

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Date", "Sun, 10 Mar 2024 01:05:03 GMT")
		if r.URL.Path == "/say" {
			w.Write([]byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit."))
		}
		if r.URL.Path == "/say/gzip" {
			w.Header().Set("Content-Encoding", "gzip")
			g := gzip.NewWriter(w)
			g.Write([]byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit."))
			g.Close()
		}
	}))

	u, _ := url.Parse(srv.URL)

	captures := make(chan Capture)

	proxy, _ := NewProxy(u, captures)

	// Test regular response.

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/say", nil)

	go proxy(rec, req)
	c := <-captures

	fmt.Println("Test regular response")
	fmt.Println(rec.Code == 200, rec.Body.String() == "Lorem ipsum dolor sit amet, consectetur adipiscing elit.")
	fmt.Println(c.Res == "HTTP/1.1 200 OK\r\nContent-Length: 56\r\nContent-Type: text/plain; charset=utf-8\r\nDate: Sun, 10 Mar 2024 01:05:03 GMT\r\n\r\nLorem ipsum dolor sit amet, consectetur adipiscing elit.")

	// Test gzip response.

	rec = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/say/gzip", nil)

	go proxy(rec, req)
	c = <-captures

	fmt.Println("Test gzip response")
	fmt.Println(rec.Code == 200, rec.Body.String() == "Lorem ipsum dolor sit amet, consectetur adipiscing elit.")
	fmt.Println(c.Res == "HTTP/1.1 200 OK\r\nConnection: close\r\nDate: Sun, 10 Mar 2024 01:05:03 GMT\r\n\r\nLorem ipsum dolor sit amet, consectetur adipiscing elit.")

	// Output:
	// Test regular response
	// true true
	// true
	// Test gzip response
	// true true
	// true
}
