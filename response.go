package main

import (
	"bytes"
	"io"
	"net/http"
)

type ResponseWrapper struct {
	http.ResponseWriter
	Status int
	Body   io.Reader
}

func NewResponseWrapper(response http.ResponseWriter) *ResponseWrapper {
	return &ResponseWrapper{response, http.StatusInternalServerError, nil}
}

func (response *ResponseWrapper) WriteHeader(code int) {
	response.Status = code
	response.ResponseWriter.WriteHeader(code)
}

func (response *ResponseWrapper) Write(body []byte) (int, error) {
	response.Body = bytes.NewBuffer(body)
	return response.ResponseWriter.Write(body)
}

func (response *ResponseWrapper) Header() http.Header {
	return response.ResponseWriter.Header()
}
