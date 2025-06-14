package main

import (
	"context"
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"time"
)

// Example Colors:
// "#5470c6", "#91cc75", "#fac858", "#ee6666", "#73c0de",
// "#3ba272", "#fc8452", "#9a60b4", "#ea7ccc",
const (
	GolangSeriesName   = "Golang"
	GolangSeriesColor  = "#5470c6"
	JujuSeriesName     = "Juju"
	JujuSeriesColor    = "#ea7ccc"
	UberSeriesName     = "Uber"
	UberSeriesColor    = "#fac858"
	IMadmonSeriesName  = "IMadmon"
	IMadmonSeriesColor = "#ee6666"
)

func RunTestWithMonitor(testFn BenchmarkTest, factory ReaderFactory,
	seriesName, color string, seriesValueTypes []MonitorValueType) BenchmarkReaderData {

	ctx, ctxCancel := context.WithCancel(context.Background())
	resultsC := make(chan []monitorResult)

	go monitorLoop(ctx, resultsC)
	time.Sleep(300 * time.Millisecond)

	RunTest(testFn, factory)

	time.Sleep(700 * time.Millisecond)
	ctxCancel()
	time.Sleep(300 * time.Millisecond)

	results := <-resultsC
	seriesData := make(BenchmarkReaderData)
	for _, seriesValueType := range seriesValueTypes {
		seriesData[seriesValueType] = SeriesData{
			Title:  seriesName,
			Values: parseGraphValue(results, seriesValueType),
			Color:  color,
		}
	}

	return seriesData
}

func RunTest(testFn BenchmarkTest, factory ReaderFactory) {
	testName := strings.TrimPrefix(filepath.Ext(runtime.FuncForPC(reflect.ValueOf(testFn).Pointer()).Name()), ".")
	factoryName := strings.TrimPrefix(filepath.Ext(runtime.FuncForPC(reflect.ValueOf(factory).Pointer()).Name()), ".")
	fmt.Printf("Starting %s using %s...\n", testName, factoryName)
	testFn(factory)
	time.Sleep(250 * time.Millisecond)
	fmt.Printf("Finished %s using %s\n", testName, factoryName)
	time.Sleep(250 * time.Millisecond)
}

func RunGolangTest(testFn BenchmarkTest, seriesValueTypes []MonitorValueType) BenchmarkReaderData {
	return RunTestWithMonitor(
		testFn,
		GolangBurstsRateLimitReaderFactory,
		GolangSeriesName,
		GolangSeriesColor,
		seriesValueTypes,
	)
}

func RunJujuTest(testFn BenchmarkTest, seriesValueTypes []MonitorValueType) BenchmarkReaderData {
	return RunTestWithMonitor(
		testFn,
		JujuBurstsRateLimitReaderFactory,
		JujuSeriesName,
		JujuSeriesColor,
		seriesValueTypes,
	)
}

func RunUberTest(testFn BenchmarkTest, seriesValueTypes []MonitorValueType) BenchmarkReaderData {
	return RunTestWithMonitor(
		testFn,
		UberDeterministicRateLimitReaderFactory,
		UberSeriesName,
		UberSeriesColor,
		seriesValueTypes,
	)
}

func RunIMadmonTest(testFn BenchmarkTest, seriesValueTypes []MonitorValueType) BenchmarkReaderData {
	return RunTestWithMonitor(
		testFn,
		IMadmonDeterministicRateLimitReaderFactory,
		IMadmonSeriesName,
		IMadmonSeriesColor,
		seriesValueTypes,
	)
}
