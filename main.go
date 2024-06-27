package main

import (
	"fmt"
	"strings"

	"golang.org/x/tour/wc"
)

func WordCount(s string) map[string]int {
	word_map := make(map[string]int)
	var string_arr = strings.Fields(s)
	for _, word := range string_arr {
		if _, ok := word_map[word]; !ok {
		word_map[word] = 0
		} else {
		word_map[word] += 1
		}
		fmt.Println(word)
	}
	return word_map
}

func main() {
	wc.Test(WordCount)
}