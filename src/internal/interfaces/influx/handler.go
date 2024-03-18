package influx

import (
	"context"
	"github.com/Muscaw/GitFortress/internal/domain/metrics/entity"
	"github.com/Muscaw/GitFortress/internal/domain/metrics/service"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/rs/zerolog/log"
	"time"
)

type influxMetricHandler struct {
	influxDbServerUrl string
	influxDbAuthToken string
	org               string
	bucket            string
	metricChan        chan entity.Metric
}

func (i *influxMetricHandler) Handle(metric entity.Metric) {
	i.metricChan <- metric
}

func (i *influxMetricHandler) handleCounter(ctx context.Context, writeApi api.WriteAPIBlocking, counter entity.Counter) {
	values := counter.Values()
	interfaceValues := make(map[string]interface{}, len(values))
	for k, v := range values {
		interfaceValues[k] = v
	}

	point := influxdb2.NewPoint(counter.Name(), make(map[string]string, 0), interfaceValues, time.Now())
	err := writeApi.WritePoint(ctx, point)
	if err != nil {
		log.Error().Err(err).Msg("could not write point to influx")
		return
	}
}

func (i *influxMetricHandler) Start(ctx context.Context) {
	influxClient := influxdb2.NewClient(i.influxDbServerUrl, i.influxDbAuthToken)
	writeApi := influxClient.WriteAPIBlocking(i.org, i.bucket)

	for {
		select {
		case m := <-i.metricChan:
			switch metric := m.(type) {
			case entity.Counter:
				i.handleCounter(ctx, writeApi, metric)
			}
		case <-ctx.Done():
			return
		}
	}
}

func NewInfluxMetricsHandler(influxDbServerUrl string, influxDbAuthToken string, org string, bucket string) service.MetricsPort {
	return &influxMetricHandler{influxDbServerUrl: influxDbServerUrl, influxDbAuthToken: influxDbAuthToken, org: org, bucket: bucket, metricChan: make(chan entity.Metric)}
}
