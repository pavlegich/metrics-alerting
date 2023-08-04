package storage

type (
	Metric interface {
		Type() string
		Name() string
		Value() string
	}

	gauge struct {
		mtype string
		name  string
		value string
	}

	counter struct {
		mtype string
		name  string
		value string
	}
)

func (m gauge) Type() string {
	return m.mtype
}

func (m gauge) Name() string {
	return m.name
}

func (m gauge) Value() string {
	return m.value
}

func NewGauge(metricType string, metricName string, metricValue string) Metric {
	return &gauge{mtype: metricType, name: metricName, value: metricValue}
}

func (m counter) Type() string {
	return m.mtype
}

func (m counter) Name() string {
	return m.name
}

func (m counter) Value() string {
	return m.value
}

func NewCounter(metricType string, metricName string, metricValue string) Metric {
	return &counter{mtype: metricType, name: metricName, value: metricValue}
}
