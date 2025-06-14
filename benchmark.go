package main

import (
	"fmt"
	"io"
	"math"
	"strings"
	"time"
)

type BenchmarkTest func(ReaderFactory)
type BenchmarkReaderData map[MonitorValueType]SeriesData
type BenchmarkData map[ReaderType]BenchmarkReaderData
type AllBenchmarkData map[BenchmarkType]BenchmarkData

type BenchmarkType string

var (
	BenchmarkRateLimitingSynthetic       BenchmarkType = "BenchmarkRateLimitingSynthetic"
	BenchmarkRateLimitingRealWorldLocal  BenchmarkType = "BenchmarkRateLimitingRealWorldLocal"
	BenchmarkMaxReadOverTimeSynthetic    BenchmarkType = "BenchmarkMaxReadOverTimeSynthetic"
	BenchmarkSpikeRecoveryRealWorldLocal BenchmarkType = "BenchmarkSpikeRecoveryRealWorldLocal"
)

type SeriesData struct {
	Title  string
	Values []int
	Color  string
}

func RunBenchmark() AllBenchmarkData {
	return AllBenchmarkData{
		BenchmarkRateLimitingSynthetic:       RunBenchmarkRateLimitingSynthetic(),
		BenchmarkRateLimitingRealWorldLocal:  RunBenchmarkRateLimitingRealWorldLocal(),
		BenchmarkMaxReadOverTimeSynthetic:    RunBenchmarkMaxReadOverTimeSynthetic(),
		BenchmarkSpikeRecoveryRealWorldLocal: RunBenchmarkSpikeRecoveryRealWorldLocal(),
	}
}

func RunBenchmarkRateLimitingSynthetic() BenchmarkData {
	golangSeries := RunGolangTest(RateLimitingSyntheticTest, []MonitorValueType{SyntheticRX})
	jujuSeries := RunJujuTest(RateLimitingSyntheticTest, []MonitorValueType{SyntheticRX})
	uberSeries := RunUberTest(RateLimitingSyntheticTest, []MonitorValueType{SyntheticRX})
	imadmonSeries := RunIMadmonTest(RateLimitingSyntheticTest, []MonitorValueType{SyntheticRX})
	return BenchmarkData{
		GolangReader: {
			SyntheticRX: golangSeries[SyntheticRX],
		},
		JujuReader: {
			SyntheticRX: jujuSeries[SyntheticRX],
		},
		UberReader: {
			SyntheticRX: uberSeries[SyntheticRX],
		},
		IMadmonReader: {
			SyntheticRX: imadmonSeries[SyntheticRX],
		},
	}
}

func RunBenchmarkRateLimitingRealWorldLocal() BenchmarkData {
	golangSeries := RunGolangTest(RateLimitingRealWorldLocalTest, []MonitorValueType{RX, CPU, RAM})
	jujuSeries := RunJujuTest(RateLimitingRealWorldLocalTest, []MonitorValueType{RX, CPU, RAM})
	uberSeries := RunUberTest(RateLimitingRealWorldLocalTest, []MonitorValueType{RX, CPU, RAM})
	imadmonSeries := RunIMadmonTest(RateLimitingRealWorldLocalTest, []MonitorValueType{RX, CPU, RAM})
	return BenchmarkData{
		GolangReader: {
			RX:  golangSeries[RX],
			CPU: golangSeries[CPU],
			RAM: golangSeries[RAM],
		},
		JujuReader: {
			RX:  jujuSeries[RX],
			CPU: jujuSeries[CPU],
			RAM: jujuSeries[RAM],
		},
		UberReader: {
			RX:  uberSeries[RX],
			CPU: uberSeries[CPU],
			RAM: uberSeries[RAM],
		},
		IMadmonReader: {
			RX:  imadmonSeries[RX],
			CPU: imadmonSeries[CPU],
			RAM: imadmonSeries[RAM],
		},
	}
}

