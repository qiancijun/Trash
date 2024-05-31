package internal

type IScrab interface {
	Init() error
	Run()
}