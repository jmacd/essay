package c

// #include "k.h"
import "C"

func K(n int, d float64) float64 {
	return float64(C.K(C.int(n), C.double(d)))
}
