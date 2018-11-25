package main

import (
	"net/http"
	"strconv"
	"sync"
)

var captureID int
var captures CaptureList

type CaptureList struct {
	items    []Capture
	mux      sync.Mutex
	maxItems int
	// signals any change in "items"
	Updated chan struct{}
}

type Capture struct {
	ID  int
	Req *http.Request
	Res *http.Response
}

type CaptureMetadata struct {
	ID     int    `json:"id"`
	Path   string `json:"path"`
	Method string `json:"method"`
	Status int    `json:"status"`
}

type CaptureDump struct {
	Request  string `json:"request"`
	Response string `json:"response"`
	Curl     string `json:"curl"`
}

func (c *Capture) Metadata() CaptureMetadata {
	return CaptureMetadata{
		ID:     c.ID,
		Path:   c.Req.URL.Path,
		Method: c.Req.Method,
		Status: c.Res.StatusCode,
	}
}

func NewCaptureList(maxItems int) *CaptureList {
	return &CaptureList{
		maxItems: maxItems,
		Updated:  make(chan struct{}),
	}
}

func (c *CaptureList) Insert(capture Capture) {
	c.mux.Lock()
	defer c.mux.Unlock()
	capture.ID = newID()
	c.items = append(c.items, capture)
	if len(c.items) > c.maxItems {
		c.items = c.items[1:]
	}
	c.signalsItemsChange()
}

func (c *CaptureList) Find(captureID string) *Capture {
	c.mux.Lock()
	defer c.mux.Unlock()
	idInt, _ := strconv.Atoi(captureID)
	for _, c := range c.items {
		if c.ID == idInt {
			return &c
		}
	}
	return nil
}

func (c *CaptureList) RemoveAll() {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.items = nil
	c.signalsItemsChange()
}

func (c *CaptureList) Items() []Capture {
	return c.items
}

func (c *CaptureList) ItemsAsMetadata() []CaptureMetadata {
	c.mux.Lock()
	defer c.mux.Unlock()
	metadatas := make([]CaptureMetadata, len(c.items))
	for i, capture := range c.items {
		metadatas[i] = capture.Metadata()
	}
	return metadatas
}

func newID() int {
	captureID++
	return captureID
}

func (c *CaptureList) signalsItemsChange() {
	select {
	case c.Updated <- struct{}{}:
	default:
	}
}
