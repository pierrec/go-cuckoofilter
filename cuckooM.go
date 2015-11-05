package cuckoofilter

import (
	"math"

	"github.com/pierrec/xxHash/xxHash32"
	"github.com/pierrec/xxHash/xxHash64"
)

// bucketBitsNum is the number of bits present in a bucket.
const bucketBitsNumM = 16

// bucket can be changed to any Go uint type.
// bucketBitsNum must be changed accordingly.
type bucketM uint16

// fingerprint has fpBitsNum bits
// bucketBitsNum/fpBitsNum fingerprint entries per bucket
const (
	// number of bits for a fingerprint
	// 8 seems achieves a less than 1% failure rate
	// even under a high filter load.
	fpBitsNumM = 8
	fpMaskM    = 1<<fpBitsNumM - 1
	bucketNumM = bucketBitsNumM / fpBitsNumM
)

// fingerprint is stored in a bucket.
type fingerprintM uint8

// String prints out a bucket fingerprint items for debugging purposes.
// func (b bucketM) String() string {
// 	str := "{"
// 	for s := uint(0); s < bucketBitsNumM; s += fpBitsNumM {
// 		str += fmt.Sprintf(" %x", int(b>>s&fpMaskM))
// 	}
// 	return str + " }"
// }

// FilterM represents a Cuckoo Filter with 2 bytes per item and an error rate of 1.
type FilterM struct {
	f64 *Filter64M
}

// NewM creates a Filter containing up to n items with 2 bytes per item and an error rate of 1.
func NewM(n uint) *FilterM {
	return &FilterM{
		f64: New64M(n),
	}
}

func fpValueM(b []byte) fingerprintM {
	h := uint64(xxHash32.Checksum(b, seed))
	return fpValue64M(h)
}

// fpValue returns a non-zero n bits fingerprint derived from the item hash.
func fpValue64M(h uint64) fingerprintM {
	for ; h&fpMaskM == 0; h >>= 4 {
		if h == 0 {
			return 1
		}
	}
	return fingerprintM(h & fpMaskM)
}

// func (cf *FilterM) String() string {
// 	return fmt.Sprint(cf.f64.bucketsM)
// }

// Insert add an item to the Filter and returns whether it was inserted or not.
//
// NB. The same item cannot be inserted more than 2 times.
func (cf *FilterM) Insert(b []byte) bool {
	return cf.f64.insert(xxHash64.Checksum(b, seed), fpValueM(b))
}

// Has checks if the item is in the Filter.
func (cf *FilterM) Has(b []byte) bool {
	return cf.f64.has(xxHash64.Checksum(b, seed), fpValueM(b))
}

// Delete removes an item from the Filter and returns whether or not it was present.
// To delete an item safely it must have been previously inserted.
func (cf *FilterM) Delete(b []byte) bool {
	return cf.f64.delete(xxHash64.Checksum(b, seed), fpValueM(b))
}

// Len returns the number of items in the filter.
func (cf *FilterM) Len() int { return cf.f64.Len() }

// Cap returns the filter capacity.
func (cf *FilterM) Cap() int { return cf.f64.Cap() }

// Filter64M represents a Cuckoo Filter that only stores uint64 items
// and as such is much faster than FilterM.
type Filter64M struct {
	buckets []bucketM
	mask    uint64 // mask to be used for buckets indexing
	max     int    // maximum number of relocations
}

// New64M creates a Filter64M containing up to n uint64 items.
// A Filter64M contains a minimum of 2 items.
func New64M(n uint) *Filter64M {
	// each bucket holds bucketBitsNum/fpBitsNum fingerprint entries
	// of fpBitsNum bits each
	n2 := power2(n)
	if n2 < 2 {
		n2 = 2
	}

	// base the maximum numver of relocations on the number of buckets
	max := int(20 * math.Log(float64(n2)))

	return &Filter64M{
		buckets: make([]bucketM, n2, n2),
		mask:    uint64(n2 - 1),
		max:     max,
	}
}

// index1 computes the first index from the hash and the fingerprint.
func (cf *Filter64M) index1(h uint64, fp fingerprintM) uint64 {
	return h & cf.mask
}

// index2 computes the second index from the previous index and the fingerprint.
// Note that index2(index2(i, fp)) == i, which is leveraged when reverting relocation.
func (cf *Filter64M) index2(i uint64, fp fingerprintM) uint64 {
	return (i ^ fpHash(uint64(fp))) & cf.mask
}

// insertAt adds fingerprint fp to the filter at index i and returns
// if the insert succeeded.
func (cf *Filter64M) insertAt(idx uint64, fp fingerprintM) bool {
	u := cf.buckets[idx]
	for i := 0; i < bucketNumM; i++ {
		s := uint(i * fpBitsNumM)
		if u>>s&fpMaskM == 0 {
			cf.buckets[idx] |= bucketM(fp) << s
			return true
		}
	}
	return false
}

// Insert add an item to the Filter and returns whether it was inserted or not.
//
// NB. The same item cannot be inserted more than 2 times.
func (cf *Filter64M) Insert(x uint64) bool {
	return cf.insert(x, fpValue64M(x))
}

// insert adds an item with hash h and fingerprint fp.
func (cf *Filter64M) insert(h uint64, fp fingerprintM) bool {
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
		pfp := fingerprintM(cf.buckets[i] & fpMaskM)
		cf.buckets[i] &= ^bucketM(fpMaskM)
		cf.buckets[i] |= bucketM(fp)
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
		pfp := fingerprintM(cf.buckets[i] & fpMaskM)
		cf.buckets[i] &= ^bucketM(fpMaskM)
		cf.buckets[i] |= bucketM(fp)
		fp = pfp
	}

	return false
}

// hasIn checks bucket b for fingerprint fp and returns whether it was found.
func (cf *Filter64M) hasIn(b bucketM, fp fingerprintM) bool {
	for i := 0; i < bucketNumM; i++ {
		if b>>uint(i*fpBitsNumM)&fpMaskM == bucketM(fp) {
			return true
		}
	}
	return false
}

// Has checks if the item is in the Filter.
// If it returns false, then the item is definitely not in the filter.
// If it returns true, then the item *may* be in the filter, although
// with a low probability of not being in it.
func (cf *Filter64M) Has(x uint64) bool {
	return cf.has(x, fpValue64M(x))
}

// has checks if the item hash and fingerprint are found.
func (cf *Filter64M) has(h uint64, fp fingerprintM) bool {
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
func (cf *Filter64M) deleteAt(idx uint64, fp fingerprintM) bool {
	u := cf.buckets[idx]
	for i := 0; i < bucketNumM; i++ {
		s := uint(i * fpBitsNumM)
		if u>>s&fpMaskM == bucketM(fp) {
			cf.buckets[idx] &^= bucketM(fpMaskM) << s
			return true
		}
	}
	return false
}

// Delete removes an item from the Filter and returns whether or not it was present.
// To delete an item safely it must have been previously inserted.
func (cf *Filter64M) Delete(x uint64) bool {
	return cf.delete(x, fpValue64M(x))
}

// delete removes the item with the corresponding hash and fingerprint.
func (cf *Filter64M) delete(h uint64, fp fingerprintM) bool {
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
func (cf *Filter64M) Len() (n int) {
	for _, b := range cf.buckets {
		for ; b > 0; b >>= fpBitsNumM {
			if b&fpMaskM > 0 {
				n++
			}
		}
	}
	return
}

// Cap returns the filter capacity (maximum number of items it may contain).
func (cf *Filter64M) Cap() int {
	return len(cf.buckets)
}
