package main

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

type Capture map[string]interface{}

func (capture Capture) Write(request *http.Request, reqBody io.Reader, response *ResponseWrapper) {
	capture["url"] = request.URL.Path
	capture["method"] = request.Method
	capture["request"] = createRequestMap(request, reqBody)
	capture["response"] = createResponseMap(response)
}

func createRequestMap(request *http.Request, reqBody io.Reader) map[string]interface{} {
	return createHeaderAndBodyMap(request.Header, reqBody)
}

func createResponseMap(response *ResponseWrapper) map[string]interface{} {
	responseMap := createHeaderAndBodyMap(response.Header(), response.Body)
	responseMap["status"] = response.Status
	return responseMap
}

func createHeaderAndBodyMap(headers http.Header, body io.Reader) map[string]interface{} {
	obj := make(map[string]interface{})
	obj["headers"] = getHeaders(headers)
	obj["body"] = getBody(headers, body)
	return obj
}

func getHeaders(headers http.Header) map[string]string {
	flatHeaders := make(map[string]string)
	for key, values := range headers {
		flatHeaders[key] = strings.Join(values, "; ")
	}
	return flatHeaders
}

func getBody(headers http.Header, body io.Reader) map[string]interface{} {
	body = unzip(headers, body)
	bbody, _ := ioutil.ReadAll(body)
	bodyUnmarshal := make(map[string]interface{})
	json.Unmarshal(bbody, &bodyUnmarshal)
	return bodyUnmarshal
}

func unzip(headers http.Header, body io.Reader) io.Reader {
	if headers.Get("Content-Encoding") == "gzip" {
		uncompressed, _ := gzip.NewReader(body)
		return uncompressed
	}
	return body
}