func RunBenchmarkMaxReadOverTimeSynthetic() BenchmarkData {
	golangSeries := RunGolangTest(MaxReadOverTimeSyntheticTest, []MonitorValueType{TotalSyntheticRX, CPU})
	jujuSeries := RunJujuTest(MaxReadOverTimeSyntheticTest, []MonitorValueType{TotalSyntheticRX, CPU})
	uberSeries := RunUberTest(MaxReadOverTimeSyntheticTest, []MonitorValueType{TotalSyntheticRX, CPU})
	imadmonSeries := RunIMadmonTest(MaxReadOverTimeSyntheticTest, []MonitorValueType{TotalSyntheticRX, CPU})
	return BenchmarkData{
		GolangReader: {
			TotalSyntheticRX: golangSeries[TotalSyntheticRX],
			CPU:              golangSeries[CPU],
		},
		JujuReader: {
			TotalSyntheticRX: jujuSeries[TotalSyntheticRX],
			CPU:              jujuSeries[CPU],
		},
		UberReader: {
			TotalSyntheticRX: uberSeries[TotalSyntheticRX],
			CPU:              uberSeries[CPU],
		},
		IMadmonReader: {
			TotalSyntheticRX: imadmonSeries[TotalSyntheticRX],
			CPU:              imadmonSeries[CPU],
		},
	}
}

func RunBenchmarkSpikeRecoveryRealWorldLocal() BenchmarkData {
	golangSeries := RunGolangTest(SpikeRecoveryRealWorldLocalTest, []MonitorValueType{RX, CPU, RAM})
	jujuSeries := RunJujuTest(SpikeRecoveryRealWorldLocalTest, []MonitorValueType{RX, CPU, RAM})
	uberSeries := RunUberTest(SpikeRecoveryRealWorldLocalTest, []MonitorValueType{RX, CPU, RAM})
	imadmonSeries := RunIMadmonTest(SpikeRecoveryRealWorldLocalTest, []MonitorValueType{RX, CPU, RAM})
	return BenchmarkData{
		GolangReader: {
			RX:  golangSeries[RX],
			CPU: golangSeries[CPU],
			RAM: golangSeries[RAM],
		},
		JujuReader: {
			RX:  jujuSeries[RX],
			CPU: jujuSeries[CPU],
			RAM: jujuSeries[RAM],
		},
		UberReader: {
			RX:  uberSeries[RX],
			CPU: uberSeries[CPU],
			RAM: uberSeries[RAM],
		},
		IMadmonReader: {
			RX:  imadmonSeries[RX],
			CPU: imadmonSeries[CPU],
			RAM: imadmonSeries[RAM],
		},
	}
}

func RateLimitingSyntheticTest(readerFactory ReaderFactory) {
	const dataSize = 100 * 1024 * 1024 // 100MB
	const bufferSize = 32 * 1024       // 32KB classic io.Copy
	const limit = dataSize / 4         // should take 4 seconds
	var total int

	reader := &syntheticReader{size: dataSize}
	limitedReader := readerFactory(reader, bufferSize, limit)
	buffer := make([]byte, bufferSize)

	start := time.Now()
	for {
		n, err := limitedReader.Read(buffer)
		total += n
		if err != nil {
			if err != io.EOF {
				fmt.Printf("Error: %v\n", err)
			}
			break
		}
	}
	elapsed := time.Since(start)

	if total != dataSize {
		fmt.Printf("Read incomplete data, read: %d expected: %d\n", total, dataSize)
	}

	fmt.Printf("RateLimitingSyntheticTest Took %v\n", elapsed)
}

func MaxReadOverTimeSyntheticTest(readerFactory ReaderFactory) {
	const durationInSeconds = 10
	const bufferSize = 32 * 1024 // 32KB classic io.Copy
	const limit = math.MaxInt    // bufferSize * 1_000_000_000 // large limit
	fmt.Printf("Duration set: %d seconds\n", durationInSeconds)

	buffer := make([]byte, bufferSize)
	var totalBytes int64

	reader := &syntheticReader{}
	rateLimitedReader := readerFactory(reader, bufferSize, limit)

	deadline := time.Now().Add(durationInSeconds * time.Second)
	for time.Now().Before(deadline) {
		n, err := rateLimitedReader.Read(buffer)
		if n > 0 {
			totalBytes += int64(n)
		}
		if err != nil {
			fmt.Printf("Read error: %v\n", err)
			break
		}
	}

	mb := float64(totalBytes) / 1024.0 / 1024.0
	fmt.Printf("MaxReadOverTimeSyntheticTest: Read %.3f MB in 10 seconds\n", mb)
}

