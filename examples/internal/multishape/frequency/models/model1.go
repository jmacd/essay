package models

import (
	"fmt"

	ms "github.com/jmacd/essay/examples/internal/multishape"
	"github.com/jmacd/essay/examples/internal/multishape/continuous"
	"github.com/jmacd/essay/examples/internal/multishape/continuous/tdigest"
	"github.com/jmacd/essay/examples/internal/multishape/frequency"
	"github.com/jmacd/essay/examples/internal/multishape/sampler"
	"gonum.org/v1/gonum/mat"
)

// Model1 is transposed from Model0: rows are the variables and
// columns are the individual samples, but essentially the same logic.

type model1 struct {
	latencyVar ms.Variable
	colorVar   ms.Variable

	lquality tdigest.Quality
}

func NewModel1(latencyVar, colorVar ms.Variable, lquality tdigest.Quality) sampler.Model {
	return model1{latencyVar: latencyVar, colorVar: colorVar, lquality: lquality}
}

func (m1 model1) Update(inputSampled ms.Population, prior sampler.Scale) sampler.Scale {
	var sampled ms.Population
	if priorScale, ok := prior.(*frequency.MultiVarScale); ok {
		sampled = make(ms.Population, 0, len(inputSampled)+len(priorScale.Sampled))
		sampled = append(sampled, inputSampled...)
		sampled = append(sampled, priorScale.Sampled...)
	} else {
		sampled = inputSampled
	}

	const numVariables = 2.

	// Compute a latency digest
	latencies := sampled.Numbers(m1.latencyVar)

	// Count known colors
	colormap, _, _ := sampled.Categorize(m1.colorVar)
	numColors := len(colormap)

	latencyDigest := m1.lquality.New(
		latencies,
		sampled.Weights(),
		continuous.PositiveRange(),
		continuous.RangeSupported)

	numLatencies := latencyDigest.Size()

	// Treat columns as rows, points as columns.
	numRows := numLatencies + numColors
	numColumns := len(sampled)

	// Construct the indicator matrix X.
	X := mat.NewDense(numRows, numColumns, nil)
	W := 0.0 // Total mass

	// First construct an indicator matrix.
	for i, p := range sampled {
		latencyCol := latencyDigest.Lookup(p.Number(m1.latencyVar))

		if latencyCol < 0 || latencyCol >= numLatencies {
			panic(fmt.Sprintln("logic error", latencyCol, numLatencies))
		}

		colorCol := numLatencies + colormap[p.Category(m1.colorVar)]

		W += numVariables * p.Weight()
		X.Set(latencyCol, i, p.Weight())
		X.Set(colorCol, i, p.Weight())
	}

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
	// distance ratios, however this tells us something about the
	// data distribution and it should compute the same Chi^2
	// distance ratios.
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
		// fmt.Println("Normalized eigenvalue", i, "=", normalized.At(i, i))
	}

	// Chi-square for input rows:
	totalChiSq := 0.0
	chiSq := make([]float64, numRows)

	for row := range chiSq {
		frow := F.Slice(row, row+1, 0, numComponents)

		c2 := mat.NewDense(1, 1, nil)
		c2.Mul(frow, frow.T())

		chiSq[row] = c2.At(0, 0)
		totalChiSq += c2.At(0, 0)
	}

	for row := range chiSq {
		chiSq[row] /= totalChiSq
	}

	// sumLat, sumCol := 0.0, 0.0
	// for row := 0; row < numLatencies; row++ {
	// 	// fmt.Println("Latency inertia:", row, "=", chiSq[row])
	// 	sumLat += chiSq[row]
	// }
	// for row := numLatencies; row < numRows; row++ {
	// 	// fmt.Println("Color inertia:", row, "=", chiSq[row])
	// 	sumCol += chiSq[row]
	// }

	// fmt.Println("Sum of latency:", sumLat, "sum of color", sumCol)

	return frequency.NewMultiVarScale(inputSampled).Add(
		func(pt ms.Individual) float64 {
			lval := pt.Number(m1.latencyVar)
			latencyCol := latencyDigest.Lookup(lval)
			return chiSq[latencyCol]
		}).Add(
		func(pt ms.Individual) float64 {
			cval := pt.Category(m1.colorVar)
			colorNum, ok := colormap[cval]
			if !ok {
				// TODO This is bogus; see model2.
				return 0
			}
			colorCol := numLatencies + colorNum
			return chiSq[colorCol]
		})
}
