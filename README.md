# 📊 Go Rate Limiter Benchmark Suite

Welcome to the official benchmark suite used to evaluate and compare different Go rate limiting reader libraries — including 
[`golang.org/x/time/rate`](https://pkg.go.dev/golang.org/x/time/rate), 
[`uber-go/ratelimit`](https://github.com/uber-go/ratelimit), 
[`juju/ratelimit`](https://github.com/juju/ratelimit) and 
[`imadmon/limitedreader`](https://github.com/imadmon/limitedreader).  
This project runs various real-world and synthetic tests and outputs data for visual benchmarking (e.g., RX throughput, CPU usage, RAM usage) in **200ms intervals**.

> 🚀 Built to support the article: *"Burst vs Deterministic Rate Limiting in Go for Real-time Systems"*

</br>


## ✨ What This Repo Includes

- 📦 Benchmark framework using real and synthetic data streams
- 📈 Output data suitable for graphing and visual analysis
- 🔍 Multiple rate limiting strategies tested under the same conditions
- 🧠 Designed to uncover behaviors like burst handling, spike recovery, and deterministic consistency

</br>


## 🧪 Benchmark Scenarios

Each benchmark outputs to a graph with metrics like RX bytes, CPU percentage, RAM usage — sampled every 200ms.

| Test Name            | Description                                                                 |
|----------------------|-----------------------------------------------------------------------------|
| **BasicRateLimit**   | Classic case — stream with 1/4 throttle rate, measuring RX over time         |
| **RealStreamLimit**  | Actual TCP stream between two servers under rate limit                       |
| **MaxReadSpeed**     | Limit set to "unlimited", tests raw read throughput capacity                 |
| **SpikeRecovery**    | Test spike handling: data burst midstream, then return to steady rate        |

</br>


## 🔧 Usage

Clone and run:

```bash
git clone https://github.com/imadmon/limitedreader-benchmark
cd limitedreader-benchmark
go mod tidy
go build && ./limitedreader-benchmark
```

</br>


## 🚩 Benchmark Results

A full visualizaation of all benchmark results


### Benchmark Graphs
> 👉 [View Interactive Benchmark Graphs](https://imadmon.github.io/limitedreader-benchmark/benchmark.html)


### Libraries Usages Graphs
> 👉 [View Interactive Usages Graphs](https://imadmon.github.io/limitedreader-benchmark/usage.html)
