[![godoc](https://godoc.org/github.com/pierrec/go-cuckoofilter?status.png)](https://godoc.org/github.com/pierrec/go-cuckoofilter)
[![Build Status](https://travis-ci.org/pierrec/go-cuckoofilter.svg?branch=master)](https://travis-ci.org/pierrec/go-cuckoofilter)

Package cuckoofilter implements the Cuckoo Filter algorithm for approximated
set-membership queries. See https://www.cs.cmu.edu/~dga/papers/cuckoo-conext2014.pdf for the theory.

Six types are defined depending on the use cases:

|Filter Type Name|Input Type|memory usage (per item)|error rate (%)|
|----------------|----------|-----------------------|--------------|
|Filter64S|uint64|1 byte|~11    |
|Filter64M|uint64|2 bytes|~1    |
|Filter64L|uint64|8 bytes|~0.003|
|FilterS  |[]byte|1 byte |~11   |
|FilterM  |[]byte|2 bytes|~1    |
|FilterL  |[]byte|8 bytes|~0.003|

Filter64* types are much faster then their siblings, see the benchmarks below.

The following benchmarks were performed on randomly generated items of sizes up to 16 bytes.

Benchmarks for FilterL:

|Benchmark Name | # tests | time/op | bits/op | allocations/op|
|---------------|---------|---------|---------|---------------|
|Benchmark64InsertL-4 |	 2000000|	      1283 ns/op|	       4 B/op|	       0 allocs/op|
|Benchmark64HasL-4    |	30000000|	        39.4 ns/op|	       0 B/op|	       0 allocs/op|
|BenchmarkInsertL-4   |	 1000000|	      1060 ns/op|	       8 B/op|	       0 allocs/op|
|BenchmarkHasL-4      |	20000000|	       101 ns/op|	       0 B/op|	       0 allocs/op|

Benchmarks for FilterM:

|Benchmark Name | # tests | time/op | bits/op | allocations/op|
|---------------|---------|---------|---------|---------------|
|Benchmark64InsertM-4 |	 2000000|	      2152 ns/op|	       1 B/op|	       0 allocs/op|
|Benchmark64HasM-4    |	50000000|	        22.6 ns/op|	       0 B/op|	       0 allocs/op|
|BenchmarkInsertM-4   |	 2000000|	      1691 ns/op|	       1 B/op|	       0 allocs/op|
|BenchmarkHasM-4      |	20000000|	        79.0 ns/op|	       0 B/op|	       0 allocs/op|

Benchmarks for FilterS:

|Benchmark Name | # tests | time/op | bits/op | allocations/op|
|---------------|---------|---------|---------|---------------|
|Benchmark64InsertS-4 |	 1000000|	      1077 ns/op|	       1 B/op|	       0 allocs/op|
|Benchmark64HasS-4    |	100000000|	        16.4 ns/op|	       0 B/op|	       0 allocs/op|
|BenchmarkInsertS-4   |	 1000000|	      1004 ns/op|	       1 B/op|	       0 allocs/op|
|BenchmarkHasS-4      |	20000000|	        65.6 ns/op|	       0 B/op|	       0 allocs/op|

Benchmarks for github.com/seiflotfy/cuckoofilter:

|Benchmark Name | # tests | time/op | bits/op | allocations/op|
|---------------|---------|---------|---------|---------------|
|BenchmarkExt1Insert-4|	 1000000|	      4950 ns/op|	      17 B/op|	       1 allocs/op|
|BenchmarkExt1Has-4   |	10000000|	       175 ns/op|	      16 B/op|	       1 allocs/op|

Benchmarks for github.com/tylertreat/BoomFilters:

|Benchmark Name | # tests | time/op | bits/op | allocations/op|
|---------------|---------|---------|---------|---------------|
|BenchmarkExt2Insert-4|	       1|	1656973329 ns/op|	1006633088 B/op|	 8388613 allocs/op|
|BenchmarkExt2Has-4   |	 3000000|	       455 ns/op|	      32 B/op|	       2 allocs/op|
