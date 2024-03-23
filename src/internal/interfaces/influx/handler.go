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

type handleTuple struct {
	metric     entity.Metric
	valueNames []string
}

type influxMetricHandler struct {
	influxDbServerUrl string
	influxDbAuthToken string
	org               string
	bucket            string
	metricChan        chan handleTuple
}

func (i *influxMetricHandler) Handle(metric entity.Metric, valueNames []string) {
	i.metricChan <- handleTuple{metric, valueNames}
}

func (i *influxMetricHandler) handleCounter(ctx context.Context, writeApi api.WriteAPIBlocking, counter entity.Counter) {
	values := counter.Values()
	interfaceValues := make(map[string]interface{}, len(values))
	for k, v := range values {
		interfaceValues[k] = v
	}

	i.handleMetric(ctx, writeApi, counter.Name(), interfaceValues)
}

func (i *influxMetricHandler) handleGauge(ctx context.Context, writeApi api.WriteAPIBlocking, gauge entity.Gauge) {
	i.handleMetric(ctx, writeApi, gauge.Name(), gauge.Values())
}

func (i *influxMetricHandler) handleMetric(ctx context.Context, writeApi api.WriteAPIBlocking, metricName string, values map[string]any) {

	point := influxdb2.NewPoint(metricName, make(map[string]string, 0), values, time.Now())
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
			switch metric := m.metric.(type) {
			case entity.Counter:
				i.handleCounter(ctx, writeApi, metric)
			case entity.Gauge:
				i.handleGauge(ctx, writeApi, metric)
			default:
				log.Warn().Msgf("metric type %T is currently unsupported by influx handler", metric)
			}

		case <-ctx.Done():
			log.Info().Msg("finished processing influxdb handler")
			return
		}
	}
}

func NewInfluxMetricsHandler(influxDbServerUrl string, influxDbAuthToken string, org string, bucket string) service.MetricsPort {
	return &influxMetricHandler{influxDbServerUrl: influxDbServerUrl, influxDbAuthToken: influxDbAuthToken, org: org, bucket: bucket, metricChan: make(chan handleTuple)}
}
