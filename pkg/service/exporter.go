package service

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	sequenceStatus = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "nats_jetstream_sequence_status_is_abnormal",
		Help: "Check the NATS Jetstream sequence status is abnormal",
	}, []string{"stream"})
)
