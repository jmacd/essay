package num

import (
	"fmt"
	"math"

	"gonum.org/v1/plot"
)

type ClipScale struct {
	Min  float64
	Max  float64
	Norm plot.Normalizer
}

func (cs ClipScale) Normalize(min, max, x float64) float64 {
	min = math.Max(cs.Min, min)
	max = math.Min(cs.Max, max)
	switch {
	case x < cs.Min:
		x = cs.Min
	case x > cs.Max:
		x = cs.Max
	}
	return cs.Norm.Normalize(min, max, x)
}

var _ plot.Normalizer = ClipScale{}

type ClipTicker struct {
	Min    float64
	Max    float64
	Ticker plot.Ticker
}

var _ plot.Ticker = ClipTicker{}

func (ct ClipTicker) Ticks(min, max float64) []plot.Tick {
	min = math.Max(min, ct.Min)
	max = math.Min(max, ct.Max)
	tcks := ct.Ticker.Ticks(min, max)
	for i := range tcks {
		if len(tcks[i].Label) != 0 {
			tcks[i].Label = fmt.Sprintf("%.3g", tcks[i].Value)
		}
	}
	return tcks
}

func MinLogScale(minval float64) plot.Normalizer {
	return ClipScale{
		Min:  minval,
		Max:  math.Inf(+1),
		Norm: plot.LogScale{},
	}
}

func MinLogTicks(minval float64) plot.Ticker {
	return ClipTicker{
		Min:    minval,
		Max:    math.Inf(+1),
		Ticker: plot.LogTicks{},
	}
}
