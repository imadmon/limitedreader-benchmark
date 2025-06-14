package main

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"go.uber.org/ratelimit"
)

func main() {
	const chunkSize = 1000 * 1000
	const dataSize = 32 * chunkSize
	reader := bytes.NewBuffer(make([]byte, dataSize))

	// 8 events per second, 125ms between calls
	// should take 32 / 8 = 32 * 125ms = 4 seconds
	limit := 8
	limiter := ratelimit.New(limit)

	var total int
	buffer := make([]byte, chunkSize)
	start := time.Now()
	for {
		// wait for 1 event to be available (time.Second / limit)
		limiter.Take()

		n, err := reader.Read(buffer)
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
