package polyfit_test

import (
	"fmt"

	. "github.com/openacid/polyarray/polyfit"
)

func Example() {

	// Fit 4 points with a polynomial of degree=2,
	// the result should be:
	// y = 6.2 - 0.7x + 0.2x²
	//                                    .
	//                                  ..
	//                                 .
	//                         (3,7) ..(4,7)
	//                             ..
	//                           ..
	// .                       ..
	//  ...    (1,6)        ...
	//     ......     ......
	//           .....
	//
	//                 (2,5)

	xs := []float64{1, 2, 3, 4}
	ys := []float64{6, 5, 7, 7}

	f := NewFitting(xs, ys, 2)
	poly := f.Solve(true)

	fmt.Printf("y = %.1f + %.1fx + %.1fx²\n",
		poly[0], poly[1], poly[2])

	for i, x := range xs {
		fmt.Printf("point[%d]=(%.0f, %.0f), evaluated=(%.0f, %.1f)\n",
			i, x, ys[i], x, evalPoly(poly, x))
	}

	// Output:
	// y = 6.2 + -0.7x + 0.2x²
	// point[0]=(1, 6), evaluated=(1, 5.7)
	// point[1]=(2, 5), evaluated=(2, 5.8)
	// point[2]=(3, 7), evaluated=(3, 6.3)
	// point[3]=(4, 7), evaluated=(4, 7.2)

}

func evalPoly(poly []float64, x float64) float64 {
	rst := float64(0)
	pow := float64(1)
	for _, coef := range poly {
		rst += coef * pow
		pow *= x
	}

	return rst
}
