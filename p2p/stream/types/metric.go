package sttypes

import (
	"github.com/prometheus/client_golang/prometheus"
	prom "github.com/zennittians/intelchain/api/service/prometheus"
)

func init() {
	prom.PromRegistry().MustRegister(
		bytesReadCounter,
		bytesWriteCounter,
		msgReadCounter,
		msgWriteCounter,
		msgReadFailedCounterVec,
		msgWriteFailedCounterVec,
	)
}

var (
	bytesReadCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "itc",
			Subsystem: "stream",
			Name:      "bytes_read",
			Help:      "total bytes read from stream",
		},
	)

	bytesWriteCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "itc",
			Subsystem: "stream",
			Name:      "bytes_write",
			Help:      "total bytes write to stream",
		},
	)

	msgReadCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "itc",
			Subsystem: "stream",
			Name:      "msg_read",
			Help:      "number of messages read from stream",
		},
	)

	msgWriteCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "itc",
			Subsystem: "stream",
			Name:      "msg_write",
			Help:      "number of messages write to stream",
		},
	)

	msgReadFailedCounterVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "itc",
			Subsystem: "stream",
			Name:      "msg_read_failed",
			Help:      "number of messages failed reading from stream",
		},
		[]string{"error"},
	)

	msgWriteFailedCounterVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "itc",
			Subsystem: "stream",
			Name:      "msg_write_failed",
			Help:      "number of messages failed writing to stream",
		},
		[]string{"error"},
	)
)
