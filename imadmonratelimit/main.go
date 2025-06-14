package main

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"github.com/imadmon/limitedreader"
)

func main() {
	const chunkSize = 1000 * 1000
	dataSize := 32 * chunkSize
	reader := bytes.NewBuffer(make([]byte, dataSize))

	// reading 8 chunkSize will take a second
	// should take 32 / 8 = 4 seconds
	// reads interval divided evenly by limitedreader.ReadIntervalMilliseconds
	limit := 8 * chunkSize
	limitedReader := limitedreader.NewRateLimitedReader(reader, int64(limit))

	var total int
	buffer := make([]byte, chunkSize)
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
	fmt.Printf("Total: %d, Elapsed: %s\n", total, elapsed)
}
