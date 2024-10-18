package tdigest

import (
	"testing"

	ms "github.com/jmacd/essay/examples/internal/multishape"
	"github.com/jmacd/essay/examples/internal/multishape/universe"
)

const (
	rate     = 1000
	duration = 10
)

func TestBasic(t *testing.T) {
	u := universe.New("testing", 123)

	basic := universe.NewTest_1D_Lat(u, rate, duration)
	var points []ms.Individual
	for {
		var pt ms.Individual
		if err := basic.Produce(&pt); err != nil {
			break

		}
		points = append(points, pt)
	}

	var latency = ms.Variable("latency")
	var lats []float64
	var err error

	if lats, err = latency.Numbers(points); err != nil {
		t.Fatal("Invalid number generator")
	}

	comp := Quality(10)

	scale := comp.NewScale(lats, Ones(len(lats)))

	lastq := 0.0
	scale.Quantiles(func(ctrd Centroid) {
		if ctrd.quantile > 1 || ctrd.quantile <= lastq {
			t.Error("Impossible quantile", ctrd)
		}
		lastq = ctrd.quantile
	})
	if lastq != 1 {
		t.Error("Wrong quantile sumn", lastq)
	}
}
