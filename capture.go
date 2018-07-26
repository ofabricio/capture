package main

import "strconv"

type Capture struct {
	ID       int    `json:"id"`
	Path     string `json:"path"`
	Method   string `json:"method"`
	Status   int    `json:"status"`
	Request  string `json:"request"`
	Response string `json:"response"`
}

type CaptureRef struct {
	ID      int    `json:"id"`
	Path    string `json:"path"`
	Method  string `json:"method"`
	Status  int    `json:"status"`
	ItemUrl string `json:"itemUrl"`
}

type Captures []Capture

func (items *Captures) Add(capture Capture) {
	*items = append(*items, capture)
	size := len(*items)
	if size > args.maxCaptures {
		*items = (*items)[1:]
	}
}

func (items *Captures) ToReferences(itemBaseUrl string) []CaptureRef {
	refs := make([]CaptureRef, len(*items))
	for i, item := range *items {
		refs[i] = CaptureRef{
			ID:      item.ID,
			Path:    item.Path,
			Method:  item.Method,
			Status:  item.Status,
			ItemUrl: itemBaseUrl + strconv.Itoa(i),
		}
	}
	return refs
}
