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

func BenchmarkIbf_a(b *testing.B) {
	var nDecodes = make([]int, b.N)

	for n := 0; n < b.N; n++ {
		nDecodes[n] = runTest()
	}
	statistics(nDecodes, b)
}

// doubling the amount of buckets, more than doubles the set difference that can be solved
func runTest() int {
	numBuckets := 1024
	ibfA := NewIbf(numBuckets)
	ibfB := NewIbf(numBuckets)

	N := 600 // common size and set difference, each set has N/2 keys the other doesn't have
	for i := 0; i < N; i++ {
		a := generateData()
		ibfA.Add(a)
		if i%2 == 0 {
			ibfB.Add(generateData())
		} else {
			ibfB.Add(a)
		}
	}
	ibfA.Subtract(ibfB)
	if _, _, err := ibfA.Decode(); err != nil {
		return 0
	}

	return 1
}

func TestIbf_Decode(t *testing.T) {
	numBuckets := 256
	ibfA := NewIbf(numBuckets)
	ibfB := NewIbf(numBuckets)

	N := 200 // common size and set difference, each set has N/2 keys the other doesn't have
	for i := 0; i < N; i++ {
		a := generateData()
		ibfA.Add(a)
		if i%2 == 0 {
			ibfB.Add(generateData())
		} else {
			ibfB.Add(a)
		}
	}
	fmt.Println("initialized:")
	fmt.Printf("bucketA[0]: %v\n", ibfA.buckets[0])
	fmt.Printf("bucketB[0]: %v\n", ibfB.buckets[0])

	err := ibfA.Subtract(ibfB)
	//fmt.Println("subtracted:")
	//if err != nil {
	//	fmt.Printf("subtract error: %s\n", err.Error())
	//}
	//for i := range ibfA.buckets {
	//	fmt.Printf("bucket[%03d]: %v\n", i, ibfA.buckets[i])
	//}

	_, _, err = ibfA.Decode()
	fmt.Println("decoded:")
	if err != nil {
		fmt.Printf("decode error: %s\n", err.Error())
	}
	//for i := range ibfA.buckets {
	//	fmt.Printf("bucket[%03d]: %v\n", i, ibfA.buckets[i])
	//}
	//
	//fmt.Printf("remaining: (%d)\n", len(remaining))
	//for _, x := range remaining {
	//	fmt.Printf("\t%x\n", x)
	//}
	//fmt.Printf("missing: %d\n", len(missing))
	//for _, x := range missing {
	//	fmt.Printf("\t%x\n", x)
	//}
}
