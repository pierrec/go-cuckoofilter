package cuckoofilter_test

import (
	"testing"

	"github.com/pierrec/go-cuckoofilter"
)

func Benchmark64InsertM(b *testing.B) {
	j, n := 0, len(data)
	f := cuckoofilter.New64M(uint(n))
	for i := 0; i < b.N; i++ {
		f.Insert(data64[j])
		if j++; j == n {
			j = 0
		}
	}
}

func Benchmark64HasM(b *testing.B) {
	f := *fullFilter64M
	j, n := 0, f.Len()
	for i := 0; i < b.N; i++ {
		f.Has(data64[j])
		if j++; j == n {
			j = 0
		}
	}
}

func BenchmarkInsertM(b *testing.B) {
	j, n := 0, len(data)
	f := cuckoofilter.NewM(uint(n))
	for i := 0; i < b.N; i++ {
		f.Insert(data[j])
		if j++; j == n {
			j = 0
		}
	}
}

func BenchmarkHasM(b *testing.B) {
	f := *fullFilterM
	j, n := 0, f.Len()
	for i := 0; i < b.N; i++ {
		f.Has(data[j])
		if j++; j == n {
			j = 0
		}
	}
}
