package bloom

import (
	"errors"
	"fmt"
	"github.com/spaolacci/murmur3"
	"sort"
)

//type IBF interface {
//	Add(keySum []byte) bool
//	Bloom
//
//	// Delete remove keySum from IBF. Does not verify if the IBF contains the keySum and can results in negative bucket counts.
//	Delete(keySum []byte)
//
//	// Subtract returns some other IBF subtracted from this IBF. Returns an error if the number of buckets or the hashSum seeds differ.
//	Subtract(other IBF)
//
//	// Decode peels of 'pure' entries into remaining (in this ibf) or missing (in subtracted ibf). This returns an error if
//	Decode(remaining, missing [][]byte) error
//}

type ibf struct {
	buckets    []*bucket `json:"buckets"`
	numBuckets int       `json:"num_buckets"`
	seeds      []uint32  `json:"seeds"`
	keySeed    uint32    `json:"key_seed"`
	keyLength  int       `json:"key_length"`
}

func NewIbf(numBuckets int) *ibf {
	buckets := make([]*bucket, numBuckets)
	for i := 0; i < numBuckets; i++ {
		buckets[i] = newBucket(KeyLength)
	}
	return &ibf{
		buckets:    buckets,
		seeds:      []uint32{0, 1, 2, 3},
		keySeed:    uint32(33),
		keyLength:  KeyLength,
		numBuckets: numBuckets,
	}
}

func (i *ibf) clone() Bloom {
	newIbf := NewIbf(i.numBuckets)
	seeds := make([]uint32, len(i.seeds))
	copy(seeds, i.seeds)
	newIbf.seeds = seeds
	return newIbf
}

func (i *ibf) Add(key []byte) bool {
	hash := i.hashKey(key)
	idxs := i.hashIndices(key)
	for _, h := range idxs {
		i.buckets[h].add(key, hash)
	}

	// validity can only be guaranteed if nothing has been subtracted or deleted from the ibf
	for _, h := range idxs {
		if i.buckets[h].count < 2 {
			return true
		}
	}
	return false
}

func (i *ibf) Delete(key []byte) {
	hash := i.hashKey(key)
	for _, h := range i.hashIndices(key) {
		i.buckets[h].delete(key, hash)
	}
}

func (i *ibf) Subtract(other *ibf) error {
	if err := i.validateSubtrahend(other); err != nil {
		return fmt.Errorf("subtraction failed: %w", err)
	}
	for idx, b := range i.buckets {
		b.subtract(other.buckets[idx])
	}
	return nil
}

func (i *ibf) validateSubtrahend(o *ibf) error {
	if i.numBuckets != o.numBuckets {
		return fmt.Errorf("unequal number of buckets, expected (%d) got (%d)", i.numBuckets, o.numBuckets)
	}
	if i.keySeed != o.keySeed {
		return fmt.Errorf("keySeeds do not match, expected (%d) got (%d)", i.keySeed, o.keySeed)
	}
	if i.keyLength != o.keyLength {
		return fmt.Errorf("keyLengths do not match, expected (%d) got (%d)", i.keySeed, o.keySeed)
	}
	if len(i.seeds) != len(o.seeds) {
		return fmt.Errorf("kunequal number of seeds, expected (%d) got (%d)", i.seeds, o.seeds)
	}
	sort.Slice(i.seeds, func(x, y int) bool { return i.seeds[x] < i.seeds[y] })
	sort.Slice(o.seeds, func(x, y int) bool { return o.seeds[x] < o.seeds[y] })
	for idx := range i.seeds {
		if i.seeds[idx] != o.seeds[idx] {
			return fmt.Errorf("seeds do not match, expected %v got %v", i.seeds, o.seeds)
		}
	}
	return nil
}

func (i *ibf) Decode() (remaining [][]byte, missing [][]byte, err error) {
	for {
		updated := false

		// for each pure (count == +1 or -1), if hashSum = h(key) -> Add(count == -1)/Delete(count == 1) key
		for _, b := range i.buckets {
			if (b.count == 1 || b.count == -1) && i.hashKey(b.keySum) == b.hashSum {
				if b.count == 1 {
					remaining = append(remaining, b.keySum)
					i.Delete(b.keySum)
				} else { // b.count == -1
					missing = append(missing, b.keySum)
					i.Add(b.keySum)
				}
				updated = true
			}
		}

		// if no pures exist, the ibf is empty or cannot be decoded
		if !updated {
			for _, b := range i.buckets {
				if !b.isEmpty() {
					return remaining, missing, errors.New("decode failed")
				}
			}
			return remaining, missing, nil
		}
	}
}

func (i *ibf) hashIndices(key []byte) []uint32 {
	hashes := make([]uint32, len(i.seeds))
	for idx, seed := range i.seeds {
		hashes[idx] = murmur3.Sum32WithSeed(key, seed) % uint32(i.numBuckets)
	}
	return unique(hashes)
}

func unique(uintSlice []uint32) []uint32 {
	keys := make(map[uint32]struct{})
	var uniques []uint32
	for _, v := range uintSlice {
		if _, exists := keys[v]; !exists {
			keys[v] = struct{}{}
			uniques = append(uniques, v)
		}
	}
	return uniques
}

func (i *ibf) hashKey(key []byte) uint32 {
	return murmur3.Sum32WithSeed(key, i.keySeed)
}

// bucket
type bucket struct {
	count   int // signed to allow for negative counts after subtraction
	keySum  []byte
	hashSum uint32
}

func newBucket(keyLength int) *bucket {
	return &bucket{
		count:   0,
		keySum:  make([]byte, keyLength),
		hashSum: 0,
	}
}

func (b *bucket) add(key []byte, hash uint32) {
	b.count++
	b.update(key, hash)
}

func (b *bucket) delete(key []byte, hash uint32) {
	b.count--
	b.update(key, hash)
}

func (b *bucket) subtract(o *bucket) {
	b.count -= o.count
	b.update(o.keySum, o.hashSum)
}

func (b *bucket) update(key []byte, hash uint32) {
	b.keySum = xor(b.keySum, key)
	b.hashSum ^= hash
}

func (b *bucket) isEmpty() bool {
	return b.count == 0 && b.hashSum == 0 && eq(b.keySum, make([]byte, KeyLength))
}

func (b *bucket) String() string {
	return fmt.Sprintf("[count: %3d, keySum: %x, hashSum: %10d]", b.count, b.keySum, b.hashSum)
}
