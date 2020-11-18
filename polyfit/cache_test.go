package polyfit

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitCache2(t *testing.T) {

	ta := assert.New(t)

	xs := make([]float64, 1024)
	ys := make([]float64, 1024)
	for i := 0; i < 1024; i++ {
		xs[i] = float64(i)
	}

	for i := 1; i < 1024+1; i++ {
		f := NewFit(xs[:i], ys[:i], 2)
		ta.Equal(XTXCache3[i], f.xtx)
	}

}

func TestGetCachedXTX(t *testing.T) {

	ta := assert.New(t)

	xs := make([]float64, 1024)
	ys := make([]float64, 1024)
	for i := 0; i < 1024; i++ {
		xs[i] = float64(i)
	}

	rst := make([]float64, 9)
	for i := 0; i < 1024+1; i++ {
		for j := i; j < i+5 && j < 1024+1; j++ {
			f := NewFit(xs[i:j], ys[i:j], 2)
			getCachedXTX3(i, j, rst)
			ta.Equal(rst, f.xtx)
		}
	}
}

func TestNewIntRange(t *testing.T) {

	ta := assert.New(t)

	xs := make([]float64, 1024)
	ys := make([]float64, 1024)
	for i := 0; i < 1024; i++ {
		xs[i] = float64(i)
	}

	for i := 0; i < 1024+1; i++ {
		for j := i; j < i+5 && j < 1024+1; j++ {
			{
				// cached degree 2:
				f := NewFit(xs[i:j], ys[i:j], 2)
				fint := NewFitIntRange(i, j, ys[i:j], 2)
				ta.Equal(f, fint)
			}

			{
				// non-cached degree:
				f := NewFit(xs[i:j], ys[i:j], 3)
				fint := NewFitIntRange(i, j, ys[i:j], 3)
				ta.Equal(f, fint)
			}
		}
	}
}
