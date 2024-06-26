package prometheus

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/Muscaw/GitFortress/internal/domain/metrics/entity"
	"github.com/Muscaw/GitFortress/internal/domain/metrics/service"
)

type fakeMetricsService struct {
	metricsPort service.MetricsPort
}

func (f *fakeMetricsService) Push(metric entity.Metric, valueNames []string) {
	f.metricsPort.Handle(metric, valueNames)
}

func getMetricsBody(t *testing.T) string {
	res, err := http.Get("http://localhost:8080/metrics")
	if err != nil {
		t.Fatal("could not get metrics endpoint")
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.FailNow()
	}
	return string(body)
}

func Test_published_counters_are_present_on_page(t *testing.T) {
	prometheusHandler := NewPrometheusMetricsHandler(
		MetricsHandlerOpts{
			ExposedPort:      8080,
			AutoConvertNames: false,
			MetricPrefix:     "gitfortress",
		},
	)

	ctx := context.Background()
	go prometheusHandler.Start(ctx)
	defer ctx.Done()

	body := getMetricsBody(t)

	if strings.Contains(string(body), "gitfortress_") {
		t.Fatalf("expects no gitfortress metrics at the moment. Got \n%v", string(body))
	}

	fakeMetricsService := fakeMetricsService{metricsPort: prometheusHandler}
	counter := entity.NewCounter("some_counter", &fakeMetricsService)

	counter.Increment("some_value")

	body = getMetricsBody(t)

	if !strings.Contains(string(body), "gitfortress_some_counter_some_value 1") {
		t.Fatalf("expects gitfortress metrics at the moment. Got \n%v", string(body))
	}

	counter.Increment("some_value")

	body = getMetricsBody(t)

	if !strings.Contains(string(body), "gitfortress_some_counter_some_value 2") {
		t.Fatalf("expects gitfortress metrics at the moment. Got \n%v", string(body))
	}
}

func Test_published_gauges_are_present_on_page(t *testing.T) {
	prometheusHandler := NewPrometheusMetricsHandler(
		MetricsHandlerOpts{
			ExposedPort:      8080,
			AutoConvertNames: false,
			MetricPrefix:     "gitfortress",
		},
	)

	ctx := context.Background()
	go prometheusHandler.Start(ctx)
	defer ctx.Done()

	body := getMetricsBody(t)

	if strings.Contains(string(body), "gitfortress_") {
		t.Fatalf("expects no gitfortress metrics at the moment. Got \n%v", string(body))
	}

	fakeMetricsService := fakeMetricsService{metricsPort: prometheusHandler}
	gauge := entity.NewGauge("some_gauge", &fakeMetricsService)

	gauge.SetInt("some_value", 0)

	body = getMetricsBody(t)

	if !strings.Contains(string(body), "gitfortress_some_gauge_some_value 0") {
		t.Fatalf("expects gitfortress metrics at the moment. Got \n%v", string(body))
	}

	gauge.SetFloat("some_value", 1.2)
	gauge.SetInts(map[string]int{"some_other_value": 1, "some_different_value": 3})

	body = getMetricsBody(t)

	if !strings.Contains(string(body), "gitfortress_some_gauge_some_value 1.2") {
		t.Fatalf("expects gitfortress metrics at the moment. Got \n%v", string(body))
	}
	if !strings.Contains(string(body), "gitfortress_some_gauge_some_other_value 1") {
		t.Fatalf("expects gitfortress metrics at the moment. Got \n%v", string(body))
	}
	if !strings.Contains(string(body), "gitfortress_some_gauge_some_different_value 3") {
		t.Fatalf("expects gitfortress metrics at the moment. Got \n%v", string(body))
	}
}
