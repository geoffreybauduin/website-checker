package checker

import (
	"sync"

	validator "github.com/geoffreybauduin/yandex-structured-data-validator"
)

// CheckResult is the result of a full check performed
type CheckResult struct {
	sync.RWMutex
	Checked map[string]*result `json:"checked"`
}

type result struct {
	sync.RWMutex
	StatusCode     int                    `json:"status_code"`
	Links          []string               `json:"links"`
	StructuredData []ResultStructuredData `json:"structured_data"`
}

func (r *result) registerLink(url string) {
	r.Lock()
	defer r.Unlock()

	if r.Links == nil {
		r.Links = make([]string, 0)
	}
	r.Links = append(r.Links, url)
}

// ResultStructuredData stores multiple informations regarding structured data validation
type ResultStructuredData struct {
	Content string                      `json:"content"`
	Yandex  *validator.StandardResponse `json:"yandex"`
}
