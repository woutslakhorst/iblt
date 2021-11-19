package bloom

type Adder interface {
	Add(data []byte) bool
	clone() Adder
}
