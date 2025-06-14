package main

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"github.com/juju/ratelimit"
)

func main() {
	const chunkSize = 1000 * 1000
	const dataSize = 32 * chunkSize
	reader := bytes.NewBuffer(make([]byte, dataSize))

	// 8 events per second + burst of 8
	// should take (dataSize - burstSize) / limit => (32 - 8) / 8 = 3 seconds
	limit := 8
	limiter := ratelimit.NewBucketWithRate(float64(limit), int64(limit))

	var total int
	buffer := make([]byte, chunkSize)
	start := time.Now()
	for {
		// wait for 1 event to be available (time.Second / limit)
		limiter.Wait(1)

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
