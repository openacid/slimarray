package polyfit

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSolve(t *testing.T) {

	ta := assert.New(t)

	{
		x := []float64{0}
		solve1(
			[]float64{
				2,
			}, []float64{
				4,
			}, x)

		ta.Equal([]float64{2}, x)
	}

	{
		x := []float64{0, 0}

		solve2(
			[]float64{
				3, 5,
				1, 2,
			}, []float64{
				4,
				1,
			}, x)

		ta.Equal([]float64{3, -1}, x)
	}
	{
		x := []float64{0, 0, 0}

		solve3(
			[]float64{
				1, 3, -2,
				3, 5, 6,
				2, 4, 3,
			}, []float64{
				5,
				7,
				8,
			}, x)

		ta.Equal([]float64{-15, 8, 2}, x)
	}

}
