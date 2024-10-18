package datashape

// datashape implements several algorithms for monitoring the "shape"
// of a data stream in constant space.
import (
	"fmt"
	"math"
)

// See Chiu, Chao, Wang, and Walther, "An improved nonparametric lower
// bound of species richness via a modified good-turing frequency
// formula", International Biometric Society, 2012.
//
//   http://chao.stat.nthu.edu.tw/wordpress/paper/104.pdf
//
// The Chao1 estimator derives from Chao's 1984 inequality, which
// places a lower-bound on the number of species in an *abundance*
// sample, applicable in surveys that aim to compute diversity in a
// population.  An abundance sample is where we have a single sample
// and know the count of each species.
//
// The Chao2 estimator derives from Chao's 1987 inequality, which is
// based on repeated *observed/non-observed* sampling (sometimes
// referred to as "capture/recpature").
//
// In both cases, the estimators are "nonparametric", meaning they
// apply under certain assumptions _without regard to the probability
// distribution_ of frequencies in the population.  Both estimate the
// number of _unobserved_ species in the sample (by clever application
// of the Cauchy-Schwarz inequality).
//
// The Chao2 estimator is easier to analyze, because the frequency of
// observations can be modeled as a binomial distribution (for each
// sample, there is a bernouli trial asking "was species X present?").
// Chao2 is more applicable where sampling is done with replacement.
//
// The Chao1 estimator is based on the observed frequency of
// frequencies f_K in the sample, which count the number of species
// with frequency K present.  Chao1 is based on f_1 and f_2, that is,
// it is computed soley from the number of singleton and doubleton
// species (observed once and twice) present in the sample.  The total
// number of species is at least (there is a modified formula for the
// case where f2 == 0, and the sampled is presumed complete when f_1
// == 0).
//
// The 2012 paper re-presents the Chao1/Chao2 estimators with improved
// clarity, then applies the Good-Turing frequency formula to improve
// bias, i.e., they lead to less undercounting, at the expense of
// increased variance.  The so-called iChao1/iChao2 estimators are
// derived from f_1, f_2, f_3, and f_4.
//
// Presently, we implement Chao1 for abundance data, not iChao1.

type (
	Chao1 struct {
		sampleSize  float64
		estimate    float64
		iestimate   float64
		variance    float64
		speciesSeen float64
		coverage    float64
		sentropy    float64
		pentropy    float64

		F1, F2, F3, F4, FMAX Frequency
	}

	Chao1Stats struct {
		iestimate   statsData
		speciesSeen statsData
		coverage    statsData
		sentropy    statsData
		pentropy    statsData

		F1, F2, F3, F4, FMAX statsData
	}
)

func NewChao1(counts Frequencies) (chao1 Chao1) {
	var (
		F1 Frequency
		F2 Frequency
		F3 Frequency
		F4 Frequency
		N  Frequency
	)
	for _, c := range counts {
		N += c
		if c > chao1.FMAX {
			chao1.FMAX = c
		}
		switch {
		case c == 1:
			F1++
		case c == 2:
			F2++
		case c == 3:
			F3++
		case c == 4:
			F4++
		}
	}
	var (
		n       = float64(N)
		f1      = float64(F1)
		f2      = float64(F2)
		f3      = float64(F3)
		f4      = float64(F4)
		n1      = n - 1
		n1_n    = n1 / n
		n1_n_sq = square(n1_n)
		f1_2    = f1 / f2
		f1_2_sq = square(f1_2)
	)
	chao1.speciesSeen = float64(len(counts))
	chao1.sampleSize = n
	chao1.F1, chao1.F2, chao1.F3, chao1.F4 = F1, F2, F3, F4
	if f2 != 0 {
		// 2b
		chao1.estimate = chao1.speciesSeen + n1_n*f1*f1/f2/2
		// 3a
		chao1.variance = f2 * ((n1_n_sq * square(f1_2_sq) / 4) +
			(n1_n_sq * f1_2_sq * f1_2) +
			(n1_n * f1_2_sq / 2))
		// 7b
		chao1.coverage = 1 - (f1/n)*(n1*f1)/(n1*f1+2*f2)
	} else {
		// 2c
		chao1.estimate = chao1.speciesSeen + n1_n*f1*(f1-1)/2
		// 3b
		chao1.variance = ((f1 * square(2*f1-1) * square(n1) / square(n) / 4) +
			(f1 * (f1 - 1) / 2) +
			(square(square(f1)) / chao1.estimate / 4))
		// 7b
		chao1.coverage = 1 - (f1/n)*(n1*(f1-1))/(n1*(f1-1)+2)
	}

	// 10a
	if f4 == 0 {
		f4 = 1
	}
	chao1.iestimate = chao1.estimate

	opnd := f1 - (n-3)*f2*f3/(2*n1*f4)
	if opnd > 0 {
		chao1.iestimate += f3 * (n - 3) * opnd / (4 * n * f4)
	}

	chao1.pentropy = estimateEntropy(N, F1, F2, counts)

	for _, count := range counts {
		np := float64(count) / n
		chao1.sentropy -= np * math.Log(np)
	}

	return
}