func RateLimitingRealWorldLocalTest(readerFactory ReaderFactory) {
	const dataSize = 100 * 1024 * 1024 // 100MB
	const bufferSize = 32 * 1024       // 32KB classic io.Copy
	const limit = dataSize / 4         // should take 4 seconds
	var elapsed time.Duration

	rf := func(connReader io.ReadCloser) (int, error) {
		rateLimitedReader := readerFactory(connReader, bufferSize, limit)

		var total, n int
		var err error
		buffer := make([]byte, bufferSize)
		start := time.Now()
		for {
			n, err = rateLimitedReader.Read(buffer)
			total += n
			if err != nil {
				if err != io.EOF {
					fmt.Printf("Unexpected error while reading: %v\n", err)
				}
				break
			}
		}

		elapsed = time.Since(start)
		if total != dataSize {
			fmt.Printf("Read incomplete data, read: %d expected: %d\n", n, dataSize)
		}

		return total, err
	}

	wf := func(connWriter io.Writer) (int, error) {
		message := strings.Repeat("A", dataSize)
		return connWriter.Write([]byte(message))
	}

	go func() {
		// give the server a sec to start
		time.Sleep(100 * time.Millisecond)

		n, err := sendTCPMessage(wf)
		if err != nil {
			fmt.Println("Failed to send message:", err)
		}
		if n != dataSize {
			fmt.Printf("Failed to send message: sent insufficient size=%d expectedSize=%d\n", n, dataSize)
		}
	}()

	n, err := receiveOnceTCPServer(rf)
	if err != nil && err != io.EOF {
		fmt.Printf("Unexpected error from server: %v\n", err)
	}
	if n != dataSize {
		fmt.Printf("Failed to get message: got insufficient size=%d expectedSize=%d\n", n, dataSize)
	}

	fmt.Printf("RateLimitingRealWorldLocalTest Took %v\n", elapsed)
}

func RateLimitingRealWorldServerTest(readerFactory ReaderFactory) {
	const dataSize = 100 * 1024 * 1024 // 100MB
	const bufferSize = 32 * 1024       // 32KB classic io.Copy
	const limit = dataSize / 4         // should take 4 seconds
	var elapsed time.Duration

	rf := func(connReader io.ReadCloser) (int, error) {
		rateLimitedReader := readerFactory(connReader, bufferSize, limit)

		var total, n int
		var err error
		buffer := make([]byte, bufferSize)
		start := time.Now()
		for {
			n, err = rateLimitedReader.Read(buffer)
			total += n
			if err != nil {
				if err != io.EOF {
					fmt.Printf("Unexpected error while reading: %v\n", err)
				}
				break
			}
		}

		elapsed = time.Since(start)
		if total != dataSize {
			fmt.Printf("Read incomplete data, read: %d expected: %d\n", n, dataSize)
		}

		return total, err
	}

	n, err := receiveOnceTCPServer(rf)
	if err != nil && err != io.EOF {
		fmt.Printf("Unexpected error from server: %v\n", err)
	}
	if n != dataSize {
		fmt.Printf("Failed to get message: got insufficient size=%d expectedSize=%d\n", n, dataSize)
	}

	fmt.Printf("RateLimitingRealWorldServerTest Took %v\n", elapsed)
}

