package universe

import (
	"fmt"
	"math"

	"golang.org/x/exp/rand"
	"gonum.org/v1/gonum/stat/distuv"
)

type (
	positive struct {
		rander AnonRander
	}

	namedRander struct {
		AnonRander
		name string
	}

	fixedRander struct {
		value float64
	}

	mixedRander struct {
		p       float64
		rnd     *rand.Rand
		options []Rander
	}

	clampedRander struct {
		rnd      Rander
		min, max float64
	}

	shiftedRander struct {
		Offset float64
		Value  Rander
	}
)

func name(name string, r AnonRander) Rander {
	return namedRander{
		AnonRander: r,
		name:       name,
	}
}

func (n namedRander) Name() string {
	return n.name
}

func Positive(r Rander) Rander {
	ar := r.(namedRander)
	return name(fmt.Sprint("Posi-", ar.name), positive{ar.AnonRander})
}

// func (u *Universe) Mix(choice Rander, options ...Rander) Rander {
// 	dd := toDiscreteDistribution(choice, len(options))
// 	return name("Mixed", mixedRander{dd, u.Random, options})
// }

func (u *Universe) Mix(p float64, options ...Rander) Rander {
	return name("Mixed", mixedRander{p, u.Random, options})
}

func (u *Universe) Clamp(rnd Rander, min, max float64) Rander {
	return name("Clamped", clampedRander{rnd, min, max})
}

func (u *Universe) Fixed(v float64) Rander {
	return u.Uniform(v, v)
}

func (u *Universe) PositiveNormal(mu, sigma float64) Rander {
	return Positive(u.Normal(mu, sigma))
}

func (u *Universe) Normal(mu, sigma float64) Rander {
	return name(fmt.Sprintf("Normal mu=%.1f sigma=%.1f", mu, sigma),
		distuv.Normal{
			Mu:    mu,
			Sigma: sigma,
			Src:   u.Random,
		})
}

func (u *Universe) Uniform(min, max float64) Rander {
	return name(fmt.Sprintf("Uniform min=%.1f max=%.1f", min, max),
		distuv.Uniform{
			Min: min,
			Max: max,
			Src: u.Random,
		})
}

func (u *Universe) Exponential(rate float64) Rander {
	return name(fmt.Sprintf("Exponential rate=%.1f", rate),
		distuv.Exponential{
			Rate: rate,
			Src:  u.Random,
		})
}

func (u *Universe) PositiveLogNormal(mu, sigma float64) Rander {
	return Positive(u.LogNormal(mu, sigma))
}

func (u *Universe) LogNormal(mu, sigma float64) Rander {
	return name(fmt.Sprintf("LogNormal mu=%.1f sigma=%.1f", mu, sigma),
		distuv.LogNormal{
			Mu:    mu,
			Sigma: sigma,
			Src:   u.Random,
		})
}

func (u *Universe) Pareto(xm, alpha float64) Rander {
	return name(fmt.Sprintf("Pareto xm=%.1f alpha=%.1f", xm, alpha),
		distuv.Pareto{
			Xm:    xm,
			Alpha: alpha,
			Src:   u.Random,
		})
}

func (u *Universe) Gamma(alpha, beta float64) Rander {
	return name(fmt.Sprintf("Gamma alpha=%.1f beta=%.1f", alpha, beta),
		distuv.Gamma{
			Alpha: alpha,
			Beta:  beta,
			Src:   u.Random,
		})
}

func (u *Universe) Shift(r Rander, offset float64) Rander {
	return name(fmt.Sprintf("Shift offset=%.1f", offset),
		shiftedRander{
			Offset: offset,
			Value:  r,
		})
}

func (p positive) Rand() float64 {
	for {
		r := p.rander.Rand()
		if r >= 0 {
			return r
		}
	}
}

func (p positive) Prob(x float64) float64 {
	if x < 0 {
		return math.NaN()
	}
	return p.rander.Prob(x)
}

func (m mixedRander) Rand() float64 {
	if m.rnd.Float64() < m.p {
		return m.options[0].Rand()
	}
	return m.options[1].Rand()
}

func (m mixedRander) Prob(x float64) float64 {
	panic("Ha")
	return 0
}

func (c clampedRander) Rand() float64 {
	for {
		x := c.rnd.Rand()
		if x >= c.min && x <= c.max {
			return x
		}
	}
}

func (c clampedRander) Prob(x float64) float64 {
	panic("Ha")
	return 0
}

func (s shiftedRander) Rand() float64 {
	return s.Value.Rand() + s.Offset
}

func (s shiftedRander) Prob(x float64) float64 {
	panic("Ha")
	return 0
}
