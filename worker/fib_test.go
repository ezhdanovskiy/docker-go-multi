package worker

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFib(t *testing.T) {
	var tests = []struct {
		in, out int64
	}{
		{
			in:  0,
			out: 1,
		},
		{
			in:  1,
			out: 1,
		},
		{
			in:  2,
			out: 2,
		},
		{
			in:  3,
			out: 3,
		},
		{
			in:  4,
			out: 5,
		},
		{
			in:  5,
			out: 8,
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprint(tt.in), func(t *testing.T) {
			assert.Equal(t, tt.out, fib(tt.in))
		})
	}
}
