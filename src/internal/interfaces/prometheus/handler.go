package prometheus

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Muscaw/GitFortress/internal/domain/metrics/entity"
	"github.com/Muscaw/GitFortress/internal/domain/metrics/service"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
)

type handleTuple struct {
	metric     entity.Metric
	valueNames []string
}

type metricHandler struct {
	counters         map[string]prometheus.Counter
	gauges           map[string]prometheus.Gauge
	autoConvertNames bool
	metricPrefixName string
}

func newMetricHandler(autoConvertNames bool, metricPrefixName string) metricHandler {
	return metricHandler{
		counters:         map[string]prometheus.Counter{},
		gauges:           map[string]prometheus.Gauge{},
		autoConvertNames: autoConvertNames,
		metricPrefixName: metricPrefixName,
	}
}

func (m *metricHandler) getCounterName(metric entity.Counter, valueName string) string {
	format := "%v_%v"
	if m.autoConvertNames {
		format = "%v_%v_total"
	}

	if m.metricPrefixName != "" {
		return fmt.Sprintf(fmt.Sprintf("%v_%v", m.metricPrefixName, format), metric.Name(), valueName)
	} else {
		return fmt.Sprintf(format, metric.Name(), valueName)
	}
}

func (m *metricHandler) handleCounter(counter entity.Counter, valueNames []string) {
	for _, valueName := range valueNames {
		name := m.getCounterName(counter, valueName)
		val, ok := m.counters[name]
		if !ok {
			val = promauto.NewCounter(prometheus.CounterOpts{
				Name: name,
			})
			m.counters[name] = val
		}

		val.Inc()
	}
}

func (m *metricHandler) getGaugeName(metric entity.Gauge, valueName string) string {
	if m.metricPrefixName != "" {
		return fmt.Sprintf("%v_%v_%v", m.metricPrefixName, metric.Name(), valueName)
	} else {
		return fmt.Sprintf("%v_%v", metric.Name(), valueName)
	}
}

func convertToFloat(value any) (float64, bool) {
	switch v := value.(type) {
	case float32:
		return float64(v), true
	case float64:
		return v, true
	case int:
		return float64(v), true
	case int8:
		return float64(v), true
	case int16:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint8:
		return float64(v), true
	case uint16:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint64:
		return float64(v), true
	}
	return 0, false
}

func (m *metricHandler) handleGauge(gauge entity.Gauge, valueNames []string) {
	for _, valueName := range valueNames {
		name := m.getGaugeName(gauge, valueName)
		val, ok := m.gauges[name]
		if !ok {
			val = promauto.NewGauge(prometheus.GaugeOpts{Name: name})
			m.gauges[name] = val
		}

		convertedValue, ok := convertToFloat(gauge.Values()[valueName])
		if ok {
			val.Set(convertedValue)
		} else {
			log.Warn().Msgf("could not convert value to float for metric %v", name)
		}
	}
}

type prometheusMetricHandler struct {
	server           *http.Server
	exposedPort      int
	autoConvertNames bool
	metricChan       chan handleTuple
	metricHandler    metricHandler
}

func (p *prometheusMetricHandler) handleMetric(ctx context.Context) {
	for {
		select {
		case m := <-p.metricChan:
			switch metric := m.metric.(type) {
			case entity.Counter:
				p.metricHandler.handleCounter(metric, m.valueNames)
			case entity.Gauge:
				p.metricHandler.handleGauge(metric, m.valueNames)
			default:
				log.Warn().Msgf("metric type %T is currently unsupported by prometheus handler", metric)
			}

		case <-ctx.Done():
			log.Info().Msg("finished processing prometheus handler")
			p.server.Shutdown(context.Background())
			return
		}
	}
}

func (p *prometheusMetricHandler) Start(ctx context.Context) {
	go p.handleMetric(ctx)
	if err := p.server.ListenAndServe(); err != http.ErrServerClosed {
		log.Err(err).Msgf("could not start http listener on port %v", p.exposedPort)
	}
}

func (p *prometheusMetricHandler) Handle(metric entity.Metric, valueNames []string) {
	p.metricChan <- handleTuple{metric, valueNames}
}

type MetricsHandlerOpts struct {
	ExposedPort      int
	AutoConvertNames bool
	MetricPrefix     string
}

func NewPrometheusMetricsHandler(options MetricsHandlerOpts) service.MetricsPort {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	server := &http.Server{Addr: fmt.Sprintf(":%v", options.ExposedPort), Handler: mux}
	return &prometheusMetricHandler{server: server, exposedPort: options.ExposedPort, autoConvertNames: options.AutoConvertNames, metricHandler: newMetricHandler(options.AutoConvertNames, options.MetricPrefix), metricChan: make(chan handleTuple)}
}
