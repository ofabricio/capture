package main

import (
	"net/http"
	"strconv"
	"sync"
)

var captureID int
var captures CaptureList

// CaptureList stores all captures
type CaptureList struct {
	items    []Capture
	mux      sync.Mutex
	maxItems int
	updated  chan struct{} // signals any change in "items"
}

// Capture saves our traffic data
type Capture struct {
	ID  int
	Req *http.Request
	Res *http.Response
}

// CaptureMetadata is the data for each list item in the dashboard
type CaptureMetadata struct {
	ID     int    `json:"id"`
	Path   string `json:"path"`
	Method string `json:"method"`
	Status int    `json:"status"`
}

// CaptureDump saves all the dumps shown in the dashboard
type CaptureDump struct {
	Request  string `json:"request"`
	Response string `json:"response"`
	Curl     string `json:"curl"`
}

// Metadata returns the metadada of a capture
func (c *Capture) Metadata() CaptureMetadata {
	return CaptureMetadata{
		ID:     c.ID,
		Path:   c.Req.URL.Path,
		Method: c.Req.Method,
		Status: c.Res.StatusCode,
	}
}

// NewCaptureList creates a new list of captures
func NewCaptureList(maxItems int) *CaptureList {
	return &CaptureList{
		maxItems: maxItems,
		updated:  make(chan struct{}),
	}
}

// Insert adds a new capture
func (c *CaptureList) Insert(capture Capture) {
	c.mux.Lock()
	defer c.mux.Unlock()
	capture.ID = newID()
	c.items = append(c.items, capture)
	if len(c.items) > c.maxItems {
		c.items = c.items[1:]
	}
	c.signalsChange()
}

// Find finds a capture by its id
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

// RemoveAll removes all the captures
func (c *CaptureList) RemoveAll() {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.items = nil
	c.signalsChange()
}

// Items returns all the captures
func (c *CaptureList) Items() []Capture {
	return c.items
}

// ItemsAsMetadata returns all the captures as metadata
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

func (c *CaptureList) signalsChange() {
	close(c.updated)
	c.updated = make(chan struct{})
}

// Updated signals any change in the list
func (c *CaptureList) Updated() <-chan struct{} {
	return c.updated
}
