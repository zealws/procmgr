package main

type Event interface {
	Id() string
}

type BeginEvent string
type EndEvent string
type FailEvent struct {
	Name string
	Err  Error
}

func (e BeginEvent) Id() string {
	return string(e)
}

func (e EndEvent) Id() string {
	return string(e)
}

func (e FailEvent) Id() string {
	return e.Name
}
