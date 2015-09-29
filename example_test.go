package cuckoofilter_test

import (
	"fmt"
	"strings"

	"github.com/pierrec/go-cuckoofilter"
)

func Example_rowIds() {
	// some row ids
	rowIds := []uint64{24, 98, 58, 345, 111, 156800, 123, 961}

	// create a filter to hold the row ids and populate it
	f := cuckoofilter.New64S(uint(128))
	for _, id := range rowIds {
		if !f.Insert(id) {
			fmt.Printf("could not insert row id %d\n", id)
		}
	}

	// perform some checks
	fmt.Printf("number of items in the filter: %d\n", f.Len())
	fmt.Printf("item 123 is in the filter: %v\n", f.Has(123))
	fmt.Printf("item 456 is in the filter: %v\n", f.Has(456))

	// remove one item
	if !f.Delete(123) {
		fmt.Printf("could not remove row id %d\n", 123)
	}
	fmt.Printf("item 123 is in the filter: %v\n", f.Has(123))
	// Output:
	// number of items in the filter: 8
	// item 123 is in the filter: true
	// item 456 is in the filter: false
	// item 123 is in the filter: false
}

func Example_words() {
	// some words
	words := strings.Split(
		strings.Replace(
			"Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua",
			",",
			"",
			-1,
		),
		" ",
	)

	// create a filter to hold the words and populate it
	f := cuckoofilter.NewM(uint(len(words)))
	for _, w := range words {
		if !f.Insert([]byte(w)) {
			fmt.Printf("could not insert word %s\n", w)
		}
	}

	// perform some checks
	fmt.Printf("number of words in the filter: %d\n", f.Len())
	w := "elit"
	fmt.Printf("word '%s' is in the filter: %v\n", w, f.Has([]byte(w)))
	w = "e_lit"
	fmt.Printf("word '%s' is in the filter: %v\n", w, f.Has([]byte(w)))

	// remove one item
	w = "elit"
	if !f.Delete([]byte(w)) {
		fmt.Printf("could not remove word %s\n", w)
	}
	fmt.Printf("word '%s' is in the filter: %v\n", w, f.Has([]byte(w)))
	// Output:
	// number of words in the filter: 19
	// word 'elit' is in the filter: true
	// word 'e_lit' is in the filter: false
	// word 'elit' is in the filter: false
}
