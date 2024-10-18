package datashape

import (
	"errors"
	"math"
)

// Journal of Plant Ecology VOLUME 5, NUMBER 1, PAGES 3–21 MARCH 2012
// doi: 10.1093/jpe/rtr044 available online at
// www.jpe.oxfordjournals.org
//
// Models and estimators linking individual-based and sample-based
// rarefaction, extrapolation and comparison of assemblages. Robert
// K. Colwell, Anne Chao, Nicholas J. Gotelli, Shang-Yi Lin, Chang
// Xuan Mao, Robin L. Chazdon and John T. Longino
//
// Section "Individual-Based Extrapolation, the Multinomial Model
// http://chao.stat.nthu.edu.tw/wordpress/paper/90.pdf
//
func (c *Chao1) Extrapolate(g float64) (totalSpecies, additionalIndividuals float64, err error) {
	sEst := c.EstimateObservationRatio()
	if g < sEst || g >= 1 {
		return 0, 0, errors.New("g out of range")
	}
	var (
		f0 = c.EstimateUnseenValues()
		G  = 1 - g
		a  = math.Log(f0 / (G * sEst))
		b  = c.sampleSize * float64(c.F1) / float64(2*c.F2)
		m  = a * b
	)
	if m == math.Inf(+1) {
		m = c.sampleSize
	}
	if m < 1 {
		m = 1
	}
	var (
		p = m * float64(c.F1) / (c.sampleSize * f0)
		q = 1 - math.Exp(-p)
		r = c.speciesSeen + f0*q
	)
	return r, m, nil
}

// Predicting the Number of New Species in Further Taxonomic Sampling
// Tsung-Jen Shen, Anne Chao, and Chih-Feng Lin
// http://chao.stat.nthu.edu.tw/wordpress/paper/2003_Ecology_84_P798.pdf
func (c *Chao1) PredictFurther(m float64) float64 {
	unseen := c.EstimateUnseenValues()
	a := 1 - (1-c.EstimateCoverageRatio())/unseen
	b := math.Pow(a, m)
	return unseen * (1 - b)
}
