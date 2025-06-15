package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/samber/lo"
)

type LineSeriesData struct {
	Title  string
	Values []float32
	Color  string
}

func GraphBenchmark(benchmark AllBenchmarkData, filename string) {
	graphs := make([]*charts.Line, 0)

	graphs = append(graphs, BenchmarkRateLimitingSyntheticGraph(benchmark[BenchmarkRateLimitingSynthetic]))
	graphs = append(graphs, BenchmarkRateLimitingRealWorldLocalGraph(benchmark[BenchmarkRateLimitingRealWorldLocal])...)
	graphs = append(graphs, BenchmarkMaxReadOverTimeSyntheticGraph(benchmark[BenchmarkMaxReadOverTimeSynthetic])...)
	graphs = append(graphs, BenchmarkSpikeRecoveryRealWorldLocalGraph(benchmark[BenchmarkSpikeRecoveryRealWorldLocal])...)

	WriteGraphsToFile(graphs, filename)
}

func BenchmarkRateLimitingSyntheticGraph(data BenchmarkData) *charts.Line {
	return GenerateGraphChart(
		"Classic Usage Synthetic Rate Limiting - SyntheticRX MB",
		"Passing X data with X/4 limit with synthetic reader",
		nil,
		MoveOverlappingSeriesData([]SeriesData{
			data[GolangReader][SyntheticRX],
			data[JujuReader][SyntheticRX],
			data[UberReader][SyntheticRX],
			data[IMadmonReader][SyntheticRX],
		}),
	)
}

func BenchmarkRateLimitingRealWorldLocalGraph(data BenchmarkData) []*charts.Line {
	title := "Real-World Rate Limiting"
	subtitle := "Passing X data with X/4 limit between 2 servers"
	return []*charts.Line{
		GenerateGraphChart(
			title+" - RX MB",
			subtitle,
			nil,
			MoveOverlappingSeriesData([]SeriesData{
				data[GolangReader][RX],
				data[JujuReader][RX],
				data[UberReader][RX],
				data[IMadmonReader][RX],
			}),
		),
		GenerateGraphChart(
			title+" - CPU Usage",
			subtitle,
			nil,
			MoveOverlappingSeriesData([]SeriesData{
				data[GolangReader][CPU],
				data[JujuReader][CPU],
				data[UberReader][CPU],
				data[IMadmonReader][CPU],
			}),
		),
		GenerateGraphChart(
			title+" - RAM MB Usage",
			subtitle,
			nil,
			MoveOverlappingSeriesData([]SeriesData{
				data[GolangReader][RAM],
				data[JujuReader][RAM],
				data[UberReader][RAM],
				data[IMadmonReader][RAM],
			}),
		),
	}
}

func BenchmarkMaxReadOverTimeSyntheticGraph(data BenchmarkData) []*charts.Line {
	title := "Max Read Over 10 Seconds"
	subtitle := "Passing infinite data with no limit with synthetic reader"
	return []*charts.Line{
		GenerateGraphChart(
			title+" - Total SyntheticRX MB",
			subtitle,
			nil,
			MoveOverlappingSeriesData([]SeriesData{
				data[GolangReader][TotalSyntheticRX],
				data[JujuReader][TotalSyntheticRX],
				data[UberReader][TotalSyntheticRX],
				data[IMadmonReader][TotalSyntheticRX],
			}),
		),
		GenerateGraphChart(
			title+" - CPU Usage",
			subtitle,
			nil,
			MoveOverlappingSeriesData([]SeriesData{
				data[GolangReader][CPU],
				data[JujuReader][CPU],
				data[UberReader][CPU],
				data[IMadmonReader][CPU],
			}),
		),
	}
}

func BenchmarkSpikeRecoveryRealWorldLocalGraph(data BenchmarkData) []*charts.Line {
	title := "Real-World Spike Recovery"
	subtitle := "Rate limit between 2 servers with a spike after 1 second"
	markLines := map[string]float64{
		"Spike Start": 1.0,
		"Spike End":   3.0,
	}
	return []*charts.Line{
		GenerateGraphChart(
			title+" - RX MB",
			subtitle,
			markLines,
			MoveOverlappingSeriesData([]SeriesData{
				data[GolangReader][RX],
				data[JujuReader][RX],
				data[UberReader][RX],
				data[IMadmonReader][RX],
			}),
		),
		GenerateGraphChart(
			title+" - CPU Usage",
			subtitle,
			markLines,
			MoveOverlappingSeriesData([]SeriesData{
				data[GolangReader][CPU],
				data[JujuReader][CPU],
				data[UberReader][CPU],
				data[IMadmonReader][CPU],
			}),
		),
		GenerateGraphChart(
			title+" - RAM MB Usage",
			subtitle,
			markLines,
			MoveOverlappingSeriesData([]SeriesData{
				data[GolangReader][RAM],
				data[JujuReader][RAM],
				data[UberReader][RAM],
				data[IMadmonReader][RAM],
			}),
		),
	}
}

