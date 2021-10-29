package num

type (
	ABuilder struct {
		logscale bool
		min, max *float64
		factor   float64
		nominal  []string
		hide     bool
	}
)

func Axis() ABuilder {
	return ABuilder{}
}

func (a ABuilder) LogScale(x bool) ABuilder {
	a.logscale = x
	if x {
		a.factor = defaultLogBaseFactor
	} else {
		a.factor = 0
	}
	return a
}

func (a ABuilder) LogScaleFactor(f float64) ABuilder {
	a.logscale = true
	a.factor = f
	return a
}

func (a ABuilder) Min(value float64) ABuilder {
	a.min = &value
	return a
}

func (a ABuilder) Max(value float64) ABuilder {
	a.max = &value
	return a
}

func (a ABuilder) Range(min, max float64) ABuilder {
	a.min = &min
	a.max = &max
	return a
}

func (a ABuilder) Nominal(names []string) ABuilder {
	a.nominal = names
	return a
}

func (a ABuilder) Hide() ABuilder {
	a.hide = true
	return a
}
