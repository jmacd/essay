package universe

import (
	"fmt"

	ms "github.com/jmacd/essay/examples/internal/multishape"
)

type (
	alphaAssigner struct {
		u  *Universe
		mv MultiVariate
	}
)

// Returns a 1D timeseries of values with average "rate"
// (events/second) spanning "duration" (seconds).
func NewContinuousTimeseries(u *Universe, v ms.Variable, r Rander, rate, duration float64) MultiVariate {
	lp := u.NewLeafProfile()

	nv := u.NewNumerical(r)
	po := u.NewPoisson(rate, StartTime, StartTime+duration)

	lp.Add(Time, po)
	lp.Add(v, nv)

	return u.randomize(lp)
}

func NewTimeseries(u *Universe, rate, duration float64, vs ...NamedVariate) MultiVariate {
	lp := u.NewLeafProfile()

	po := u.NewPoisson(rate, StartTime, StartTime+duration)
	lp.Add(Time, po)

	for _, v := range vs {
		lp.Add(v.Name, v.V)
	}

	return u.randomize(lp)
}

func NewVariableTimeseries(u *Universe, vs ...NamedVariate) MultiVariate {
	lp := u.NewLeafProfile()

	for _, v := range vs {
		lp.Add(v.Name, v.V)
	}

	return u.randomize(lp)
}

func (u *Universe) Generate(mv MultiVariate) MultiVariate {
	return u.randomize(mv)
}

func (u *Universe) randomize(mv MultiVariate) MultiVariate {
	return alphaAssigner{
		u:  u,
		mv: mv,
	}
}

func (aa alphaAssigner) Produce(ind *ms.Individual) error {
	if err := aa.mv.Produce(ind); err != nil {
		return err
	}
	ind.SetAlpha(aa.u.Random.Float64())
	ind.SetWeight(1)
	return nil
}

func MakePopulation(assembly MultiVariate) (points ms.Population) {
	for {
		var pt ms.Individual
		if err := assembly.Produce(&pt); err != nil {
			if err != NotSupported {
				panic(fmt.Sprint("Invalid error: ", err))
			}
			break
		}
		points = append(points, pt)
	}
	return
}
