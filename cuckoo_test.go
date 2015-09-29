package cuckoofilter_test

import (
	crand "crypto/rand"
	"flag"
	"fmt"
	"math/rand"
	"regexp"
	"sync"

	cuckoo2 "github.com/seiflotfy/cuckoofilter"
	cuckoo3 "github.com/tylertreat/BoomFilters"
)

var (
	data        [][]byte
	data64      []uint64
	fullFilter2 *cuckoo2.CuckooFilter
	fullFilter3 *cuckoo3.CuckooFilter
)

func init() {
	// generate 2 random data sets with 1million entries:
	// one with uint64 items, the other with []byte items.
	maxSize := 16 << 20
	randomData := make([]byte, maxSize)
	if _, err := crand.Read(randomData); err != nil {
		panic(fmt.Sprintf("cannot initialize random data for size %d", maxSize))
	}
	// []byte random data set
	for i, j, n := 0, 0, 1000000; j < len(randomData) && i < n; i++ {
		bn := rand.Intn(16)
		// fmt.Printf("%d: j=%d bn=%d \n", len(randomData), j, bn)
		data = append(data, randomData[j:j+bn])
		j += bn
	}

	// populate the random data sets and populate S, M and L filters with their data.
	// uint64 random data set
	j, n := 0, len(data)
	for i := 0; i < n; i++ {
		var u uint64
		for i := 0; i < 8 && i < len(data[j]); i++ {
			u |= uint64(data[j][i]) << uint(i*8)
		}
		data64 = append(data64, u)
		if j++; j == n {
			j = 0
		}
	}
	// populate the filters
	var wg sync.WaitGroup
	wg.Add(3)
	go initFilterS(&wg)
	go initFilterM(&wg)
	go initFilterL(&wg)

	// populate external filters if required
	// check the bench command line argument to go test
	// and check if it matches the BenchmarkExt function names
	flag.Parse()
	bench := flag.CommandLine.Lookup("test.bench")
	if v := bench.Value.String(); v != "" {
		rex, _ := regexp.Compile(v)
		if rex.MatchString("Ext1Has") {
			wg.Add(1)
			go initExt1(&wg)
		}
		if rex.MatchString("Ext2Has") {
			wg.Add(1)
			go initExt2(&wg)
		}
	}

	wg.Wait()
}

func initExt1(wg *sync.WaitGroup) {
	j, n := 0, len(data)
	fullFilter2 = cuckoo2.NewCuckooFilter(uint(n))
	for i := 0; i < n; i++ {
		fullFilter2.Insert(data[j])
		if j++; j == n {
			j = 0
		}
	}
	wg.Done()
}

func initExt2(wg *sync.WaitGroup) {
	j, n := 0, len(data)
	fullFilter3 = cuckoo3.NewCuckooFilter(uint(n), 0.01)
	for i := 0; i < n; i++ {
		fullFilter3.Add(data[j])
		if j++; j == n {
			j = 0
		}
	}
	wg.Done()
}
