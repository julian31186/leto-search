package main

import (
	"fmt"
	"leto-search/search"
)

const resultLimit = 10
const searchPhrase = "Sandtrout"

// Ensure to only build the index if the index file is not populated. If it already is, read it into memory and use that

func main() {
	idx,err := search.BuildIndex()
	if err != nil {
		fmt.Println(err)
		return
	}

	results,err := search.Search(searchPhrase, idx, resultLimit)
	if err != nil {
		fmt.Println("Error searching")
		return
	}

	for _,res := range results {
		fmt.Println(res)
	}
}