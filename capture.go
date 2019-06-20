package main

import (
	"net/http"
	"strconv"
	"sync"
	"time"
)

var captureID int

// CaptureService handles captures
type CaptureService struct {
	items    []Capture
	mu       sync.RWMutex
	maxItems int
	updated  chan struct{} // signals any change in "items"
}

// Capture is our traffic data
type Capture struct {
	ID  int
	Req *http.Request
	Res *http.Response
	// Elapsed time of the request, in milliseconds
	Elapsed time.Duration
}

// DashboardItem is an item in the dashboard's list
type DashboardItem struct {
	ID     int    `json:"id"`
	Path   string `json:"path"`
	Method string `json:"method"`
	Status int    `json:"status"`

	Elapsed time.Duration `json:"elapsed"`
}

// CaptureDump is all the dumps shown in the dashboard
type CaptureDump struct {
	Request  string `json:"request"`
	Response string `json:"response"`
	Curl     string `json:"curl"`
}

// NewCaptureService creates a new service of captures
func NewCaptureService(maxItems int) *CaptureService {
	return &CaptureService{
		maxItems: maxItems,
		updated:  make(chan struct{}),
	}
}

// Insert inserts a new capture
func (s *CaptureService) Insert(capture Capture) {
	s.mu.Lock()
	defer s.mu.Unlock()

	captureID++
	capture.ID = captureID
	s.items = append(s.items, capture)
	if len(s.items) > s.maxItems {
		s.items = s.items[1:]
	}
	s.signalsUpdate()
}

// Find finds a capture by its id
func (s *CaptureService) Find(captureID string) *Capture {
	s.mu.RLock()
	defer s.mu.RUnlock()

	idInt, _ := strconv.Atoi(captureID)
	for _, c := range s.items {
		if c.ID == idInt {
			return &c
		}
	}
	return nil
}

// RemoveAll removes all the captures
func (s *CaptureService) RemoveAll() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.items = nil
	s.signalsUpdate()
}

// DashboardItems returns the dashboard's list of items
func (s *CaptureService) DashboardItems() []DashboardItem {
	s.mu.RLock()
	defer s.mu.RUnlock()

	metadatas := make([]DashboardItem, len(s.items))
	for i, capture := range s.items {
		metadatas[i] = DashboardItem{
			ID:      capture.ID,
			Path:    capture.Req.URL.Path,
			Method:  capture.Req.Method,
			Status:  capture.Res.StatusCode,
			Elapsed: capture.Elapsed,
		}
	}
	return metadatas
}

// signalsUpdate fires an update signal
func (s *CaptureService) signalsUpdate() {
	close(s.updated)
	s.updated = make(chan struct{})
}

// Updated signals any change in this service,
// like inserting or removing captures
func (s *CaptureService) Updated() <-chan struct{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.updated
}
