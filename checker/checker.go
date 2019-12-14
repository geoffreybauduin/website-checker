package checker

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/sirupsen/logrus"
)

type checker struct {
	URLs            []string
	toCheck         []checkTask
	result          *CheckResult
	ignoreRegexps   []*regexp.Regexp
	inspectExternal bool
}

type checkTask struct {
	url  string
	from *url.URL
}

func (c checkTask) isExternalLink() bool {
	if c.from == nil {
		return false
	}
	part := fmt.Sprintf("%s://%s", c.from.Scheme, c.from.Host)
	return !strings.HasPrefix(c.url, part)
}

func New() Checker {
	return &checker{
		URLs:            make([]string, 0),
		ignoreRegexps:   make([]*regexp.Regexp, 0),
		inspectExternal: true,
	}
}

func (c *checker) NoExternalInspection() {
	c.inspectExternal = false
}

func (c *checker) Ignore(re *regexp.Regexp) {
	c.ignoreRegexps = append(c.ignoreRegexps, re)
}

func (c *checker) addUrlToCheck(urlToAdd string, from *url.URL) (string, error) {
	if strings.HasPrefix(urlToAdd, "/") {
		urlToAdd = fmt.Sprintf("//%s%s", from.Host, urlToAdd)
	}
	if strings.HasPrefix(urlToAdd, "//") {
		urlToAdd = fmt.Sprintf("%s:%s", from.Scheme, urlToAdd)
	}
	urlObj, err := url.Parse(urlToAdd)
	if err != nil {
		return "", err
	}
	if urlObj.Host == "" {
		// relative url
		urlObj.Host = from.Host
	}
	if urlObj.Scheme == "" {
		// same scheme than where it comes from
		urlObj.Scheme = from.Scheme
	}
	if urlObj.Scheme == "mailto" {
		return urlToAdd, nil
	}
	urlToAdd = urlObj.String()
	if c.alreadyChecked(urlToAdd) {
		// don't re-check
		logrus.Debugf("url already checked")
		return urlToAdd, nil
	} else if c.isIgnored(urlToAdd) {
		logrus.Debugf("url is ignored")
		return urlToAdd, nil
	}
	c.toCheck = append(c.toCheck, checkTask{url: urlToAdd, from: from})
	return urlToAdd, nil
}

func (c *checker) alreadyChecked(url string) bool {
	_, ok := c.result.Checked[url]
	return ok
}

func (c *checker) isIgnored(url string) bool {
	for _, re := range c.ignoreRegexps {
		if re.MatchString(url) {
			return true
		}
	}
	return false
}

func (c *checker) AddURL(url string) {
	c.URLs = append(c.URLs, url)
}

func (c *checker) Run() (*CheckResult, error) {
	if c.result != nil {
		return c.result, nil
	}
	c.result = &CheckResult{
		Checked: map[string]*result{},
	}
	// copy
	for _, url := range c.URLs {
		c.addUrlToCheck(url, nil)
	}
	checked := 0
	for len(c.toCheck) > 0 {
		logrus.Infof("%d checked, %d remaining", checked, len(c.toCheck))
		task := c.toCheck[0]
		if !c.alreadyChecked(task.url) {
			logrus.Infof(task.url)
			err := c.runCheck(task)
			if err != nil {
				return nil, err
			}
		} else {
			logrus.Debugf("already checked")
		}
		c.toCheck = c.toCheck[1:]
		checked++
	}
	return c.result, nil
}

func (c *checker) runCheck(task checkTask) error {
	logger := logrus.WithField("url", task.url)
	req, errRequest := http.NewRequest("GET", task.url, nil)
	if errRequest != nil {
		return errRequest
	}
	logger.Debug("requesting url")
	resp, errResponse := http.DefaultClient.Do(req)
	if errResponse != nil {
		return errResponse
	}
	logger.Debug("got response")
	c.result.Checked[task.url] = &result{
		StatusCode: resp.StatusCode,
	}
	defer resp.Body.Close()
	if c.inspectExternal == false && task.isExternalLink() {
		logger.Warnf("is external")
		return nil
	}
	doc, errDoc := goquery.NewDocumentFromReader(resp.Body)
	if errDoc != nil {
		return errDoc
	}
	doc.Find("a").Each(func(_ int, s *goquery.Selection) {
		if aUrl, ok := s.Attr("href"); ok {
			if strings.HasPrefix(aUrl, "#") {
				// TODO: check anchor
			} else {
				urlAdded, err := c.addUrlToCheck(aUrl, req.URL)
				if err != nil {
					panic(err)
				}
				c.result.Checked[task.url].registerLink(urlAdded)
			}
		}
	})
	doc.Find("img").Each(func(_ int, s *goquery.Selection) {
		if aUrl, ok := s.Attr("src"); ok {
			urlAdded, err := c.addUrlToCheck(aUrl, req.URL)
			if err != nil {
				panic(err)
			}
			c.result.Checked[task.url].registerLink(urlAdded)
		}
	})
	doc.Find("script").Each(func(_ int, s *goquery.Selection) {
		if aUrl, ok := s.Attr("src"); ok {
			urlAdded, err := c.addUrlToCheck(aUrl, req.URL)
			if err != nil {
				panic(err)
			}
			c.result.Checked[task.url].registerLink(urlAdded)
		}
	})
	doc.Find(`link[rel="stylesheet"]`).Each(func(_ int, s *goquery.Selection) {
		if aUrl, ok := s.Attr("href"); ok {
			urlAdded, err := c.addUrlToCheck(aUrl, req.URL)
			if err != nil {
				panic(err)
			}
			c.result.Checked[task.url].registerLink(urlAdded)
		}
	})
	return nil
}