func (c *Chao1) EstimateRichness() float64 {
	return c.iestimate
}

func (c *Chao1) EstimateRichness2() float64 {
	return c.estimate
}

// EstimateCoverageRatio estimates the fraction of the population
// whose types (species) are represented in the sample.
func (c *Chao1) EstimateCoverageRatio() float64 {
	return c.coverage
}

// EstimateObservationRatio estimates the fraction of the species
// count that is represented in the sample.
func (c *Chao1) EstimateObservationRatio() float64 {
	return c.speciesSeen / c.iestimate
}

func (c *Chao1) EstimateUnseenValues() float64 {
	return c.iestimate - c.speciesSeen
}
func (c *Chao1) EstimateUnseenValues2() float64 {
	return c.estimate - c.speciesSeen
}

func (c *Chao1) EstimatePopulationEntropy() float64 {
	return c.pentropy
}

func (c *Chao1) SampleEntropy() float64 {
	return c.sentropy
}

// This is the Chao1 confidence interval formula, not iChao1.
func (c *Chao1) NormalConfidence() (lower, upper float64, err error) {
	var (
		s0 = c.EstimateUnseenValues2()

		R = math.Exp(1.64 * math.Sqrt(1+c.variance/square(s0)))
	)
	return (c.speciesSeen + (s0 / R)), (c.speciesSeen + s0*R), nil
}

func (s *Chao1Stats) add(c Chao1) {
	s.iestimate.add(c.iestimate)
	s.speciesSeen.add(c.speciesSeen)
	s.coverage.add(c.coverage)
	s.sentropy.add(c.sentropy)
	s.pentropy.add(c.pentropy)

	s.F1.add(float64(c.F1))
	s.F2.add(float64(c.F2))
	s.F3.add(float64(c.F3))
	s.F4.add(float64(c.F4))
	s.FMAX.add(float64(c.FMAX))
}

func square(x float64) float64 { return x * x }

func (c Chao1) String() string {
	return fmt.Sprintf("SSize=%v Obs=%.0f Cov=%.3f S0=%.1f H_s=%.3f H_p=%.3f f1..4=%d/%d/%d/%d (sum %d) max %d",
		c.sampleSize, c.speciesSeen, c.coverage, c.EstimateUnseenValues(),
		c.sentropy, c.pentropy, c.F1, c.F2, c.F3, c.F4, c.F1+c.F2+c.F3+c.F4, c.FMAX)
}

func (s Chao1Stats) String() string {
	hs := s.sentropy.Mean()
	hp := s.pentropy.Mean()
	o := s.speciesSeen.Mean()
	e := s.iestimate.Mean()

	return fmt.Sprintf("Ĥ_s=%.3f(%.0f) Ĥ_p=%.3f(%.0f) seen %.0f est %.0f cov %.4f",
		hs, math.Exp(hs), hp, math.Exp(hp),
		o, e, s.coverage.Mean())
}
