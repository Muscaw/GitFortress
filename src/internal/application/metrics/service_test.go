package metrics

import (
	"context"
	"sync"
	"testing"

	"github.com/Muscaw/GitFortress/internal/domain/metrics/entity"
)

func Test_metricsService_creates_only_one_instance(t *testing.T) {
	metricsService1 := GetMetricsService()
	metricsService2 := GetMetricsService()

	if metricsService1 != metricsService2 {
		t.Fatal("metrics service are not the same")
	}
}

type fakePort struct {
	isStarted         bool
	handledMetric     entity.Metric
	handledValueNames []string
	startWg           sync.WaitGroup
	handleWg          sync.WaitGroup
	handleCallCount   int
}

func (f *fakePort) Start(ctx context.Context) {
	defer f.startWg.Done()
	f.isStarted = true
}

func (f *fakePort) Handle(metric entity.Metric, valueNames []string) {
	defer f.handleWg.Done()
	f.handledMetric = metric
	f.handledValueNames = valueNames
	f.handleCallCount += 1
}

func Test_metricsService_TrackCounter(t *testing.T) {
	// Arrange
	metricsService := newMetricsService()
	fakePort := fakePort{}
	metricsService.RegisterHandler(&fakePort)
	fakePort.startWg.Add(1)
	metricsService.Start(context.Background())
	fakePort.startWg.Wait()

	// Act
	fakePort.handleWg.Add(1)
	counter := metricsService.TrackCounter("counter_name")
	counter.Increment("some_value")
	fakePort.handleWg.Wait()

	// Assert
	if !fakePort.isStarted {
		t.Fatal("port is not started")
	}

	if fakePort.handleCallCount != 1 {
		t.Fatalf("handle call count is different from 1. Got %v", fakePort.handleCallCount)
	}

	if fakePort.handledMetric != counter {
		t.Fatalf("fakePort handled an incorrect metric. Expected %v, got %v", counter, fakePort.handledMetric)
	}

	if fakePort.handledValueNames[0] != "some_value" {
		t.Fatalf("fakePort handled an incorrect value name. Expected some_value, got %v", fakePort.handledValueNames[0])
	}
}

func Test_metricsService_TrackGauge(t *testing.T) {
	// Arrange
	metricsService := newMetricsService()
	fakePort := fakePort{}
	metricsService.RegisterHandler(&fakePort)
	fakePort.startWg.Add(1)
	metricsService.Start(context.Background())
	fakePort.startWg.Wait()

	// Act
	fakePort.handleWg.Add(1)
	counter := metricsService.TrackGauge("counter_name")
	counter.SetInt("some_value", 2)
	fakePort.handleWg.Wait()

	// Assert
	if !fakePort.isStarted {
		t.Fatal("port is not started")
	}

	if fakePort.handleCallCount != 1 {
		t.Fatalf("handle call count is different from 1. Got %v", fakePort.handleCallCount)
	}

	if fakePort.handledMetric != counter {
		t.Fatalf("fakePort handled an incorrect metric. Expected %v, got %v", counter, fakePort.handledMetric)
	}

	if fakePort.handledValueNames[0] != "some_value" {
		t.Fatalf("fakePort handled an incorrect value name. Expected some_value, got %v", fakePort.handledValueNames[0])
	}
}

func Test_metricsService_StartMultipleHandlers(t *testing.T) {
	// Arrange
	metricsService := newMetricsService()

	fakePort1 := fakePort{}
	fakePort2 := fakePort{}
	metricsService.RegisterHandler(&fakePort1)
	metricsService.RegisterHandler(&fakePort2)

	// Act
	fakePort1.startWg.Add(1)
	fakePort2.startWg.Add(1)
	metricsService.Start(context.Background())
	fakePort1.startWg.Wait()
	fakePort2.startWg.Wait()

	if !fakePort1.isStarted || !fakePort2.isStarted {
		t.Fatal("one of the fake ports is not started")
	}
}
