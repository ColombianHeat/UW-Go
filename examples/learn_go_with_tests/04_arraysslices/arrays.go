package main

func Sum(arr []int) int {
	var sum int = 0
	for _, number := range arr {
		sum += number
	}
	return sum
}

func SumAllTails(numsToSum ...[]int) []int {
	var sums []int
	for _, numbers := range numsToSum {
		if len(numbers) == 0 {
			sums = append(sums, 0)
		} else {
			tail := numbers[1:]
			sums = append(sums, Sum(tail))
		}
	}
	return sums
}