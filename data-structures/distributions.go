package datastructures

// Try to integrate the google-dp : https://github.com/google/differential-privacy/tree/main/go
import (
	// "crypto/rand"
	"math"
	"math/rand"
)

type GeomDistribution struct {
	epsilon     float64
	sensitivity float64
}

func NewGeomDistribution(epsilon, sensitivity float64) *GeomDistribution {
	return &GeomDistribution{
		epsilon:     epsilon,
		sensitivity: sensitivity,
	}
}

func (gd *GeomDistribution) DoubleGeomSample() int64 {
	p := 1.0 - math.Exp(-gd.epsilon/gd.sensitivity)
	geometric := rand.New(rand.NewSource(rand.Int63())).Intn(1000000) // Using a large constant for precision
	z1 := gd.sampleGeometric(geometric, p)
	z2 := gd.sampleGeometric(geometric, p)
	return int64(z1 - z2)
}

func (gd *GeomDistribution) sampleGeometric(n int, p float64) int {
	for i := 0; i < n; i++ {
		if rand.Float64() < p {
			return i
		}
	}
	return n
}

// func (gd *GeomDistribution) DoubleGeomSample() int {
// 	p := 1.0 - math.Exp(-gd.epsilon/gd.sensitivity)
// 	// geometric := gd.sampleGeometric(p)
// 	z1 := gd.sampleGeometric(p)
// 	z2 := gd.sampleGeometric(p)
// 	return int(z1 - z2)
// }

// func (gd *GeomDistribution) sampleGeometric(p float64) int {
// 	for i := 0; ; i++ {
// 		if randFloat64Secure() < p {
// 			return i
// 		}
// 	}
// }

// func randFloat64Secure() float64 {
// 	max := new(big.Int).SetUint64(1 << 63) // Explicitly set the type to int64
// 	randomBigInt, _ := rand.Int(rand.Reader, max)
// 	return float64(randomBigInt.Int64()) / float64(1<<63)
// }
