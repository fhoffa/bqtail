package contract

import (
	"github.com/viant/bqtail/stage/activity"
)

//Metrics represents BqQuery jobs metric
type Metrics struct {
	CopyJobs  int `json:",omitempty"`
	QueryJobs int `json:",omitempty"`
	LoadJobs  int `json:",omitempty"`
	OtherJobs int `json:",omitempty"`
	BatchJobs int `json:",omitempty"`
}

//Count returns total metrics count
func (m Metrics) Count() int {
	return m.QueryJobs + m.CopyJobs + m.LoadJobs + m.OtherJobs
}

//Update updates a metrics with job ID
func (m *Metrics) Update(jobID string) *activity.Meta {
	stageInfo := activity.Parse(jobID)
	m.Add(stageInfo, 1)
	return stageInfo
}

//Put updates a metrics with supplied stage action and count
func (m *Metrics) Add(stageInfo *activity.Meta, count int) {
	switch stageInfo.Action {
	case "query":
		m.QueryJobs += count
	case "copy":
		m.CopyJobs += count
	case "load", "reload":
		m.LoadJobs += count
	default:
		m.OtherJobs += count
	}
}

//Merge merges metrics
func (m *Metrics) Merge(metrics *Metrics) {
	m.CopyJobs += metrics.CopyJobs
	m.QueryJobs += metrics.QueryJobs
	m.LoadJobs += metrics.LoadJobs
	m.OtherJobs += metrics.OtherJobs
}
