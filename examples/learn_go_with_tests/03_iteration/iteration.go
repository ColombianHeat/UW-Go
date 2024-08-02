package iteration

func Repeat(char string, n_chars int) string {
	var repeated string
	for i := 0; i < n_chars; i++ {
		repeated += char
	}
	return repeated
}
