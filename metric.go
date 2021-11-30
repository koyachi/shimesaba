package shimesaba

import (
	"fmt"
	"log"
	"math"
	"time"

	"github.com/mashiike/shimesaba/internal/timeutils"
)

// Metric handles aggregated Mackerel metrics
type Metric struct {
	id                  string
	values              map[time.Time][]float64
	aggregationInterval time.Duration
	aggregationMethod   func([]float64) float64
	interpolatedValue   *float64
	startAt             time.Time
	endAt               time.Time
}

func NewMetric(cfg *MetricConfig) *Metric {
	return &Metric{
		id:                  cfg.ID,
		values:              make(map[time.Time][]float64),
		aggregationInterval: cfg.DurationAggregation(),
		aggregationMethod:   getAggregationMethod(cfg.AggregationMethod),
		startAt:             time.Date(9999, 12, 31, 59, 59, 59, 999999999, time.UTC),
		interpolatedValue:   cfg.InterpolatedValue,
		endAt:               time.Unix(0, 0).In(time.UTC),
	}
}

func getAggregationMethod(str string) func([]float64) float64 {
	totalFunc := func(values []float64) float64 {
		t := 0.0
		for _, v := range values {
			t += v
		}
		return t
	}
	switch str {
	case "total", "sum":
		return totalFunc
	case "avg":
		return func(values []float64) float64 {
			if len(values) == 0 {
				return math.NaN()
			}
			t := totalFunc(values)
			return t / float64(len(values))
		}
	case "max":
	case "":
		log.Println("[warn] aggregation_method is empty. select default method `max`")
	default:
		log.Printf("[warn] aggregation_method `%s` is not found. select default method `max`\n", str)
	}
	return func(values []float64) float64 {
		maxValue := 0.0
		for _, v := range values {
			if v > maxValue {
				maxValue = v
			}
		}
		return maxValue
	}

}

// ID is the identifier of the metric
func (m *Metric) ID() string {
	return m.id
}

// AppendValue adds a value to the metric
func (m *Metric) AppendValue(t time.Time, v interface{}) error {
	t = t.Truncate(m.aggregationInterval)

	var value float64
	switch v := v.(type) {
	case float64:
		value = v
	case int64:
		value = float64(v)
	case int32:
		value = float64(v)
	case uint64:
		value = float64(v)
	case uint32:
		value = float64(v)
	case float32:
		value = float64(v)
	case int:
		value = float64(v)
	default:
		return fmt.Errorf("Metric.Append() unknown value type = %T", v)
	}
	values, ok := m.values[t]
	if !ok {
		values = make([]float64, 0, 1)
	}
	values = append(values, value)
	m.values[t] = values
	if m.startAt.After(t) {
		m.startAt = t
	}
	if m.endAt.Before(t) {
		m.endAt = t
	}
	return nil
}

// GetValue gets the value at the specified time
func (m *Metric) GetValue(t time.Time) (float64, bool) {
	t = t.Truncate(m.aggregationInterval)
	values, ok := m.values[t]
	if !ok {
		if m.interpolatedValue != nil {
			return *m.interpolatedValue, true
		}
		return math.NaN(), false
	}
	return m.aggregationMethod(values), true
}

// GetValues ​​gets the values ​​for the specified time period
func (m *Metric) GetValues(startAt time.Time, endAt time.Time) map[time.Time]float64 {
	iter := timeutils.NewIterator(
		startAt,
		endAt,
		m.aggregationInterval,
	)
	ret := make(map[time.Time]float64)
	for iter.HasNext() {
		curAt, _ := iter.Next()
		if v, ok := m.GetValue(curAt); ok {
			ret[curAt] = v
		}
	}
	return ret
}

// StartAt returns the start time of the metric
func (m *Metric) StartAt() time.Time {
	return m.startAt
}

// EndAt returns the end time of the metric
func (m *Metric) EndAt() time.Time {
	return m.endAt.Add(m.aggregationInterval - time.Nanosecond)
}

// AggregationInterval returns the aggregation interval for metrics
func (m *Metric) AggregationInterval() time.Duration {
	return m.aggregationInterval
}

//String implements fmt.Stringer
func (m *Metric) String() string {
	return fmt.Sprintf("[id:%s len(values):%d aggregate_interval:%s, range:%s~%s<%s>]", m.id, len(m.values), m.aggregationInterval, m.startAt, m.endAt, m.endAt.Sub(m.startAt))
}

// Metrics is a collection of metrics
type Metrics map[string]*Metric

// Set adds a metric to the collection
func (ms Metrics) Set(m *Metric) {
	ms[m.ID()] = m
}

// Get uses an identifier to get the metric
func (ms Metrics) Get(id string) (*Metric, bool) {
	m, ok := ms[id]
	return m, ok
}

func (ms Metrics) ToSlice() []*Metric {
	ret := make([]*Metric, 0, len(ms))
	for _, m := range ms {
		ret = append(ret, m)
	}
	return ret
}

// ToSlice converts the collection to Slice
func (ms Metrics) String() string {
	return fmt.Sprintf("%v", ms.ToSlice())
}

// StartAt returns the earliest start time in the metric in the collection
func (ms Metrics) StartAt() time.Time {

	startAt := time.Date(9999, 12, 31, 59, 59, 59, 999999999, time.UTC)
	for _, m := range ms {
		if startAt.After(m.StartAt()) {
			startAt = m.StartAt()
		}
	}

	return startAt
}

// EndAt returns the latest end time of the metric in the collection
func (ms Metrics) EndAt() time.Time {
	endAt := time.Unix(0, 0).In(time.UTC)
	for _, m := range ms {
		if endAt.Before(m.EndAt()) {
			endAt = m.EndAt()
		}
	}
	return endAt
}

// AggregationInterval returns the longest aggregation period for the metric in the collection
func (ms Metrics) AggregationInterval() time.Duration {
	ret := time.Duration(0)
	for _, m := range ms {
		a := m.AggregationInterval()
		if a > ret {
			ret = a
		}
	}
	return ret
}
