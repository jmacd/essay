package models

// Model2 operates on a single categorical variable, introduces the
// Chao corrections for unknown species.  Note this uses SVD to
// compute Chi^2 (an unnecessary expense).
//
// Uses
// http://chao.stat.nthu.edu.tw/wordpress/paper/99_pdf_appendix.pdf
// Appendex S2 has a single-parameter (λ) adjustement to the species
// probability in which the rare element probabilities are reduced
// to account for all the uncovered rare elements that were unseen.
//
// See also
// http://chao.stat.nthu.edu.tw/wordpress/paper/102_pdf_appendix.pdf

import (
	"fmt"
	"math"

	"github.com/jmacd/essay/examples/internal/datashape"
	ms "github.com/jmacd/essay/examples/internal/multishape"
	"github.com/jmacd/essay/examples/internal/multishape/frequency"
	"github.com/jmacd/essay/examples/internal/multishape/sampler"
	"gonum.org/v1/gonum/mat"
)

const (
	perfectCoverage = 0.999

	UnknownCategory = ms.Category("__unknown__")
)

type model2 struct {
	colorVar ms.Variable
}

func NewModel2(colorVar ms.Variable) sampler.Model {
	return model2{colorVar: colorVar}
}

func (m2 model2) Update(inputSampled ms.Population, prior sampler.Scale) sampler.Scale {
	inputSampled = inputSampled.Copy()
	colormapInput := inputSampled.Categorize(m2.colorVar).Indexed

	numColors := len(colormapInput)
	numPoints := len(inputSampled)

	// Treat columns as rows, points as columns.  Add one row for
	// the unknown color column and row.
	numRows := numColors + 1
	numColumns := numPoints + 1

	// Count unweighted freqencies, estimate the sample coverage
	// and number of unseen types.
	unitfreqs := datashape.Frequencies{}

	W := 0.0

	for _, pt := range inputSampled {
		unitfreqs[pt.Category(m2.colorVar)]++

		W += pt.Weight()
	}

	// missingAdj is a multiplier, includes the unknown value
	missingAdj, unknownProb, unknownColors := m2.missingAdjustment(inputSampled, unitfreqs, prior)

	totalUnknownWeightCurrent := W * unknownProb

	effectiveProbs := map[ms.Category]float64{}

	// Apply missing-values adjustment.
	renorm := totalUnknownWeightCurrent
	for idx, pt := range inputSampled {
		color := pt.Category(m2.colorVar)
		output := pt.Weight() * missingAdj[color]

		inputSampled[idx].SetWeight(output)
		renorm += output
	}
	renormFactor := W / renorm
	for idx := range inputSampled {
		inputSampled[idx].SetWeight(inputSampled[idx].Weight() * renormFactor)

		color := inputSampled[idx].Category(m2.colorVar)
		effectiveProbs[color] += inputSampled[idx].Weight() / W
	}

	// // DEBUG
	// sumP := 0.0
	// for k, v := range effectiveProbs {
	// 	// fmt.Println("Class", k, "prob", v)
	// 	sumP += v
	// }
	// // fmt.Println("Known prob sum", sumP, "unknown", unknownProb, "total", sumP+unknownProb)

	// compute a MAP estimate, the incorporating unknown estimate
	if priorScale, _ := prior.(*frequency.MultiVarScale); priorScale != nil {
		// Compute real vs estimated unknown count
		commonCategories := 0

		for color, _ := range colormapInput {
			if _, ok := priorScale.Map[color]; ok {
				commonCategories++
			}
		}

		// fmt.Println("prior est Us", priorScale.EstUnknowns)
		// fmt.Println("curnt est Us", unknownColors)

		// The number of values that are not represented in the other.
		priorKnownCurrentUnknowns := priorScale.Knowns - commonCategories
		currentKnownPriorUnknowns := len(colormapInput) - commonCategories

		// fmt.Println("prior KUs", priorKnownCurrentUnknowns, "prior Ks", priorScale.Knowns, "/", commonCategories)
		// fmt.Println("curnt KUs", currentKnownPriorUnknowns, "curnt Ks", len(colormapInput), "/", commonCategories)

		// The number of effective unknowns is the max of the
		// estimated unknowns and the known unknowns plus
		// 1--meaning there's always at least one true
		// unknown, even when the number of known unknowns
		// exceeds the estimated unknowns.
		priorEffectiveUnknowns := math.Max(priorScale.EstUnknowns, float64(currentKnownPriorUnknowns+1))
		currentEffectiveUnknowns := math.Max(unknownColors, float64(priorKnownCurrentUnknowns+1))

		// fmt.Println("prior eff Us", priorEffectiveUnknowns)
		// fmt.Println("curnt eff Us", currentEffectiveUnknowns)

		// The total weight of unknowns.
		totalUnknownWeightPrior := priorScale.TotalWeight * priorScale.UnknownProb

		// fmt.Println("prior U wt total", totalUnknownWeightPrior)
		// fmt.Println("curnt U wt total", totalUnknownWeightCurrent)

		// The weight of known unknowns
		weightPerUnknownPrior := totalUnknownWeightPrior / priorEffectiveUnknowns
		weightPerUnknownCurrent := totalUnknownWeightCurrent / currentEffectiveUnknowns

		// fmt.Println("prior w per U", weightPerUnknownPrior)
		// fmt.Println("curnt w per U", weightPerUnknownCurrent)

		// The sum weight of unknown unknowns, subtracting known unknowns.
		//
		// Note: prior deduction is determined by current knnown unknowns,
		//       current deduction is determined by prior knnown unknowns.
		weightUnknownSumPrior := totalUnknownWeightPrior - (weightPerUnknownPrior * float64(currentKnownPriorUnknowns))
		weightUnknownSumCurrent := totalUnknownWeightCurrent - (weightPerUnknownCurrent * float64(priorKnownCurrentUnknowns))

		// fmt.Println("prior wU sum", weightUnknownSumPrior)
		// fmt.Println("curnt wU sum", weightUnknownSumCurrent)

		// Compute the effective weight of each item, following the missing correction, for current and prior.
		currentClassWeight := map[ms.Category]float64{}
		for key, prob := range effectiveProbs {
			currentClassWeight[key] = prob * W
		}
		priorClassWeight := map[ms.Category]float64{}
		for key, prob := range priorScale.Probs {
			priorClassWeight[key] = prob * priorScale.TotalWeight
		}
		// fmt.Println("len(PCW)", len(priorClassWeight))
		// fmt.Println("len(CCW)", len(currentClassWeight))

		// Compute a MAP estimate of the categorical distribution.
		//
		// effK is "K" in the categorical distribution, which is
		effK := len(colormapInput) + priorKnownCurrentUnknowns + 1
		denom := W + priorScale.TotalWeight + float64(effK)

		numerSum := 0.0
		numerSumP := 0.0
		numerSumC := 0.0
		numerCount := 0
		sumCW := 0.0
		sumPW := 0.0

		// Compute current classes
		for curClass, cw := range currentClassWeight {
			if pw, ok := priorClassWeight[curClass]; ok {
				// Both have known weight
				numerSum += cw + pw + 1
				numerSumP += pw
				numerSumC += cw
				sumCW += cw
				sumPW += pw
				numerCount++
				effectiveProbs[curClass] = (cw + pw + 1) / denom
				continue
			}

			// Current known, prior known unknown.
			numerSum += cw + weightPerUnknownPrior + 1
			numerSumP += weightPerUnknownPrior
			numerSumC += cw
			effectiveProbs[curClass] = (cw + weightPerUnknownPrior + 1) / denom
			numerCount++
			continue
		}

		unknownProb = 0

		// Compute Theta for prior classes
		for priorClass, pw := range priorClassWeight {
			if _, ok := currentClassWeight[priorClass]; ok {
				// Handled above
				continue
			}

			// Prior known, current known unknown.
			numerSum += pw + weightPerUnknownCurrent + 1
			numerSumP += pw
			numerSumC += weightPerUnknownCurrent
			unknownProb += (pw + weightPerUnknownCurrent + 1) / denom
			numerCount++
			continue
		}

		numerSum += weightUnknownSumCurrent + weightUnknownSumPrior + 1
		numerSumP += weightUnknownSumPrior
		numerSumC += weightUnknownSumCurrent
		numerCount++
		unknownProb += (weightUnknownSumCurrent + weightUnknownSumPrior + 1) / denom
	}

	for idx := range inputSampled {
		color := inputSampled[idx].Category(m2.colorVar)
		inputSampled[idx].SetWeight(W * effectiveProbs[color] / float64(unitfreqs[color]))
	}

	// Construct the indicator matrix X.
	X := mat.NewDense(numRows, numColumns, nil)

	// Fill in the indicator matrix.
	for i, p := range inputSampled {
		color := p.Category(m2.colorVar)
		colorCol, ok := colormapInput[color]

		if !ok {
			panic("Color not found")
		}

		X.Set(colorCol, i, p.Weight())
	}

	X.Set(numColors, numPoints, unknownProb*W)

	// fmt.Printf("Adjusted indicator matrix\n%.2f\n", mat.Formatted(X, mat.Squeeze()))

	// Compute the probability matrix.
	Z := mat.NewDense(numRows, numColumns, nil)
	Z.Scale(1/W, X)

	// Compute row profile (numRows rows, 1 column).
	r := mat.NewDense(numRows, 1, nil)
	r.Mul(Z, constMat(numColumns, 1, 1))

	// Compute column profile (1 row, numColumns columns).
	c := mat.NewDense(1, numColumns, nil)
	c.Mul(constMat(1, numRows, 1), Z)

	// Diagonalize r and c.
	Dr := diagonalize(r.ColView(0))
	Dc := diagonalize(c.RowView(0))

	invSqrtDr := mat.NewDense(numRows, numRows, nil)
	invSqrtDr.Apply(invSqrt, Dr)

	invSqrtDc := mat.NewDense(numColumns, numColumns, nil)
	invSqrtDc.Apply(invSqrt, Dc)

	// Compute r * cT (N.B. 'c' is already transposed)
	rc := mat.NewDense(numRows, numColumns, nil)
	rc.Mul(r, c)

	// Compute (Z - rcT)
	centered := mat.NewDense(numRows, numColumns, nil)
	centered.Sub(Z, rc)

	// Compute the MCA product matrix (`S` in Greenacre)
	product := mat.NewDense(numRows, numColumns, nil)
	product.Mul(invSqrtDr, centered)
	product.Mul(product, invSqrtDc)

	// Note: we do not need to compute the SVD to calculate Chi^2
	// distance ratios, however this tells us something (e.g.,
	// Covariance matrix for keys) about the data distribution and
	// it should compute the same Chi^2 distance ratios.
	var svd mat.SVD
	if !svd.Factorize(product, mat.SVDThin) {
		panic("Could not factorize product matrix")
	}

	// See how many values are significant, i.e., those > 1/numColumns
	// N.B. Assumes Rows > Cols.
	K := float64(numRows)
	threshold := 1 / K
	values := svd.Values(nil)

	for i, v := range values {
		if v > threshold {
			continue
		}
		values = values[:i]
		break
	}

	numComponents := len(values)

	Values := mat.NewDiagDense(numComponents, values)

	var U mat.Dense
	svd.UTo(&U)

	// Factor scores for the rows.
	USlice := U.Slice(0, numRows, 0, numComponents)
	F := mat.NewDense(numRows, numComponents, nil)
	F.Mul(invSqrtDr, USlice)
	F.Mul(F, Values)

	// Corrected eigenvalues
	corrected := mat.NewDiagDense(numComponents, nil)
	normsum := 0.
	for i, v := range values {
		var (
			kk1 = K / (K - 1)
			l1k = v - (1 / K)
			kp  = kk1 * l1k
			kk2 = kp * kp
		)

		corrected.SetDiag(i, kk2)
		normsum += kk2
	}
	normalized := mat.NewDiagDense(numComponents, nil)
	for i := range values {
		normalized.SetDiag(i, corrected.At(i, i)/normsum)
	}

	// Chi-square for input rows:
	rowTotal := 0.0
	scaleWts := make([]float64, numRows)

	for row := range scaleWts {
		frow := F.Slice(row, row+1, 0, numComponents)

		c2 := mat.NewDense(1, 1, nil)
		c2.Mul(frow, frow.T())

		scaleWts[row] = c2.At(0, 0)
		rowTotal += c2.At(0, 0)
	}

	for row := range scaleWts {
		scaleWts[row] = scaleWts[row] / rowTotal
	}

	sumCol := 0.0
	for row := 0; row < numRows; row++ {
		sumCol += scaleWts[row]
	}

	newScale := frequency.NewMultiVarScale(inputSampled).Add(
		func(pt ms.Individual) float64 {
			cval := pt.Category(m2.colorVar)
			colorNum, ok := colormapInput[cval]
			if !ok {
				return scaleWts[numColors]
			}
			return scaleWts[colorNum]
		})

	newScale.Knowns = len(colormapInput)
	newScale.EstUnknowns = unknownColors
	newScale.TotalWeight = W
	newScale.UnknownProb = unknownProb
	newScale.Map = colormapInput
	newScale.Probs = effectiveProbs

	return newScale
}

