package resource

// CPULevels is the type returned by the Usage() method of a CPU load monitor.
type CPULevels []int

// MemStats is the type returned by the Usage() method of a memory load monitor.
type MemStats struct {
	Total     int `json:"total"`
	Available int `json:"available"`
	Used      int `json:"used"`
	UsedPct   int `json:"usedpct"`
}
