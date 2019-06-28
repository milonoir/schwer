package main

import (
	"github.com/milonoir/schwer/resource"
)

// Controller controls resource loads and monitors.
type Controller struct {
	cpuLoad    resource.Load
	memLoad    resource.Load
	cpuMonitor resource.Monitor
	memMonitor resource.Monitor
}

// NewController returns a new Controller.
func NewController(cpuLoad, memLoad resource.Load, cpuMonitor, memMonitor resource.Monitor) *Controller {
	return &Controller{
		cpuLoad:    cpuLoad,
		memLoad:    memLoad,
		cpuMonitor: cpuMonitor,
		memMonitor: memMonitor,
	}
}

// Start starts up resource monitors and loads.
func (c *Controller) Start() {
	c.cpuMonitor.Start()
	c.memMonitor.Start()
	c.cpuLoad.Start()
	c.memLoad.Start()
}

// Stop stops resource loads and monitors.
func (c *Controller) Stop() {
	c.cpuLoad.Stop()
	c.memLoad.Stop()
	c.cpuMonitor.Stop()
	c.memMonitor.Stop()
}

// UpdateCPULoad sends an update to the CPU load.
func (c *Controller) UpdateCPULoad(pct int64) {
	c.cpuLoad.Update(pct)
}

// UpdateMemLoad sends an update to the memory load.
func (c *Controller) UpdateMemLoad(size int64) {
	c.memLoad.Update(size)
}

// CPUUtilisationLevels returns the latest CPU utilisation levels from the CPU load monitor.
func (c *Controller) CPUUtilisationLevels() resource.CPULevels {
	return c.cpuMonitor.Usage().(resource.CPULevels)
}

// MemStats returns the latest memory stats from the memory load monitor.
func (c *Controller) MemStats() resource.MemStats {
	return c.memMonitor.Usage().(resource.MemStats)
}
