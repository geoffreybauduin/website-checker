package checker

// Checker describes the interface to be fullfiled to implement
// a valid checker
type Checker interface {
	Run() (*CheckResult, error)
}

type task interface {
	Execute(*checker) (*result, error)
}
