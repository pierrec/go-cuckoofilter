package cuckoofilter_test

import (
	"testing"

	"github.com/pierrec/go-cuckoofilter"
)

func Benchmark64InsertS(b *testing.B) {
	j, n := 0, len(data)
	f := cuckoofilter.New64S(uint(n))
	for i := 0; i < b.N; i++ {
		f.Insert(data64[j])
		if j++; j == n {
			j = 0
		}
	}
}

func Benchmark64HasS(b *testing.B) {
	f := *fullFilter64S
	j, n := 0, f.Len()
	for i := 0; i < b.N; i++ {
		f.Has(data64[j])
		if j++; j == n {
			j = 0
		}
	}
}

func BenchmarkInsertS(b *testing.B) {
	j, n := 0, len(data)
	f := cuckoofilter.NewS(uint(n))
	for i := 0; i < b.N; i++ {
		f.Insert(data[j])
		if j++; j == n {
			j = 0
		}
	}
}

func BenchmarkHasS(b *testing.B) {
	f := *fullFilterS
	j, n := 0, f.Len()
	for i := 0; i < b.N; i++ {
		f.Has(data[j])
		if j++; j == n {
			j = 0
		}
	}
}
