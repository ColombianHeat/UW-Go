package iteration

import (
	"fmt"
	"testing"
)

func TestRepeat(t *testing.T) {
	t.Run("Repeating 'a' seven times", func(t *testing.T) {
		want := "aaaaaaa"
		got := Repeat("a", 7)

		if want != got {
			t.Errorf("expected %q but got %q", want, got)
		}
	})

	t.Run("Repeating '6' thirteen times", func(t *testing.T) {
		want := "6666666666666"
		got := Repeat("6", 13)

		if want != got {
			t.Errorf("expected %q but got %q", want, got)
		}
	})

}

func BenchmarkRepeat(b *testing.B) {
	for i:=0; i < b.N; i++ {
		Repeat("a", 20)
	}
}

func ExampleRepeat() {
	var repeated string = Repeat("y", 4)
	fmt.Println(repeated)
	// Output: yyyy
}