func WriteGraphsToFile(graphs []*charts.Line, graphFileName string) {
	f, _ := os.Create(graphFileName)
	defer f.Close()
	for _, graph := range graphs {
		graph.Render(f)
	}
	fmt.Printf("Graph rendered at %s\n", graphFileName)
}

func StartGraphSeriesMonitor(seriesName, color string, seriesValueType MonitorValueType, stopC chan struct{}) SeriesData {
	ctx, ctxCancel := context.WithCancel(context.Background())
	resultsC := make(chan []monitorResult)

	go monitorLoop(ctx, resultsC)
	time.Sleep(300 * time.Millisecond)
	<-stopC
	time.Sleep(700 * time.Millisecond)
	ctxCancel()
	time.Sleep(300 * time.Millisecond)

	results := <-resultsC
	return SeriesData{
		Title:  seriesName,
		Values: parseGraphValue(results, seriesValueType),
		Color:  color,
	}
}

func GenerateGraphChart(title, subtitle string, markLines map[string]float64, series []LineSeriesData) *charts.Line {
	graph := charts.NewLine()
	graph.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title:    title,
			Subtitle: subtitle,
		}),
		charts.WithLegendOpts(opts.Legend{
			Left:  "right",
			Top:   "top",
			Align: "auto",
		}),
		charts.WithGridOpts(opts.Grid{
			Top: "80px",
		}),
	)

	axisSize := len(lo.MaxBy(series, func(a, b LineSeriesData) bool { return len(a.Values) >= len(b.Values) }).Values)
	var xAxis []string
	for i := 0.0; i < float64(axisSize/5); i += 0.2 {
		xAxis = append(xAxis, fmt.Sprintf("%.1f", i))
	}

	graph.SetXAxis(xAxis)

	for _, s := range series {
		items := lo.Map(s.Values, func(value float32, _ int) opts.LineData { return opts.LineData{Value: value} })
		graph.AddSeries(s.Title, items,
			charts.WithLineStyleOpts(opts.LineStyle{
				// Width: s.Width,
				Color: s.Color,
			}),
			charts.WithItemStyleOpts(opts.ItemStyle{
				Color: s.Color,
			}),
			charts.WithLineChartOpts(opts.LineChart{
				//Smooth:       opts.Bool(true),
				//ConnectNulls: opts.Bool(true),
				SymbolSize: 6,
				//Symbol:     "circle", //  'circle', 'rect', 'roundRect', 'triangle', 'diamond', 'pin', 'arrow', 'none'
			}),
			//charts.WithLabelOpts(opts.Label{
			//	Show: opts.Bool(true),
			//}),
			//charts.WithAreaStyleOpts(opts.AreaStyle{
			//	Opacity: 0.2,
			//}),
		)
	}

	for markTitle, markDim := range markLines {
		graph.SetSeriesOptions(
			charts.WithMarkLineStyleOpts(opts.MarkLineStyle{
				Symbol:     []string{"none", "none"},
				SymbolSize: 0,
				Label: &opts.Label{
					Formatter: "{b}",
					Color:     "#707070", // "#fc8452",
				},
				LineStyle: &opts.LineStyle{
					Color: "#707070", // "#fc8452",
					Width: 1,
					Type:  "dashed", // "solid", "dashed", "dotted".
				},
			}),
			charts.WithMarkLineNameXAxisItemOpts(opts.MarkLineNameXAxisItem{
				Name:     markTitle,
				XAxis:    fmt.Sprintf("%.1f", markDim),
				ValueDim: "x",
			}),
		)
	}

	return graph
}

func parseGraphValue(values []monitorResult, valueType MonitorValueType) []int {
	mb := 1024 * 1024
	return lo.Map(values, func(item monitorResult, _ int) int {
		switch valueType {
		case RX:
			return int(item.rxDelta) / mb
		case SyntheticRX:
			return int(item.syntheticRXDelta) / mb
		case TotalSyntheticRX:
			return int(item.totalSyntheticRX) / mb
		case CPU:
			return int(item.cpuPercent)
		case RAM:
			return int(item.ramMB)
		default:
			return 0
		}
	})
}

func MoveOverlappingSeriesData(values []SeriesData) []LineSeriesData {
	maxValue := 0
	for _, v := range values {
		currMaxValue := lo.Max(v.Values)
		if currMaxValue > maxValue {
			maxValue = currMaxValue
		}
	}

	yAxisSize := float32(maxValue) * 1.1
	deviation := yAxisSize / 150

	result := make([]LineSeriesData, 0)
	for i, v := range values {
		result = append(result, LineSeriesData{
			Title:  v.Title,
			Color:  v.Color,
			Values: lo.Map(v.Values, func(item, _ int) float32 { return float32(item) + (float32(i) * deviation) }),
		})
	}

	return result
}
