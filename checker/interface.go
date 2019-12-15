package checker

type Checker interface {
	Run() (*CheckResult, error)
}

type task interface {
	Execute(*checker) (*result, error)
}
