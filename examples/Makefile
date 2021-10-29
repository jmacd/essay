all: perceptual_color_spaces/index.html univariate_distributions/index.html

perceptual_color_spaces/index.html: color/color.go
	go run $<

univariate_distributions/index.html: histo/histo.go
	go run $<
