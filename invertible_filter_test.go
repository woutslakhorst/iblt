package bloom

import (
	"fmt"
	"testing"
)

func BenchmarkIbf_Add(b *testing.B) {
	buckets := 256
	filter := NewIbf(buckets)
	numEntries := 1000
	var nCollisions, firstCollision []int

	// benchmark loop
	for n := 0; n < b.N; n++ {
		nCollisions, firstCollision = benchmarkAdd(filter, n, numEntries)
	}

	b.Logf("Filter size: %d buckets, hashes: %d\n", len(filter.buckets), len(filter.seeds))
	b.Logf("First collisions - ")
	statistics(firstCollision, b)
	b.Logf("Average collisions for %d entries - ", numEntries)
	statistics(nCollisions, b)
}

func TestIbf_Decode(t *testing.T) {
	numBuckets := 32
	ibfA := NewIbf(numBuckets)
	ibfB := NewIbf(numBuckets)

	N := 22000
	for i := 0; i < N; i++ {
		a := generateData()
		ibfA.Add(a)
		if i%2000 == 0 {
			ibfB.Add(generateData())
		} else {
			ibfB.Add(a)
		}
	}
	fmt.Printf("bucketA[0]: %v\n", ibfA.buckets[0])
	fmt.Printf("bucketB[0]: %v\n", ibfB.buckets[0])

	err := ibfA.Subtract(ibfB)
	if err != nil {
		fmt.Printf("subtract error: %s\n", err.Error())
	}
	for i := range ibfA.buckets {
		fmt.Printf("bucket[%03d]: %v\n", i, ibfA.buckets[i])
	}

	remaining, missing, err := ibfA.Decode()
	if err != nil {
		fmt.Printf("decode error: %s\n", err.Error())
	}

	fmt.Printf("remaining: %d\n", len(remaining))
	fmt.Printf("missing: %d\n", len(missing))
	for i := range ibfA.buckets {
		fmt.Printf("bucket[%03d]: %v\n", i, ibfA.buckets[i])
	}
}
