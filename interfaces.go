package bloom

const KeyLength = 32

type Bloom interface {
	Add(data []byte) bool
	clone() Bloom
}
