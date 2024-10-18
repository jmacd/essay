all: perceptual_color_spaces/index.html univariate_distributions/index.html

perceptual_color_spaces/index.html: color/color.go
	go run $<

univariate_distributions/index.html: histo/histo.go
	go run $<

kolmogorov_distribution/index.html: kolmogorov/kolmo.go
	go run $<

sampling_intro/index.html: samp
