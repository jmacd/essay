module github.com/lightstep/sandbox/jmacd/essay

go 1.13

replace github.com/lightstep/sandbox/jmacd/multishape => ../multishape

replace github.com/lightstep/sandbox/jmacd/datashape => ../datashape

replace github.com/lightstep/sandbox/jmacd/gonum => ../gonum

replace github.com/jmacd/gospline => ../../../../jmacd/gospline

require (
	github.com/andybons/gogif v0.0.0-20140526152223-16d573594812
	github.com/jmacd/gospline v0.0.0-00010101000000-000000000000
	github.com/lightstep/sandbox/jmacd/gonum v0.0.0-00010101000000-000000000000
	github.com/lucasb-eyer/go-colorful v1.0.3
	golang.org/x/exp v0.0.0-20190731235908-ec7cb31e5a56
	gonum.org/v1/gonum v0.6.1
	gonum.org/v1/plot v0.0.0-20191107103940-ca91d9d40d0a
)
