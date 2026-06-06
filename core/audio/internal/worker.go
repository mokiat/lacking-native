package internal

type Worker interface {
	Schedule(func())
}
