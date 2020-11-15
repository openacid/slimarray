package slimarray_test

import (
	"fmt"
	"math/rand"
	"sort"

	"github.com/openacid/slimarray"
)

func ExampleSlimArray_Stat() {

	fmt.Println("== Memory cost stats of sorted random uint array ==")

	cases := []struct {
		n   int
		rng uint32
	}{
		{1000, 1000},
		{1000 * 1000, 1000 * 1000},
		{1000 * 1000, 1000 * 1000 * 1000},
	}

	for _, c := range cases {
		n := c.n
		rng := c.rng

		nums := []uint32{}
		rnd := rand.New(rand.NewSource(int64(n) * int64(rng)))

		for i := 0; i < n; i++ {
			s := uint32(rnd.Float64() * float64(rng))
			nums = append(nums, s)
		}

		sort.Slice(nums, func(i, j int) bool { return nums[i] < nums[j] })

		a := slimarray.NewU32(nums)

		st := a.Stat()
		fmt.Printf("\nn=%d rng=[0, %d]:\n\n", n, rng)

		for _, k := range []string{
			// "mem_elts", "span_cnt", "spans/seg", "elt_width",
			"n", "mem_total", "bits/elt",
		} {
			fmt.Printf("  %10s: %d\n", k, st[k])
		}
	}

	// Output:
	// == Memory cost stats of sorted random uint array ==
	//
	// n=1000 rng=[0, 1000]:
	//
	//            n: 1000
	//    mem_total: 856
	//     bits/elt: 6
	//
	// n=1000000 rng=[0, 1000000]:
	//
	//            n: 1000000
	//    mem_total: 705720
	//     bits/elt: 5
	//
	// n=1000000 rng=[0, 1000000000]:
	//
	//            n: 1000000
	//    mem_total: 2078336
	//     bits/elt: 16
}