func SpikeRecoveryRealWorldLocalTest(readerFactory ReaderFactory) {
	const dataSize = 100 * 1024 * 1024 // 100MB
	const bufferSize = 32 * 1024       // 32KB classic io.Copy
	const limit = bufferSize * 500     // should take 6 seconds
	//const a = limit / bufferSize
	//const b = 1000 / a
	//const c = dataSize / limit
	var elapsed time.Duration

	rf := func(connReader io.ReadCloser) (int, error) {
		rateLimitedReader := readerFactory(connReader, bufferSize, limit)

		var total, n int
		var err error
		buffer := make([]byte, bufferSize)
		start := time.Now()
		for {
			n, err = rateLimitedReader.Read(buffer)
			total += n
			if err != nil {
				if err != io.EOF {
					fmt.Printf("Unexpected error while reading: %v\n", err)
				}
				break
			}
		}

		elapsed = time.Since(start)
		if total != dataSize {
			fmt.Printf("Read incomplete data, read: %d expected: %d\n", n, dataSize)
		}

		return total, err
	}

	wf := func(connWriter io.Writer) (int, error) {
		chunkIntervalsMilliseconds := 50
		chunkSize := limit / (1000 / chunkIntervalsMilliseconds)
		spikedChunkSize := chunkSize * 3
		var total int
		var chunkCounter int
		chunkCounterSpikeStart := 1000 / chunkIntervalsMilliseconds     // at 1 second
		chunkCounterSpikeEnd := 3 * (1000 / chunkIntervalsMilliseconds) // at 3 second
		ticker := time.NewTicker(time.Duration(chunkIntervalsMilliseconds) * time.Millisecond)
		defer ticker.Stop()

		for total < dataSize {
			select {
			case <-ticker.C:
				chunkCounter++
				size := chunkSize
				if chunkCounter > chunkCounterSpikeStart && chunkCounter < chunkCounterSpikeEnd {
					size = spikedChunkSize
				}

				message := strings.Repeat("A", size)
				n, err := connWriter.Write([]byte(message))
				total += n
				if err != nil {
					fmt.Printf("Unexpected error while writing: %v\n", err)
					return total, err
				}
			}
		}

		return total, nil
	}

	go func() {
		// give the server a sec to start
		time.Sleep(100 * time.Millisecond)

		n, err := sendTCPMessage(wf)
		if err != nil {
			fmt.Println("Failed to send message:", err)
		}
		if n != dataSize {
			fmt.Printf("Failed to send message: sent insufficient size=%d expectedSize=%d\n", n, dataSize)
		}
	}()

	n, err := receiveOnceTCPServer(rf)
	if err != nil && err != io.EOF {
		fmt.Printf("Unexpected error from server: %v\n", err)
	}
	if n != dataSize {
		fmt.Printf("Failed to get message: got insufficient size=%d expectedSize=%d\n", n, dataSize)
	}

	fmt.Printf("SpikeRecoveryRealWorldLocalTest Took %v\n", elapsed)
}

func SpikeRecoveryRealWorldServerTest(readerFactory ReaderFactory) {
	const dataSize = 100 * 1024 * 1024 // 100MB
	const bufferSize = 32 * 1024       // 32KB classic io.Copy
	const limit = dataSize / 8         // should take 8 seconds
	var elapsed time.Duration

	rf := func(connReader io.ReadCloser) (int, error) {
		rateLimitedReader := readerFactory(connReader, bufferSize, limit)

		var total, n int
		var err error
		buffer := make([]byte, bufferSize)
		start := time.Now()
		for {
			n, err = rateLimitedReader.Read(buffer)
			total += n
			if err != nil {
				if err != io.EOF {
					fmt.Printf("Unexpected error while reading: %v\n", err)
				}
				break
			}
		}

		elapsed = time.Since(start)
		if total != dataSize {
			fmt.Printf("Read incomplete data, read: %d expected: %d\n", n, dataSize)
		}

		return total, err
	}

	n, err := receiveOnceTCPServer(rf)
	if err != nil && err != io.EOF {
		fmt.Printf("Unexpected error from server: %v\n", err)
	}
	if n != dataSize {
		fmt.Printf("Failed to get message: got insufficient size=%d expectedSize=%d\n", n, dataSize)
	}

	fmt.Printf("SpikeRecoveryRealWorldLocalTest Took %v\n", elapsed)
}
