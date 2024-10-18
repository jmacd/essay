package datashape

type (
	Species interface{}

	Frequency   int64
	Frequencies map[Species]Frequency
	Weights     map[Species]float64

	FreqFreqs map[Frequency]Frequency
)

func (fs Frequencies) ToFF() FreqFreqs {
	ff := FreqFreqs{}
	for _, f := range fs {
		ff[f]++
	}
	return ff
}
