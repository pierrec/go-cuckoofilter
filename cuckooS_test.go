package cuckoofilter_test

import (
	//	"fmt"
	"sync"
	"testing"

	"github.com/pierrec/go-cuckoofilter"
)

var (
	fullFilter64S *cuckoofilter.Filter64S
	fullFilterS   *cuckoofilter.FilterS
)

func initFilterS(wg *sync.WaitGroup) {
	j, n := 0, len(data)
	fullFilter64S = cuckoofilter.New64S(uint(n))
	fullFilterS = cuckoofilter.NewS(uint(n))
	for i := 0; i < n; i++ {
		fullFilterS.Insert(data[j])
		fullFilter64S.Insert(data64[j])
		if j++; j == n {
			j = 0
		}
	}
	wg.Done()
}

func TestLenCapS(t *testing.T) {
	f := cuckoofilter.New64S(100)
	if n := f.Len(); n != 0 {
		t.Errorf("expected filter len to be %d, got %d", 0, n)
		t.FailNow()
	}
	if c := f.Cap(); c != 128 {
		t.Errorf("expected filter cap to be %d, got %d", 128, c)
		t.FailNow()
	}

	f = cuckoofilter.New64S(0)
	if n := f.Len(); n != 0 {
		t.Errorf("expected filter len to be %d, got %d", 0, n)
		t.FailNow()
	}
	if c := f.Cap(); c != 2 {
		t.Errorf("expected filter cap to be %d, got %d", 2, c)
		t.FailNow()
	}
}

func TestInsertHasS(t *testing.T) {
	f := cuckoofilter.New64S(128)

	// empty filter
	for i := 0; i < f.Len(); i++ {
		if f.Has(uint64(i)) {
			t.Errorf("%d should not be a member", i)
			t.FailNow()
		}
	}

	// insert 12 into the filter
	if !f.Insert(12) {
		t.Error("insert failed")
		t.FailNow()
	}

	if !f.Has(12) {
		t.Error("item should be in the filter")
		t.FailNow()
	}

	// reinsert 12
	if !f.Insert(12) {
		t.Error("reinsert failed")
		t.FailNow()
	}

	if !f.Has(12) {
		t.Error("item should be in the filter after reinsert")
		t.FailNow()
	}

	// 123 should not be in the filter
	if f.Has(123) {
		t.Error("item should not be in the filter")
		t.FailNow()
	}

	// insert 123 into the filter
	if !f.Insert(123) {
		t.Error("insert failed")
		t.FailNow()
	}

	// 123 should be in the filter as well as 12
	if !f.Has(123) {
		t.Error("item should be in the filter")
		t.FailNow()
	}
	if !f.Has(12) {
		t.Error("item should be in the filter")
		t.FailNow()
	}

	// full filter
	for i := 0; i < 16*f.Len(); i++ {
		f.Insert(uint64(i))
	}

	if f.Insert(123) {
		t.Error("expected failed insert on full filter")
	}

	if n, c := f.Len(), f.Cap(); n < c {
		t.Errorf("filter should be full: len=%d / cap=%d", n, c)
		t.FailNow()
	}
}

func TestDeleteHasS(t *testing.T) {
	f := cuckoofilter.New64S(128)

	// insert and delete an item
	f.Insert(123)
	if !f.Has(123) {
		t.Error("item should be in the filter")
		t.FailNow()
	}
	if f.Delete(456) {
		t.Error("item should not be in the filter")
		t.FailNow()
	}
	if !f.Delete(123) {
		t.Error("item should be in the filter")
		t.FailNow()
	}
	if f.Has(123) {
		t.Error("item should not be in the filter")
		t.FailNow()
	}

	// insert 2 items and delete another one
	f.Insert(123)
	f.Insert(456)
	if !f.Has(123) || !f.Has(456) {
		t.Error("item should be in the filter")
		t.FailNow()
	}
	if f.Delete(789) {
		t.Error("item should not be in the filter")
		t.FailNow()
	}

	// now delete the items
	if !f.Delete(456) {
		t.Error("item should be in the filter")
		t.FailNow()
	}
	if f.Has(456) {
		t.Error("item should not be in the filter")
		t.FailNow()
	}
	if !f.Delete(123) {
		t.Error("item should be in the filter")
		t.FailNow()
	}
	if f.Has(123) {
		t.Error("item should not be in the filter")
		t.FailNow()
	}
}

func TestFilterS(t *testing.T) {
	f := cuckoofilter.NewS(uint(len(data)))

	if n := f.Cap(); n < len(data) {
		t.Errorf("invalid filter capacity: got %d, expected >= %d", n, len(data))
		t.FailNow()
	}

	// insert the random data items into the filter
	// removing the ones that failed.
	success := 0
	for i, w := range data {
		if f.Insert(w) {
			success++
		} else {
			data[i] = nil
		}
	}

	if n := f.Len(); n != success {
		t.Errorf("invalid filter lengh: got %d, expected %d", n, success)
		t.FailNow()
	}

	// check data items are in the filter
	errNum := 0
	for _, w := range data {
		if w == nil {
			// this item could not be inserted, no need to check it
			continue
		}
		// all items that were successfully inserted must be found
		if !f.Has(w) {
			t.Errorf("word %v not found", w)
			t.FailNow()
		}
		// w has 16 bytes max, make sure the one we test
		// has more so that we know it cannot be in the set.
		w = append([]byte{}, w...)
		w = append(w, []byte("0123456789ABCDEF")...)
		// false positive
		if f.Has(w) {
			errNum++
		}
	}
	// check the error rate
	errRate := 100 * float64(errNum) / float64(success)
	//fmt.Printf("S: ok=%d err=%d errRate=%.5f%%\n", success, errNum, errRate)
	if errRate > 11 {
		t.Errorf("error rate too high: %.5f%% = %d / %d", errRate, errNum, len(data))
	}

	// remove items from the filter
	for _, w := range data {
		if w == nil {
			// this item could not be inserted, no need to check it
			continue
		}
		if !f.Delete(w) {
			t.Errorf("word %v not found", w)
			t.FailNow()
		}
	}
}
