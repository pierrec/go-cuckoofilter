package cuckoofilter_test

import (
	"testing"

	"github.com/pierrec/go-cuckoofilter"
)

func Benchmark64InsertL(b *testing.B) {
	j, n := 0, len(data)
	f := cuckoofilter.New64L(uint(n))
	for i := 0; i < b.N; i++ {
		f.Insert(data64[j])
		if j++; j == n {
			j = 0
		}
	}
}

func Benchmark64HasL(b *testing.B) {
	f := *fullFilter64L
	j, n := 0, f.Len()
	for i := 0; i < b.N; i++ {
		f.Has(data64[j])
		if j++; j == n {
			j = 0
		}
	}
}

func BenchmarkInsertL(b *testing.B) {
	j, n := 0, len(data)
	f := cuckoofilter.NewL(uint(n))
	for i := 0; i < b.N; i++ {
		f.Insert(data[j])
		if j++; j == n {
			j = 0
		}
	}
}

func BenchmarkHasL(b *testing.B) {
	f := *fullFilterL
	j, n := 0, f.Len()
	for i := 0; i < b.N; i++ {
		f.Has(data[j])
		if j++; j == n {
			j = 0
		}
	}
}
