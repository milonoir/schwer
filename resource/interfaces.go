package resource

// StartStopper is implemented by types which can be started and stopped.
type StartStopper interface {
	Start()
	Stop()
}

// Load is implemented by resource load controllers.
type Load interface {
	StartStopper
	Update(int64)
}

// Monitor is implemented by resource consumption monitors.
type Monitor interface {
	StartStopper
	Usage() interface{}
}
