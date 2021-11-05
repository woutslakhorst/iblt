package bloom

import (
	"crypto/rand"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func BenchmarkFilter_Add(b *testing.B) {
	filter := NewFilter1024()

	nEval := 1000
	collision := 0
	for i := 0; i < nEval; i++ {
		collision += addUntilCollision(filter)
	}
	fmt.Printf("Collision occurs on average after adding %.2f datapoints\n", float32(collision) / float32(nEval))
	b.Fail()

	//collision = 0
	//collision += addDatasetUntilCollision(filter, generateDataset(1000))
	//fmt.Printf("Collision occurs on average after adding %.2f datapoints\n", float32(collision) / float32(nEval))
}

func addDatasetUntilCollision(filter Filter, dataset [][]byte) int {
	for i, v := range dataset {
		if !filter.Add(v) {
			return i
		}
	}
	fmt.Printf("complete dataset added\n")
	return -1
}

func addUntilCollision(filter Filter) int {
	for totalElem := 0;;totalElem++ {
		if !filter.Add(generateData()) {
			return totalElem
		}
	}
}

func generateData() []byte {
	bytes := make([]byte, 256)  // Tx ids use 256-byte hashes
	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}
	return bytes
}

func generateDataset(n int) [][]byte {
	dataset := make([][]byte, n)
	for i := 0; i < n; i++ {
		dataset[i] = generateData()
	}
	return dataset
}

func TestFilter_Add(t *testing.T) {
	filter := Filter{
		fingerprint: make([]byte, 32),
		seeds: []uint32{0,1,2},
	}
	d1 := generateData()
	d2 := generateData()
	assert.True(t, filter.Add(d1), "failed to add first data point")
	assert.True(t, filter.Add(d2), "failed to add second data point")
	assert.False(t, filter.Add(d2), "adding data point for the second time should fail")
}

func TestFilter_Check(t *testing.T) {
	filter := Filter{
		fingerprint: make([]byte, 32),
		seeds: []uint32{0,1,2},
	}
	d1 := generateData()
	d2 := generateData()

	if !filter.Add(d1) {
		t.Errorf("Failed to add data to an empty filter")
	}

	assert.True(t, filter.Check(d1), "failed to identify only datapoint in filter")
	assert.False(t, filter.Check(d2), "identified datapoint not in filter") // fails
}

func TestFilter_hashValues(t *testing.T) {
	// TODO
}

func TestNewFilter512(t *testing.T) {
	flt := NewFilter512()
	assert.Equal(t, 512, len(flt.fingerprint), "incorrect fingerprint length")
	assert.Equal(t, 5, len(flt.seeds), "incorrect number of hash seeds")
	//assert.Equal(t, 0, flt.checksum, "initial checksum not zero")
}

