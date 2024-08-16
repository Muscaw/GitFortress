package influx

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Muscaw/GitFortress/internal/domain/metrics/entity"
	"github.com/Muscaw/GitFortress/internal/domain/metrics/service"
)

type fakeMetricsService struct {
	metricsPort service.MetricsPort
}

func (f *fakeMetricsService) Push(metric entity.MetricInformation, valueNames []string) {
	f.metricsPort.Handle(metric, valueNames)
}

type metricRequestInformation struct {
	requestUri string
	body       string
}

func verifyMetricIsPushed(t *testing.T, fullMetricNameAndValue string, requestInformationChan chan metricRequestInformation, org string, bucket string) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	select {
	case requestInformation := <-requestInformationChan:
		if !strings.Contains(requestInformation.requestUri, fmt.Sprintf("bucket=%v&org=%v", bucket, org)) {
			t.Errorf("request uri does not contain org and bucket: %v", requestInformation.requestUri)
		}

		if !strings.Contains(requestInformation.body, fullMetricNameAndValue) {
			t.Fatalf("could not find metric in request body. Expected: %v. got %v", fullMetricNameAndValue, requestInformation.body)
		}
		fmt.Println(requestInformation)
	case <-ctx.Done():
		t.FailNow()
	}
}

func Test_influx_handler_pushes_metrics_as_expected(t *testing.T) {
	requestInformationChan := make(chan metricRequestInformation)
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.FailNow()
		}
		requestInformationChan <- metricRequestInformation{r.RequestURI, string(body)}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer testServer.Close()

	org := "some-org"
	bucket := "some-bucket"
	influxMetricsHandler := NewInfluxMetricsHandler(
		MetricHandlerOpts{
			InfluxDBUrl:       testServer.URL,
			InfluxDBAuthToken: "some-token",
			InfluxDBOrg:       org,
			InfluxDBBucket:    bucket,
			MetricNamePrefix:  "gitfortress",
		})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go influxMetricsHandler.Start(ctx, nil)

	metricsService := &fakeMetricsService{metricsPort: influxMetricsHandler}

	counter := entity.NewCounter("some_counter", metricsService)
	counter.Increment("some_value")
	verifyMetricIsPushed(t, "gitfortress_some_counter some_value=1", requestInformationChan, org, bucket)

	counter.Increment("some_value")
	verifyMetricIsPushed(t, "gitfortress_some_counter some_value=2", requestInformationChan, org, bucket)

	gauge := entity.NewGauge("some_gauge", metricsService)
	gauge.SetFloat("some_value", 1)
	verifyMetricIsPushed(t, "gitfortress_some_gauge some_value=1", requestInformationChan, org, bucket)

	gauge.SetFloat("some_value", 2.1)
	verifyMetricIsPushed(t, "gitfortress_some_gauge some_value=2.1", requestInformationChan, org, bucket)

	gauge.SetInt("some_int", 3)
	verifyMetricIsPushed(t, "gitfortress_some_gauge some_int=3i,some_value=2.1", requestInformationChan, org, bucket)

	gauge.SetInts(map[string]int{"some_value": 10, "some_int": 5})
	verifyMetricIsPushed(t, "gitfortress_some_gauge some_int=5i,some_value=10i", requestInformationChan, org, bucket)
}
