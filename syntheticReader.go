package main

import "io"

type syntheticReader struct {
	size  uint64
	total uint64
}

func (r *syntheticReader) Read(p []byte) (n int, err error) {
	n = len(p)
	uint64N := uint64(n)
	if r.size > 0 && r.total+uint64N >= r.size {
		uint64N = r.size - r.total
		n = int(uint64N)
		err = io.EOF
	}

	for i := 0; i < n; i++ {
		p[i] = 'A'
	}

	SyntheticRXBytes.Add(uint64N)
	r.total += uint64N
	return
}

func (r *syntheticReader) Close() error {
	return nil
}
