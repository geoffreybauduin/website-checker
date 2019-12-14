package checker

type CheckResult struct {
	Checked map[string]*result `json:"checked"`
}

type result struct {
	StatusCode int      `json:"status_code"`
	Links      []string `json:"links"`
}

func (r *result) registerLink(url string) {
	if r.Links == nil {
		r.Links = make([]string, 0)
	}
	r.Links = append(r.Links, url)
}
