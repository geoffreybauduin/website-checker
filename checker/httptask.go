package checker

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/sirupsen/logrus"
)

type httptask struct {
	url  string
	from *url.URL
}

func (task *httptask) Execute(c *checker) (*result, error) {
	if task.alreadyChecked(c) {
		return nil, nil
	}
	logger := logrus.WithField("url", task.url)
	req, errRequest := http.NewRequest("GET", task.url, nil)
	if errRequest != nil {
		return nil, errRequest
	}
	logger.Debug("requesting url")
	resp, errResponse := http.DefaultClient.Do(req)
	if errResponse != nil {
		return nil, errResponse
	}
	logger.Debug("got response")
	res := &result{
		StatusCode: resp.StatusCode,
	}
	defer resp.Body.Close()
	if c.inspectExternal == false && task.isExternalLink() {
		return nil, nil
	}
	doc, errDoc := goquery.NewDocumentFromReader(resp.Body)
	if errDoc != nil {
		return nil, errDoc
	}
	doc.Find("a").Each(func(_ int, s *goquery.Selection) {
		if aUrl, ok := s.Attr("href"); ok {
			if strings.HasPrefix(aUrl, "#") {
				// TODO: check anchor
			} else {
				urlAdded, err := task.addUrlToCheck(c, aUrl, req.URL)
				if err != nil {
					panic(err)
				}
				res.registerLink(urlAdded)
			}
		}
	})
	doc.Find("img").Each(func(_ int, s *goquery.Selection) {
		if aUrl, ok := s.Attr("src"); ok {
			urlAdded, err := task.addUrlToCheck(c, aUrl, req.URL)
			if err != nil {
				panic(err)
			}
			res.registerLink(urlAdded)
		}
	})
	doc.Find("script").Each(func(_ int, s *goquery.Selection) {
		if aUrl, ok := s.Attr("src"); ok {
			urlAdded, err := task.addUrlToCheck(c, aUrl, req.URL)
			if err != nil {
				panic(err)
			}
			res.registerLink(urlAdded)
		}
	})
	doc.Find(`link[rel="stylesheet"]`).Each(func(_ int, s *goquery.Selection) {
		if aUrl, ok := s.Attr("href"); ok {
			urlAdded, err := task.addUrlToCheck(c, aUrl, req.URL)
			if err != nil {
				panic(err)
			}
			res.registerLink(urlAdded)
		}
	})
	c.result.Lock()
	c.result.Checked[task.url] = res
	c.result.Unlock()
	return res, nil
}

func (h httptask) isExternalLink() bool {
	if h.from == nil {
		return false
	}
	part := fmt.Sprintf("%s://%s", h.from.Scheme, h.from.Host)
	return !strings.HasPrefix(h.url, part)
}

func (h httptask) alreadyChecked(c *checker) bool {
	c.result.RLock()
	_, ok := c.result.Checked[h.url]
	c.result.RUnlock()
	return ok
}

func (h httptask) addUrlToCheck(c *checker, urlToAdd string, from *url.URL) (string, error) {
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
	if c.isIgnored(urlToAdd) {
		logrus.Debugf("url is ignored")
		return urlToAdd, nil
	}
	c.addTask(&httptask{url: urlToAdd, from: from})
	return urlToAdd, nil
}
