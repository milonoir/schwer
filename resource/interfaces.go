package resource

// Load is implemented by resource load controllers.
type Load interface {
	Start()
	Stop()
	Update(int64)
}

// Monitor is implemented by resource consumption monitors.
type Monitor interface {
	Start()
	Stop()
	Usage() interface{}
}
