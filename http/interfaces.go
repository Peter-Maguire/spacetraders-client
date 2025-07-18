package http

type State interface {
	SetWaitingForHttp(bool)
}
