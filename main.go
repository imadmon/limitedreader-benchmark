package main

import (
	"fmt"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/samber/lo"
)

const (
	benchmarkDataFile         = "docs/benchmark.json"
	benchmarkGraphFile        = "docs/benchmark.html"
	benchmarkAverageDataFile  = "docs/benchmarkAverage.json"
	benchmarkAverageGraphFile = "docs/benchmarkAverage.html"
	usageGraphFile            = "docs/usage.html"
)

func main() {
	// Usage()
	Benchmark()
	// LoadBenchmark()
	// BenchmarkWithAverage()
	// LoadBenchmarkWithAverage()
	// BenchmarkMultipleTimes()
}

func Usage() {
	graphs := []*charts.Line{
		goRateLimitUsageOnGraph(),
		goRateLimitBurstsOnlyOnGraph(),
		uberRateLimitUsageOnGraph(),
		imadmonRateLimitUsageOnGraph(),
	}

	WriteGraphsToFile(graphs, usageGraphFile)
}

func Benchmark() {
	data := RunBenchmark()
	saveDataToFile(data, benchmarkDataFile)
	GraphBenchmark(data, benchmarkGraphFile)
}

func LoadBenchmark() {
	data, err := loadDataFromFile(benchmarkDataFile)
	if err != nil {
		return
	}

	GraphBenchmark(data, benchmarkGraphFile)
}

func BenchmarkWithAverage() {
	const benchmarkAmount = 3
	fmt.Printf("Running benchmark with average of %d iterations\n", benchmarkAmount)

	benchmarkResults := make([]AllBenchmarkData, benchmarkAmount)
	for i := 0; i < benchmarkAmount; i++ {
		fmt.Printf("Running Benchmark #%d\n", i+1)
		benchmarkResults[i] = RunBenchmark()
	}

	result := getAllBenchmarkAverage(benchmarkResults)
	fmt.Printf("Finished running benchmark with average of %d iterations\n", benchmarkAmount)

	saveDataToFile(result, benchmarkAverageDataFile)
	GraphBenchmark(result, benchmarkAverageGraphFile)
}

func LoadBenchmarkWithAverage() {
	data, err := loadDataFromFile(benchmarkAverageDataFile)
	if err != nil {
		return
	}

	GraphBenchmark(data, benchmarkAverageGraphFile)
}

func BenchmarkMultipleTimes() {
	const benchmarkAmount = 5
	fmt.Printf("Running benchmark %d times\n", benchmarkAmount)

	for i := 0; i < benchmarkAmount; i++ {
		data := RunBenchmark()
		saveDataToFile(data, addNumberToFilename(benchmarkDataFile, i+1))
		GraphBenchmark(data, addNumberToFilename(benchmarkGraphFile, i+1))
	}
	fmt.Printf("Finished running benchmark %d times\n", benchmarkAmount)
}

func getAllBenchmarkAverage(benchmarkResults []AllBenchmarkData) AllBenchmarkData {
	result := make(AllBenchmarkData)
	for benchmarkType, _ := range benchmarkResults[0] {
		result[benchmarkType] = getBenchmarkAverage(benchmarkResults, benchmarkType)
	}
	return result
}

func getBenchmarkAverage(benchmarkResults []AllBenchmarkData, benchmarkType BenchmarkType) BenchmarkData {
	result := make(BenchmarkData)
	for readerType, _ := range benchmarkResults[0][benchmarkType] {
		result[readerType] = getBenchmarkReaderAverage(benchmarkResults, benchmarkType, readerType)
	}
	return result
}

func getBenchmarkReaderAverage(benchmarkResults []AllBenchmarkData, benchmarkType BenchmarkType, readerType ReaderType) BenchmarkReaderData {
	result := make(BenchmarkReaderData)
	for monitorType, seriesData := range benchmarkResults[0][benchmarkType][readerType] {
		result[monitorType] = SeriesData{
			Title:  seriesData.Title,
			Values: getBenchmarkReaderMonitorAverage(benchmarkResults, benchmarkType, readerType, monitorType),
			Color:  seriesData.Color,
		}
	}
	return result
}

func getBenchmarkReaderMonitorAverage(benchmarkResults []AllBenchmarkData, benchmarkType BenchmarkType, readerType ReaderType, monitorType MonitorValueType) []int {
	results := make([][]int, 0)
	for _, benchmarkResult := range benchmarkResults {
		results = append(results, benchmarkResult[benchmarkType][readerType][monitorType].Values)
	}

	seriesAmount := len(results)
	valuesAmount := len(lo.MinBy(results, func(a, b []int) bool {
		return len(a) < len(b)
	}))
	result := make([]int, valuesAmount)
	for i := 0; i < valuesAmount; i++ {
		var sum int
		for j := 0; j < seriesAmount; j++ {
			sum += results[j][i]
		}

		result[i] = sum / seriesAmount
	}

	return result
}
