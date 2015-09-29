package cuckoofilter_test

import (
	"testing"

	cuckoo2 "github.com/seiflotfy/cuckoofilter"
	cuckoo3 "github.com/tylertreat/BoomFilters"
)

func BenchmarkExt1Insert(b *testing.B) {
	j, n := 0, len(data)
	f := cuckoo2.NewCuckooFilter(uint(n))
	for i := 0; i < b.N; i++ {
		f.Insert(data[j])
		if j++; j == n {
			j = 0
		}
	}
}

func BenchmarkExt1Has(b *testing.B) {
	f := *fullFilter2
	j, n := 0, int(f.GetCount())
	for i := 0; i < b.N; i++ {
		f.Lookup(data[j])
		if j++; j == n {
			j = 0
		}
	}
}

func BenchmarkExt2Insert(b *testing.B) {
	j, n := 0, len(data)
	f := cuckoo3.NewCuckooFilter(uint(n), 0.01)
	for i := 0; i < b.N; i++ {
		f.Add(data[j])
		if j++; j == n {
			j = 0
		}
	}
}

func BenchmarkExt2Has(b *testing.B) {
	f := *fullFilter3
	j, n := 0, int(f.Count())
	for i := 0; i < b.N; i++ {
		f.Test(data[j])
		if j++; j == n {
			j = 0
		}
	}
}
