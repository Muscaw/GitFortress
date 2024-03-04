package influx

import (
	"github.com/Muscaw/GitFortress/internal/domain/metrics/entity"
	"github.com/Muscaw/GitFortress/internal/domain/metrics/service"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

type influxMetricHandler struct {
	influxDbServerUrl string
	influxDbAuthToken string
	org               string
	bucket            string
	monitoredMetrics  []entity.Metric
}

func monitorCounter(counter entity.Counter, quit chan struct{}) {

}

func (i *influxMetricHandler) Start(metrics chan entity.Metric, quit chan struct{}) {
	influxClient := influxdb2.NewClient(i.influxDbServerUrl, i.influxDbAuthToken)
	api := influxClient.WriteAPIBlocking(i.org, i.bucket)

	for {
		select {
		case m := <-metrics:
			i.monitoredMetrics = append(i.monitoredMetrics, m)
			switch m.(type) {
			case entity.Counter:
				monitorCounter(m, quit)
			}
		case <-quit:
			return
		}
	}
}

func NewInfluxMetricsHandler(influxDbServerUrl string, influxDbAuthToken string, org string, bucket string) service.MetricsHandler {
	return &influxMetricHandler{influxDbServerUrl: influxDbServerUrl, influxDbAuthToken: influxDbAuthToken, org: org, bucket: bucket, monitoredMetrics: []*entity.Metric}
}
