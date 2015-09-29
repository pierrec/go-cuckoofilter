package cuckoofilter

import (
	"math"

	"github.com/pierrec/xxHash/xxHash32"
	"github.com/pierrec/xxHash/xxHash64"
)

// bucketBitsNum is the number of bits present in a bucket.
const bucketBitsNumS = 8

// bucket can be changed to any Go uint type.
// bucketBitsNum must be changed accordingly.
type bucketS uint8

// fingerprint has fpBitsNum bits
// bucketBitsNum/fpBitsNum fingerprint entries per bucket
const (
	// number of bits for a fingerprint
	// 8 seems achieves a less than 1% failure rate
	// even under a high filter load.
	fpBitsNumS = 4
	fpMaskS    = 1<<fpBitsNumS - 1
	bucketNumS = bucketBitsNumS / fpBitsNumS
)

// fingerprint is stored in a bucket.
type fingerprintS uint8

// String prints out a bucket fingerprint items for debugging purposes.
// func (b bucketS) String() string {
// 	str := "{"
// 	for s := uint(0); s < bucketBitsNumS; s += fpBitsNumS {
// 		str += fmt.Sprintf(" %x", int(b>>s&fpMaskS))
// 	}
// 	return str + " }"
// }

// Filter represents a Cuckoo Filter with 1 byte per item and an error rate of 11.
type FilterS struct {
	f64 *Filter64S
}

// New creates a Filter containing up to n items with 1 byte per item and an error rate of 11.
func NewS(n uint) *FilterS {
	return &FilterS{
		f64: New64S(n),
	}
}

func fpValueS(b []byte) fingerprintS {
	h := uint64(xxHash32.Checksum(b, seed))
	return fpValue64S(h)
}

// fpValue returns a non-zero n bits fingerprint derived from the item hash.
func fpValue64S(h uint64) fingerprintS {
	for ; h&fpMaskS == 0; h >>= 4 {
		if h == 0 {
			return 1
		}
	}
	return fingerprintS(h & fpMaskS)
}

// func (cf *FilterS) String() string {
// 	return fmt.Sprint(cf.f64.bucketsS)
// }

// Insert add an item to the Filter and returns whether it was inserted or not.
//
// NB. The same item cannot be inserted more than 2 times.
func (cf *FilterS) Insert(b []byte) bool {
	return cf.f64.insert(xxHash64.Checksum(b, seed), fpValueS(b))
}

// Has checks if the item is in the Filter.
func (cf *FilterS) Has(b []byte) bool {
	return cf.f64.has(xxHash64.Checksum(b, seed), fpValueS(b))
}

// Delete removes an item from the Filter and returns whether or not it was present.
// To delete an item safely it must have been previously inserted.
func (cf *FilterS) Delete(b []byte) bool {
	return cf.f64.delete(xxHash64.Checksum(b, seed), fpValueS(b))
}

// Len returns the number of items in the filter.
func (cf *FilterS) Len() int { return cf.f64.Len() }

// Cap returns the filter capacity.
func (cf *FilterS) Cap() int { return cf.f64.Cap() }

// Filter64 represents a Cuckoo Filter that only stores uint64 items
// and as such is much faster than FilterS.
type Filter64S struct {
	buckets []bucketS
	mask    uint64 // mask to be used for buckets indexing
	max     int    // maximum number of relocations
}

// New64 creates a Filter64S containing up to n uint64 items.
// A Filter64S contains a minimum of 2 items.
func New64S(n uint) *Filter64S {
	// each bucket holds bucketBitsNum/fpBitsNum fingerprint entries
	// of fpBitsNum bits each
	n2 := power2(n)
	if n2 < 2 {
		n2 = 2
	}

	// base the maximum numver of relocations on the number of buckets
	max := int(20 * math.Log(float64(n2)))

	return &Filter64S{
		buckets: make([]bucketS, n2, n2),
		mask:    uint64(n2 - 1),
		max:     max,
	}
}

// index1 computes the first index from the hash and the fingerprint.
func (cf *Filter64S) index1(h uint64, fp fingerprintS) uint64 {
	return h & cf.mask
}

// index2 computes the second index from the previous index and the fingerprint.
// Note that index2(index2(i, fp)) == i, which is leveraged when reverting relocation.
func (cf *Filter64S) index2(i uint64, fp fingerprintS) uint64 {
	return (i ^ fpHash(uint64(fp))) & cf.mask
}

