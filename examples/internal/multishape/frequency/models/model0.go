package models

import (
	"fmt"

	ms "github.com/jmacd/essay/examples/internal/multishape"
	"github.com/jmacd/essay/examples/internal/multishape/continuous"
	"github.com/jmacd/essay/examples/internal/multishape/continuous/tdigest"
	"github.com/jmacd/essay/examples/internal/multishape/frequency"
	"github.com/jmacd/essay/examples/internal/multishape/sampler"
	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/gonum/stat/distuv"
)

// Model0 computes the SVD computes Chi^2 in several ways, mainly to
// test the theory and practice. This is copied from the summary in
// Multiple Correspondence Analysis Abdi & Valentin (chapter
// Encyclopedia of Measurement and Statistics).
type model0 struct {
	latencyVar ms.Variable
	colorVar   ms.Variable

	lquality tdigest.Quality
}

func NewModel0(latencyVar, colorVar ms.Variable, lquality tdigest.Quality) sampler.Model {
	return model0{latencyVar: latencyVar, colorVar: colorVar, lquality: lquality}
}

func (m0 model0) Update(sampled ms.Population, scale sampler.Scale) sampler.Scale {
	const numVariables = 2.

	// Compute a latency digest
	latencies := sampled.Numbers(m0.latencyVar)
	lweights := sampled.Weights()

	// Count known colors
	colormap := sampled.Categorize(m0.colorVar).Indexed
	numColors := len(colormap)

	latencyDigest := m0.lquality.New(
		latencies,
		lweights,
		continuous.PositiveRange(),
		continuous.RangeSupported)

	numLatencies := latencyDigest.Size()

	// Number of rows and columns
	numPoints := len(sampled)
	numColumns := numLatencies + numColors

	// Construct the indicator matrix X.
	X := mat.NewDense(numPoints, numColumns, nil)
	W := 0.0 // Total row mass
	// First construct an indicator matrix.
	for i, p := range sampled {
		latencyCol := latencyDigest.Lookup(p.Number(m0.latencyVar))

		if latencyCol < 0 || latencyCol >= numLatencies {
			panic(fmt.Sprintln("logic error", latencyCol, numLatencies))
		}

		colorCol := numLatencies + colormap[p.Category(m0.colorVar)]

		W += numVariables * p.Weight()
		X.Set(i, latencyCol, p.Weight())
		X.Set(i, colorCol, p.Weight())
	}
	// Compute the probability matrix.
	Z := mat.NewDense(numPoints, numColumns, nil)
	Z.Scale(1/W, X)

	// Compute row profile (numPoints rows, 1 column).
	r := mat.NewDense(numPoints, 1, nil)
	r.Mul(Z, constMat(numColumns, 1, 1))

	// Compute column profile (1 row, numColumns columns).
	c := mat.NewDense(1, numColumns, nil)
	c.Mul(constMat(1, numPoints, 1), Z)

	// Diagonalize r and c.
	Dr := diagonalize(r.ColView(0))
	Dc := diagonalize(c.RowView(0))

	invSqrtDr := mat.NewDense(numPoints, numPoints, nil)
	invSqrtDr.Apply(invSqrt, Dr)

	invSqrtDc := mat.NewDense(numColumns, numColumns, nil)
	invSqrtDc.Apply(invSqrt, Dc)

	// Compute r * cT (N.B. 'c' is already transposed)
	rc := mat.NewDense(numPoints, numColumns, nil)
	rc.Mul(r, c)

	// Compute (Z - rcT)
	centered := mat.NewDense(numPoints, numColumns, nil)
	centered.Sub(Z, rc)

	// Compute the MCA product matrix
	product := mat.NewDense(numPoints, numColumns, nil)
	product.Mul(invSqrtDr, centered)
	product.Mul(product, invSqrtDc)

	// Note: we do not need to compute the SVD to calculate Chi^2
	// distance, however this tells us something about the data
	// distribution and it should compute the same Chi^2 distances.
	var svd mat.SVD
	if !svd.Factorize(product, mat.SVDThin) {
		panic("Could not factorize product matrix")
	}

	// See how many values are significant, i.e., those > 1/numColumns
	// N.B. Assumes Rows > Cols.
	K := float64(numColumns) - 1
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
	var V mat.Dense

	svd.UTo(&U)
	svd.VTo(&V)

	invValues := mat.NewDense(numComponents, numComponents, nil)
	invValues.Apply(inverse, Values)

	// Factor scores for the columns (N.B. assuming #cols < #rows)
	VSlice := V.Slice(0, numColumns, 0, numComponents)

	G := mat.NewDense(numColumns, numComponents, nil)
	G.Mul(invSqrtDc, VSlice)
	G.Mul(G, Values)

	// Factor scores for the rows.
	USlice := U.Slice(0, numPoints, 0, numComponents)
	F := mat.NewDense(numPoints, numComponents, nil)
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

	// First formula
	FFT := mat.NewDense(numPoints, numPoints, nil)
	FFT.Mul(F, F.T())

	// Second formula
	GVInv := mat.NewDense(numColumns, numComponents, nil)
	GVInv.Mul(G, invValues)

	chiSqTotal1 := 0.0
	chiSqTotal2 := 0.0

	for row := 0; row < numPoints; row++ {
		// First formula
		chiSqTotal1 += FFT.At(row, row)

		// Second formula
		frow := mat.NewDense(1, numComponents, nil)

		frow.Mul(X.Slice(row, row+1, 0, numColumns), GVInv)
		frow.Scale(1/(numVariables*sampled[row].Weight()), frow)

		c2 := mat.NewDense(1, 1, nil)
		c2.Mul(frow, frow.T())

		chiSqTotal2 += c2.At(0, 0)
	}

	chiSqTotal3 := 0.0

	// Third formula
	for i := 0; i < numPoints; i++ {
		for j := 0; j < numColumns; j++ {
			d := r.At(i, 0) * c.At(0, j)
			x := Z.At(i, j) - d

			// N.B. With a weighted indicator matrix, this
			// number is different than the Chi^2 computed
			// from the factorization.  I.e., this should
			// match formulas 1 and 2 when the weights are
			// identical, i.e., in the 0th window.
			chiSqTotal3 += float64(numPoints) * x * x / d
		}
	}

	fmt.Println("First  Chi^2", chiSqTotal1)
	fmt.Println("Second Chi^2", chiSqTotal2)
	fmt.Println("Third  Chi^2", chiSqTotal3)

	// Test for significance.
	significance := 0.95
	degFreedom := (numColumns - 1) * (numPoints - 1)

	testStat := distuv.ChiSquared{K: float64(degFreedom)}.Quantile(significance)

	fmt.Println("Compare Chi^2", chiSqTotal1, "vs test statistic", testStat)

	// Compute Column Chi^2 values, probabilities.
	GGT := mat.NewDense(numColumns, numColumns, nil)
	GGT.Mul(G, G.T())

	colChiSum := 0.0

	for col := 0; col < numColumns; col++ {
		colChiSum += GGT.At(col, col)
	}
	scaleWeights := make([]float64, numColumns)
	for col := 0; col < numColumns; col++ {
		scaleWeights[col] = GGT.At(col, col) / colChiSum
		fmt.Print("Col ", col, " inertia ", 100.*scaleWeights[col], "%\n")
	}

	sumLat, sumCol := 0.0, 0.0
	for col := 0; col < numLatencies; col++ {
		sumLat += scaleWeights[col]
	}
	for col := numLatencies; col < numColumns; col++ {
		sumCol += scaleWeights[col]
	}

	fmt.Println("Sum of latency probs:", sumLat, "sum of color probs", sumCol)

	// Compute the new from the column Chi^2 probabilities
	scale = frequency.NewMultiVarScale(sampled).Add(
		func(pt ms.Individual) float64 {
			lval := pt.Number(m0.latencyVar)
			latencyCol := latencyDigest.Lookup(lval)
			return scaleWeights[latencyCol]
		}).Add(
		func(pt ms.Individual) float64 {
			cval := pt.Category(m0.colorVar)
			colorNum := colormap[cval]
			colorCol := numLatencies + colorNum
			return scaleWeights[colorCol]
		})

	return scale
}
