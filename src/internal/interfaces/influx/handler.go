package influx

import (
	"context"
	"fmt"
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
	metricNamePrefix  string
	metricChan        chan handleTuple
}

func (i *influxMetricHandler) Handle(metric entity.Metric, valueNames []string) {
	i.metricChan <- handleTuple{metric, valueNames}
}

func (i *influxMetricHandler) getName(metric entity.Metric) string {
	if i.metricNamePrefix != "" {
		return fmt.Sprintf("%v_%v", i.metricNamePrefix, metric.Name())
	} else {
		return metric.Name()
	}
}

func (i *influxMetricHandler) handleCounter(ctx context.Context, writeApi api.WriteAPIBlocking, counter entity.Counter) {
	values := counter.Values()
	interfaceValues := make(map[string]interface{}, len(values))
	for k, v := range values {
		interfaceValues[k] = v
	}

	i.handleMetric(ctx, writeApi, i.getName(counter), interfaceValues)
}

func (i *influxMetricHandler) handleGauge(ctx context.Context, writeApi api.WriteAPIBlocking, gauge entity.Gauge) {
	i.handleMetric(ctx, writeApi, i.getName(gauge), gauge.Values())
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

type MetricHandlerOpts struct {
	InfluxDBUrl       string
	InfluxDBAuthToken string
	InfluxDBOrg       string
	InfluxDBBucket    string
	MetricNamePrefix  string
}

func NewInfluxMetricsHandler(opts MetricHandlerOpts) service.MetricsPort {
	return &influxMetricHandler{
		influxDbServerUrl: opts.InfluxDBUrl,
		influxDbAuthToken: opts.InfluxDBAuthToken, org: opts.InfluxDBOrg, bucket: opts.InfluxDBBucket, metricNamePrefix: opts.MetricNamePrefix, metricChan: make(chan handleTuple)}
}
