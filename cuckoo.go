//go:generate go run generate/main.go

/*
Package cuckoofilter implements the Cuckoo Filter algorithm for approximated
set-membership queries. This means that one can check for data being in a set
and either get a definitive answer if it is not in the set, or a "maybe" answer
but with a low probability of being wrong. Cuckoo filters are similar in
functionality to Bloom filters, but they support removing items from the filter
without altering future results.

Six Cuckoo filter types are defined to be used in different cases:
	Filter64S: for uint64 items, with low memory footprint (1 byte per item) and with an error rate of ~11%
	Filter64M: for uint64 items, with medium memory footprint (2 bytes per item) and with an error rate of ~1%
	Filter64L: for uint64 items, with high memory footprint (8 bytes per item) and with an error rate of ~0.005%
	FilterS: low memory footprint (1 byte per item) and with an error rate of ~11%
	FilterM: medium memory footprint (2 bytes per item) and with an error rate of ~1%
	FilterL: high memory footprint (8 bytes per item) and with an error rate of ~0.005%

Filter64* types are much faster then their siblings.

See https://www.cs.cmu.edu/~dga/papers/cuckoo-conext2014.pdf for the theory behind Cuckoo filters.
*/
package cuckoofilter

// power2 returns the next power of two for n.
// http://graphics.stanford.edu/~seander/bithacks.html#RoundUpPowerOf2
func power2(n uint) uint {
	n--
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	n |= n >> 32
	return n + 1
}

const (
	prime64_1 = uint64(11400714785074694791)
	prime64_2 = uint64(14029467366897019727)
	prime64_3 = uint64(1609587929392839161)
	prime64_4 = uint64(9650029242287828579)
	prime64_5 = uint64(2870177450012600261)
	seed      = 0xC0CC00 // "coccoo"
)

// fast xxHash64 for 8 bytes
func fpHash(x uint64) uint64 {
	h64 := prime64_5 + 8
	p64 := x * prime64_2
	h64 ^= ((p64 << 31) | (p64 >> 33)) * prime64_1
	h64 = ((h64<<27)|(h64>>37))*prime64_1 + prime64_4

	h64 ^= h64 >> 33
	h64 *= prime64_2
	h64 ^= h64 >> 29
	h64 *= prime64_3
	h64 ^= h64 >> 32

	return h64
}