// insertAt adds fingerprint fp to the filter at index i and returns
// if the insert succeeded.
func (cf *Filter64S) insertAt(idx uint64, fp fingerprintS) bool {
	u := cf.buckets[idx]
	for i := 0; i < bucketNumS; i++ {
		s := uint(i * fpBitsNumS)
		if u>>s&fpMaskS == 0 {
			cf.buckets[idx] |= bucketS(fp) << s
			return true
		}
	}
	return false
}

// Insert add an item to the Filter and returns whether it was inserted or not.
//
// NB. The same item cannot be inserted more than 2 times.
func (cf *Filter64S) Insert(x uint64) bool {
	return cf.insert(x, fpValue64S(x))
}

// insert adds an item with hash h and fingerprint fp.
func (cf *Filter64S) insert(h uint64, fp fingerprintS) bool {
	// find an empty entry slot at the first index
	i := cf.index1(h, fp)
	if cf.insertAt(i, fp) {
		return true
	}

	// find an empty entry slot at the second index
	j := cf.index2(i, fp)
	if cf.insertAt(j, fp) {
		return true
	}

	// no empty slot, kick one entry and relocate it at its next index
	// only do it so many times to avoid infinite loops
	for r := 0; r < cf.max; r++ {
		// kick the first entry located in the current bucket
		pfp := fingerprintS(cf.buckets[i] & fpMaskS)
		cf.buckets[i] &= ^bucketS(fpMaskS)
		cf.buckets[i] |= bucketS(fp)
		// find a new location for the previous fingerprint
		fp = pfp
		i = cf.index2(i, fp)
		if cf.insertAt(i, fp) {
			return true
		}
	}

	// relocation impossible: restore the relocated items
	// since the previous index can be computed via the
	// current fingerprint and the current index
	for r := 0; r < cf.max; r++ {
		i = cf.index2(i, fp)
		pfp := fingerprintS(cf.buckets[i] & fpMaskS)
		cf.buckets[i] &= ^bucketS(fpMaskS)
		cf.buckets[i] |= bucketS(fp)
		fp = pfp
	}

	return false
}

// hasIn checks bucket b for fingerprint fp and returns whether it was found.
func (cf *Filter64S) hasIn(b bucketS, fp fingerprintS) bool {
	for i := 0; i < bucketNumS; i++ {
		if b>>uint(i*fpBitsNumS)&fpMaskS == bucketS(fp) {
			return true
		}
	}
	return false
}

// Has checks if the item is in the Filter.
// If it returns false, then the item is definitely not in the filter.
// If it returns true, then the item *may* be in the filter, although
// with a low probability of not being in it.
func (cf *Filter64S) Has(x uint64) bool {
	return cf.has(x, fpValue64S(x))
}

// has checks if the item hash and fingerprint are found.
func (cf *Filter64S) has(h uint64, fp fingerprintS) bool {
	// check fingerprint at first index
	i := cf.index1(h, fp)
	if cf.hasIn(cf.buckets[i], fp) {
		return true
	}

	// check fingerprint at second index
	i = cf.index2(i, fp)
	return cf.hasIn(cf.buckets[i], fp)
}

// deleteAt removes fingerprint fp from the filter at index i and
// returns whether the fingerprint was found or not.
func (cf *Filter64S) deleteAt(idx uint64, fp fingerprintS) bool {
	u := cf.buckets[idx]
	for i := 0; i < bucketNumS; i++ {
		s := uint(i * fpBitsNumS)
		if u>>s&fpMaskS == bucketS(fp) {
			cf.buckets[idx] &^= bucketS(fpMaskS) << s
			return true
		}
	}
	return false
}

// Delete removes an item from the Filter and returns whether or not it was present.
// To delete an item safely it must have been previously inserted.
func (cf *Filter64S) Delete(x uint64) bool {
	return cf.delete(x, fpValue64S(x))
}

// delete removes the item with the corresponding hash and fingerprint.
func (cf *Filter64S) delete(h uint64, fp fingerprintS) bool {
	// delete fingerprint at first index
	i := cf.index1(h, fp)
	if cf.deleteAt(i, fp) {
		return true
	}

	// delete fingerprint at second index
	i = cf.index2(i, fp)
	return cf.deleteAt(i, fp)
}

// Len returns the number of items in the filter.
func (cf *Filter64S) Len() (n int) {
	for _, b := range cf.buckets {
		for ; b > 0; b >>= fpBitsNumS {
			if b&fpMaskS > 0 {
				n++
			}
		}
	}
	return
}

// Cap returns the filter capacity (maximum number of items it may contain).
func (cf *Filter64S) Cap() int {
	return len(cf.buckets)
}
