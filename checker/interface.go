package checker

import "regexp"

type Checker interface {
	AddURL(string)
	Run() (*CheckResult, error)
	Ignore(*regexp.Regexp)
	NoExternalInspection()
	Workers(int)
}

type task interface {
	Execute(*checker) (*result, error)
}
