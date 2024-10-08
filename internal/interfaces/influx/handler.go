package influx

import (
	"context"
	"fmt"
	"time"

	"github.com/Muscaw/GitFortress/internal/domain/metrics/entity"
	"github.com/Muscaw/GitFortress/internal/domain/metrics/service"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/rs/zerolog/log"
)

type handleTuple struct {
	metricInformation entity.MetricInformation
	valueNames        []string
}

type influxMetricHandler struct {
	influxDbServerUrl string
	influxDbAuthToken string
	org               string
	bucket            string
	metricNamePrefix  string
	metricChan        chan handleTuple
}

func (i *influxMetricHandler) Handle(metricInformation entity.MetricInformation, valueNames []string) {
	i.metricChan <- handleTuple{metricInformation, valueNames}
}

func (i *influxMetricHandler) getName(metric entity.MetricInformation) string {
	if i.metricNamePrefix != "" {
		return fmt.Sprintf("%v_%v", i.metricNamePrefix, metric.MetricName())
	} else {
		return metric.MetricName()
	}
}

func (i *influxMetricHandler) handleCounter(ctx context.Context, writeApi api.WriteAPIBlocking, counter entity.MetricInformation) {
	values := counter.Values()
	interfaceValues := make(map[string]interface{}, len(values))
	for k, v := range values {
		interfaceValues[k] = v
	}

	i.handleMetric(ctx, writeApi, i.getName(counter), interfaceValues)
}

func (i *influxMetricHandler) handleGauge(ctx context.Context, writeApi api.WriteAPIBlocking, gauge entity.MetricInformation) {
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

func (i *influxMetricHandler) Start(ctx context.Context, doneFunc service.DoneFunc) {
	defer doneFunc()
	influxClient := influxdb2.NewClient(i.influxDbServerUrl, i.influxDbAuthToken)
	writeApi := influxClient.WriteAPIBlocking(i.org, i.bucket)

	for {
		select {
		case m := <-i.metricChan:
			switch m.metricInformation.MetricType() {
			case entity.COUNTER_METRIC_TYPE:
				i.handleCounter(ctx, writeApi, m.metricInformation)
			case entity.GAUGE_METRIC_TYPE:
				i.handleGauge(ctx, writeApi, m.metricInformation)
			default:
				log.Warn().Msgf("metric type %v is currently unsupported by influx handler", m.metricInformation.MetricType())
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
