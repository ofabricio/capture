package main

type Capture struct {
	Url      string `json:"url"`
	Method   string `json:"method"`
	Status   int    `json:"status"`
	Request  string `json:"request"`
	Response string `json:"response"`
}

type Captures struct {
	items []Capture
	max   int
}

func (c *Captures) Add(capture Capture) {
	c.items = append([]Capture{capture}, c.items...)
	if len(c.items) > c.max {
		c.items = c.items[:len(c.items)-1]
	}
}
