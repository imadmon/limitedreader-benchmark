package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync/atomic"
	"time"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/imadmon/limitedreader"
	"go.uber.org/ratelimit"
	"golang.org/x/time/rate"
)

const (
	LineTitle = "Bytes/0.2 Milliseconds"
	LineText  = "#5470c6"
)

func goRateLimitUsageOnGraph() *charts.Line {
	const chunkSize = 1024
	const dataSize = 20 * chunkSize
	const limit = 5 * chunkSize

	reader := bytes.NewBuffer(make([]byte, dataSize))
	limiter := rate.NewLimiter(rate.Limit(limit/chunkSize), limit/chunkSize)

	var total atomic.Int64
	buffer := make([]byte, chunkSize)
	ctx, ctxCancel := context.WithCancel(context.Background())
	resultsC := make(chan []int)

	go usagesMonitorLoop(ctx, &total, resultsC)
	time.Sleep(300 * time.Millisecond)

	start := time.Now()
	for {
		err := limiter.Wait(context.Background())
		if err != nil {
			fmt.Printf("limiter.Wait err: %v\n", err)
		}

		n, err := reader.Read(buffer)
		total.Add(int64(n))
		if err != nil {
			if err != io.EOF {
				fmt.Printf("Error: %v\n", err)
			}
			break
		}
	}
	elapsed := time.Since(start)

	time.Sleep(700 * time.Millisecond)
	ctxCancel()
	time.Sleep(300 * time.Millisecond)
	fmt.Printf("Total: %d, Elapsed: %s\n", total.Load(), elapsed)
	results := <-resultsC

	return GenerateGraphChart(
		"Golang Burstable Rate Limiting",
		"Normal Usage",
		nil,
		[]SeriesData{
			{
				Title:  LineTitle,
				Values: results,
				Color:  LineText,
			},
		},
	)
}

func goRateLimitBurstsOnlyOnGraph() *charts.Line {
	const chunkSize = 1024
	const dataSize = 20 * chunkSize
	const burstSize = 5

	reader := bytes.NewBuffer(make([]byte, dataSize))
	limiter := rate.NewLimiter(rate.Every(time.Second/burstSize), burstSize)

	var total atomic.Int64
	buffer := make([]byte, chunkSize)
	ctx, ctxCancel := context.WithCancel(context.Background())
	resultsC := make(chan []int)

	go usagesMonitorLoop(ctx, &total, resultsC)
	time.Sleep(300 * time.Millisecond)

	start := time.Now()
	for {
		// try to reserve 1 token
		res := limiter.ReserveN(time.Now(), 1)
		if !res.OK() {
			fmt.Println("Couldn't reserve token")
			ctxCancel()
			return nil
		}

		// if there's a delay, cancel and wait until we have full burst
		if res.Delay() > 0 {
			fmt.Printf("No tokens! waiting until we have full burst at %.2fs...\n", time.Since(start).Seconds())
			res.Cancel()

			// wait until we get enough token to burst
			res = limiter.ReserveN(time.Now(), burstSize)
			if !res.OK() {
				fmt.Println("Couldn't reserve tokens to burst")
				ctxCancel()
				return nil
			}
			wait := res.Delay()
			res.Cancel()

			time.Sleep(wait)
			fmt.Printf("Burst tokens reached, Resuming at %.2fs...\n", time.Since(start).Seconds())
		}

		fmt.Printf("Operation at %.2fs Current tokens: %.f\n", time.Since(start).Seconds(), limiter.Tokens())
		n, err := reader.Read(buffer)
		total.Add(int64(n))
		if err != nil {
			if err != io.EOF {
				fmt.Printf("Error: %v\n", err)
			}
			break
		}
	}
	elapsed := time.Since(start)

	time.Sleep(700 * time.Millisecond)
	ctxCancel()
	time.Sleep(300 * time.Millisecond)
	fmt.Printf("Total: %d, Elapsed: %s\n", total.Load(), elapsed)
	results := <-resultsC

	return GenerateGraphChart(
		"Golang Burstable Rate Limiting",
		"Only Bursts Usage",
		nil,
		[]SeriesData{
			{
				Title:  LineTitle,
				Values: results,
				Color:  LineText,
			},
		},
	)
}

func uberRateLimitUsageOnGraph() *charts.Line {
	const chunkSize = 1024
	const dataSize = 20 * chunkSize
	const limit = 5

	reader := bytes.NewBuffer(make([]byte, dataSize))
	limiter := ratelimit.New(limit)

	var total atomic.Int64
	buffer := make([]byte, chunkSize)
	ctx, ctxCancel := context.WithCancel(context.Background())
	resultsC := make(chan []int)

	go usagesMonitorLoop(ctx, &total, resultsC)
	time.Sleep(300 * time.Millisecond)

	start := time.Now()
	for {
		limiter.Take()

		n, err := reader.Read(buffer)
		total.Add(int64(n))
		if err != nil {
			if err != io.EOF {
				fmt.Printf("Error: %v\n", err)
			}
			break
		}
	}
	elapsed := time.Since(start)

	time.Sleep(700 * time.Millisecond)
	ctxCancel()
	time.Sleep(300 * time.Millisecond)
	fmt.Printf("Total: %d, Elapsed: %s\n", total.Load(), elapsed)
	results := <-resultsC

	return GenerateGraphChart(
		"Uber Deterministic Rate Limiting",
		"Normal Usage",
		nil,
		[]SeriesData{
			{
				Title:  LineTitle,
				Values: results,
				Color:  LineText,
			},
		},
	)
}

func imadmonRateLimitUsageOnGraph() *charts.Line {
	const chunkSize = 1024
	const dataSize = 20 * chunkSize
	const limit = 5 * chunkSize

	reader := bytes.NewBuffer(make([]byte, dataSize))
	limitedReader := limitedreader.NewRateLimitedReader(reader, int64(limit))

	var total atomic.Int64
	buffer := make([]byte, chunkSize)
	ctx, ctxCancel := context.WithCancel(context.Background())
	resultsC := make(chan []int)

	go usagesMonitorLoop(ctx, &total, resultsC)
	time.Sleep(300 * time.Millisecond)

	start := time.Now()
	for {
		n, err := limitedReader.Read(buffer)
		total.Add(int64(n))
		if err != nil {
			if err != io.EOF {
				fmt.Printf("Error: %v\n", err)
			}
			break
		}
	}
	elapsed := time.Since(start)

	time.Sleep(700 * time.Millisecond)
	ctxCancel()
	time.Sleep(300 * time.Millisecond)
	fmt.Printf("Total: %d, Elapsed: %s\n", total.Load(), elapsed)
	results := <-resultsC

	return GenerateGraphChart(
		"IMadmon Deterministic Rate Limiting",
		"Normal Usage",
		nil,
		[]SeriesData{
			{
				Title:  LineTitle,
				Values: results,
				Color:  LineText,
			},
		},
	)
}

func usagesMonitorLoop(ctx context.Context, monitoredBytes *atomic.Int64, resultsC chan []int) {
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	results := make([]int, 0)
	var prevResult int64
	for {
		select {
		case <-ctx.Done():
			resultsC <- results
			return
		case <-ticker.C:
			currentResult := monitoredBytes.Load()
			results = append(results, int(currentResult-prevResult))
			prevResult = currentResult
		}
	}
}
