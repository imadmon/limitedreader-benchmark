package main

import (
	"context"
	"io"

	"github.com/imadmon/limitedreader"
	jujuratelimit "github.com/juju/ratelimit"
	"go.uber.org/ratelimit"
	"golang.org/x/time/rate"
)

type ReaderFactory func(reader io.ReadCloser, bufferSize, limit int) io.ReadCloser

type ReaderType string

var (
	GolangReader  ReaderType = "Golang"
	JujuReader    ReaderType = "Juju"
	UberReader    ReaderType = "Uber"
	IMadmonReader ReaderType = "IMadmon"
)

func IMadmonDeterministicRateLimitReaderFactory(reader io.ReadCloser, bufferSize, limit int) io.ReadCloser {
	return limitedreader.NewRateLimitedReadCloser(reader, int64(limit))
}

func GolangBurstsRateLimitReaderFactory(reader io.ReadCloser, bufferSize, limit int) io.ReadCloser {
	// limiter := rate.NewLimiter(rate.Every(time.Second/time.Duration(limit/bufferSize)), 1)
	// limiter := rate.NewLimiter(rate.Limit(limit/bufferSize), 1)
	limiter := rate.NewLimiter(rate.Limit(limit/bufferSize), limit/bufferSize)
	return &GolangRateLimitedReader{
		reader:  reader,
		limiter: limiter,
		ctx:     context.Background(),
	}
}

type GolangRateLimitedReader struct {
	reader  io.ReadCloser
	limiter *rate.Limiter
	ctx     context.Context
}

func (r *GolangRateLimitedReader) Read(p []byte) (n int, err error) {
	err = r.limiter.Wait(r.ctx) // wait until tokens are available
	if err != nil {
		return 0, err
	}
	return r.reader.Read(p)
}

func (r *GolangRateLimitedReader) Close() error {
	return r.reader.Close()
}

func JujuBurstsRateLimitReaderFactory(reader io.ReadCloser, _, limit int) io.ReadCloser {
	bucket := jujuratelimit.NewBucketWithRate(float64(limit), int64(limit))
	return &JujuRateLimitedReader{
		reader: reader,
		bucket: bucket,
	}
}

type JujuRateLimitedReader struct {
	reader io.ReadCloser
	bucket *jujuratelimit.Bucket
}

func (r *JujuRateLimitedReader) Read(p []byte) (n int, err error) {
	n, err = r.reader.Read(p)
	r.bucket.Wait(int64(n))
	return n, err
}

func (r *JujuRateLimitedReader) Close() error {
	return r.reader.Close()
}

func UberDeterministicRateLimitReaderFactory(reader io.ReadCloser, bufferSize, limit int) io.ReadCloser {
	rl := ratelimit.New(limit / bufferSize) // operations per second
	return &UberRateLimitedReader{
		reader:  reader,
		limiter: rl,
	}
}

type UberRateLimitedReader struct {
	reader  io.ReadCloser
	limiter ratelimit.Limiter
}

func (r *UberRateLimitedReader) Read(p []byte) (n int, err error) {
	r.limiter.Take()
	return r.reader.Read(p)
}

func (r *UberRateLimitedReader) Close() error {
	return r.reader.Close()
}
