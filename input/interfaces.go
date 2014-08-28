package input

import "github.com/rcrowley/go-metrics"

type Statistics interface {
	GetStatistics() (metrics.Registry, error)
}
