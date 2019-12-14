package checker

import "sync"

type CheckResult struct {
	sync.RWMutex
	Checked map[string]*result `json:"checked"`
}

type result struct {
	sync.RWMutex
	StatusCode int      `json:"status_code"`
	Links      []string `json:"links"`
}

func (r *result) registerLink(url string) {
	r.Lock()
	defer r.Unlock()

	if r.Links == nil {
		r.Links = make([]string, 0)
	}
	r.Links = append(r.Links, url)
}
