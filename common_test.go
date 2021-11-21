package bloom

import (
	"math"
	"math/rand"
	"testing"
)

func benchmarkAdd(f Bloom, nEval, maxEntries int) (nCollisions []int, firstCollision []int) {
	nCollisions = make([]int, nEval)
	firstCollision = make([]int, nEval)
	for i := 0; i < nEval; i++ {
		nCollisions[i] = countCollisions(f.clone(), maxEntries)
		firstCollision[i] = addUntilCollision(f.clone())
	}
	return
}

func statistics(values []int, b *testing.B) (mean float64, std float64, min int, max int) {
	sum := 0
	min = math.MaxInt
	max = math.MinInt
	for _, v := range values {
		sum += v
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	mean = float64(sum) / float64(len(values))

	ssd := 0.
	for _, v := range values {
		ssd += math.Pow(float64(v)-mean, 2)
	}
	std = math.Sqrt(ssd / float64(len(values)-1))
	b.Logf("mean: %.2f, std: %.2f, min: %d, max: %d\n", mean, std, min, max)

	return
}

func countCollisions(filter Bloom, numEntries int) int {
	nCollisions := 0
	for n := 0; n < numEntries; n++ {
		if !filter.Add(generateData()) {
			nCollisions++
		}
	}
	return nCollisions
}

func addUntilCollision(filter Bloom) int {
	for totalElem := 0; ; totalElem++ {
		if !filter.Add(generateData()) {
			return totalElem
		}
	}
}

func generateData() []byte {
	bytes := make([]byte, KeyLength) // Tx ids use 256-bit hashes
	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}
	return bytes
}
