package checker

import (
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/Jeffail/tunny"
	"github.com/schollz/progressbar/v2"
)

type checker struct {
	tasks   []task
	result  *CheckResult
	checked int
	total   int
	lock    sync.RWMutex
	pool    *tunny.Pool
	wg      sync.WaitGroup
	options *Options
}

type Options struct {
	Workers                       int
	IgnoreRegexps                 []*regexp.Regexp
	URLs                          []string
	InspectExternal               bool
	StructuredDataCheckWithYandex bool
	YandexAPIKey                  string
}

func New(o *Options) Checker {
	return &checker{
		options: o,
	}
}

func (c *checker) addTask(t task) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.total++
	c.tasks = append(c.tasks, t)
	c.wg.Add(1)
}

func (c *checker) isIgnored(url string) bool {
	for _, re := range c.options.IgnoreRegexps {
		if re.MatchString(url) {
			return true
		}
	}
	return false
}

type errors []error

func (e errors) Error() string {
	m := make([]string, len(e))
	for idx, err := range e {
		m[idx] = err.Error()
	}
	return strings.Join(m, ", ")
}

func (c *checker) Run() (*CheckResult, error) {
	if c.result != nil {
		return c.result, nil
	}
	c.result = &CheckResult{
		Checked: map[string]*result{},
	}
	bar := progressbar.NewOptions(
		len(c.options.URLs),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionShowCount(),
	)
	mu := sync.RWMutex{}
	errs := make(errors, 0)
	c.pool = tunny.NewFunc(c.options.Workers, func(in interface{}) interface{} {
		defer c.wg.Done()
		t := in.(task)
		_, err := t.Execute(c)
		c.checked++
		bar.ChangeMax(c.total)
		bar.Add(1)
		if err != nil {
			mu.Lock()
			errs = append(errs, err)
			mu.Unlock()
		}
		return err
	})
	for _, url := range c.options.URLs {
		c.addTask(&httptask{url: url, from: nil})
	}
	go func() {
		for {
			for len(c.tasks) > 0 {
				c.processTask()
			}
			if len(c.tasks) == 0 && c.pool.QueueLength() > 0 {
				time.Sleep(1 * time.Second)
			}
		}
	}()
	c.wg.Wait()
	if len(errs) > 0 {
		return nil, errs
	}
	return c.result, nil
}

func (c *checker) processTask() {
	c.lock.Lock()
	t := c.tasks[0]
	if len(c.tasks) > 1 {
		c.tasks = c.tasks[1:]
	} else {
		c.tasks = []task{}
	}
	c.lock.Unlock()
	go c.pool.Process(t)
}