// reduction fraction, will be relatively smaller for rare elements
// and nearly 1.0 for the high frequencies.
func (m2 model2) missingAdjustment(inputSampled ms.Population, units datashape.Frequencies, prior sampler.Scale) (map[ms.Category]float64, float64, float64) {
	// Estimate sample and species coverage probabilities.
	chao := datashape.NewChao1(units)
	sampleCoverage := chao.EstimateCoverageRatio() // "C-hat"
	unseenColors := chao.EstimateUnseenValues()
	numColors := len(units)
	numPoints := len(inputSampled)
	speciesCoverage := float64(numColors) / (float64(numColors) + unseenColors) // "S-hat"

	if sampleCoverage > perfectCoverage {
		sampleCoverage = perfectCoverage
	}

	fmt.Println("Num colors:", numColors, "total points:", numPoints,
		"est unseen/total colors:", unseenColors, "/", unseenColors/(1-speciesCoverage),
		"coverage", sampleCoverage)

	// Compute unit frequencies
	sum := 0.0
	for _, c := range units {
		sum += float64(c)
	}
	denom := 0.0
	for _, c := range units {
		p := float64(c) / sum
		q := 1 - p
		denom += p * math.Pow(q, sum) // "lambda-hat i"
	}
	lambda := (1 - sampleCoverage) / denom

	// Unknown adjustment
	adjustment := map[ms.Category]float64{}
	uprob := 0.0
	for spec, c := range units {
		cat := spec.(ms.Category)
		p := float64(c) / sum
		q := 1 - p

		lpow := lambda * math.Pow(q, sum)
		padj := p * (1 - lpow) // "p-hat i"
		uprob += p * lpow

		adjustment[cat] = padj / p

		if adjustment[cat] >= 1.001 || adjustment[cat] <= 0 {
			panic(fmt.Sprintln("impossible prob", p, "adjustment", padj,
				"by", cat, "computed_p", adjustment[cat],
				"count", c, "sum", sum,
				"λ", lambda, "λ*q^W", lambda*math.Pow(q, sum)))
		}
	}
	return adjustment, uprob, unseenColors
}
