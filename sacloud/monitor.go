package sacloud

import (
	"math"
	"time"
)

type MonitorValue struct {
	CPUTime *float64 `json:"CPU-TIME,omitempty"`
	Write   *float64 `json:",omitempty"`
	Read    *float64 `json:",omitempty"`
	Receive *float64 `json:",omitempty"`
	Send    *float64 `json:",omitempty"`
}

type ResourceMonitorRequest struct {
	Start *time.Time `json:",omitempty"`
	End   *time.Time `json:",omitempty"`
}
type ResourceMonitorResponse struct {
	Data *MonitorValues `json:",omitempty"`
}

type MonitorSummaryData struct {
	Max   float64
	Min   float64
	Avg   float64
	Count float64
}
type MonitorSummary struct {
	CPU  *MonitorSummaryData
	Disk *struct {
		Write *MonitorSummaryData
		Read  *MonitorSummaryData
	}
	Interface *struct {
		Receive *MonitorSummaryData
		Send    *MonitorSummaryData
	}
}

type MonitorValues map[string]*MonitorValue

type FlatMonitorValue struct {
	Time  time.Time
	Value float64
}

func (m *MonitorValues) Calc() *MonitorSummary {

	res := &MonitorSummary{}
	res.CPU = m.calcBy(func(v *MonitorValue) *float64 { return v.CPUTime })
	res.Disk = &struct {
		Write *MonitorSummaryData
		Read  *MonitorSummaryData
	}{
		Write: m.calcBy(func(v *MonitorValue) *float64 { return v.Write }),
		Read:  m.calcBy(func(v *MonitorValue) *float64 { return v.Read }),
	}
	res.Interface = &struct {
		Receive *MonitorSummaryData
		Send    *MonitorSummaryData
	}{
		Receive: m.calcBy(func(v *MonitorValue) *float64 { return v.Receive }),
		Send:    m.calcBy(func(v *MonitorValue) *float64 { return v.Send }),
	}

	return res
}

func (m *MonitorValues) calcBy(f func(m *MonitorValue) *float64) *MonitorSummaryData {
	res := &MonitorSummaryData{}
	var sum float64
	for _, data := range map[string]*MonitorValue(*m) {
		value := f(data)
		if value != nil {
			res.Count++
			res.Min = math.Min(res.Min, *value)
			res.Max = math.Max(res.Max, *value)
			sum += *value
		}
	}
	if sum > 0 && res.Count > 0 {
		res.Avg = sum / res.Count
	}

	return res
}

func (m *MonitorValues) FlattenCPUTimeValue() ([]FlatMonitorValue, error) {
	return m.flattenValue(func(v *MonitorValue) *float64 { return v.CPUTime })
}
func (m *MonitorValues) FlattenDiskWriteValue() ([]FlatMonitorValue, error) {
	return m.flattenValue(func(v *MonitorValue) *float64 { return v.Write })
}
func (m *MonitorValues) FlattenDiskReadValue() ([]FlatMonitorValue, error) {
	return m.flattenValue(func(v *MonitorValue) *float64 { return v.Read })
}
func (m *MonitorValues) FlattenPacketSendValue() ([]FlatMonitorValue, error) {
	return m.flattenValue(func(v *MonitorValue) *float64 { return v.Send })
}
func (m *MonitorValues) FlattenPacketReceiveValue() ([]FlatMonitorValue, error) {
	return m.flattenValue(func(v *MonitorValue) *float64 { return v.Receive })
}

func (m *MonitorValues) flattenValue(f func(*MonitorValue) *float64) ([]FlatMonitorValue, error) {
	var res []FlatMonitorValue

	for k, v := range map[string]*MonitorValue(*m) {
		if f(v) == nil {
			continue
		}
		time, err := time.Parse(time.RFC3339, k) // RFC3339 ≒ ISO8601
		if err != nil {
			return res, err
		}
		res = append(res, FlatMonitorValue{
			Time:  time,
			Value: *f(v),
		})
	}
	return res, nil
}

func (m *MonitorValue) HasValue() bool {
	values := []*float64{m.CPUTime, m.Read, m.Receive, m.Send, m.Write}
	for _, v := range values {
		if v != nil {
			return true
		}
	}
	return false
}
