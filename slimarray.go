// package slimarray uses polynomial to compress and store an array of uint32.
// A uint32 costs only 5 bits in a sorted array of a million number in range [0,
// 1000*1000].
//
// The General Idea
//
// We use a polynomial y = a + bx + cx² to describe the overall trend of the
// numbers.
// And for every number i we add a residual to fit the gap between y(i) and
// nums[i].
// E.g. If there are 4 numbers: 0, 15, 33, 50
// The polynomial and residuals are:
//     y = 16x
//     0, -1, 1, 2
//
// In this case the residuals require 3 bits for each of them.
// To retrieve the numbers, we evaluate y(i) and add the residual to it:
//     get(0) = y(0) + 0 = 16 * 0 + 0 = 0
//     get(1) = y(1) - 1 = 16 * 1 - 1 = 15
//     get(2) = y(2) + 1 = 16 * 2 + 1 = 33
//     get(3) = y(3) + 2 = 16 * 3 + 2 = 50
//
//
// What It Is And What It Is Not
//
// Another space efficient data structure to store uint32 array is trie or prefix
// tree or radix tree. It is possible to use bitmap-based btree like structure
// to reduce space(very likely in such case it provides higher compression rate).
// But it requires the array to be sorted.
//
// SlimArray does not have such restriction. It is more adaptive with data
// layout. To achieve high compression rate, it only requires the data has a
// overall trend, e.g., roughly sorted, as seen in the above 4 integers
// examples. Additionally, it also accept duplicated element in the array, which
// a bitmap based or tree-like data structure does not allow.
//
//
// Data Structure
//
// SlimArray splits the entire array into segments(Seg),
// each of which has 1024 numbers.
// And then it splits every segment into several spans.
// Every span has its own polynomial. A span has 16*k numbers.
// A segment has at most 64 spans.
//
//           seg[0]                      seg[1]
//           1024 nums                   1024 nums
//   |-------+---------------+---|---------------------------|...
//    span[0]    span[1]
//    16 nums    32 nums      ..
//
//
// Uncompressed Data Structures
//
// A SlimArray is a compacted data structure.
// The original data structures are defined as follow(assumes original user data
// is `nums []uint32`):
//
//   Seg struct {
//     SpansBitmap   uint64      // describe span layout
//     Rank         uint64      // count `1` in preceding Seg.
//     Spans       []Span
//   }
//
//   Span struct {
//     width         int32       // is retrieved from SpansBitmap
//
//     Polynomial [3]double      //
//     Config struct {           //
//       Offset        int32     // residual offset
//       ResidualWidth int32     // number of bits a residual requires
//     }
//     Residuals  [width][ResidualWidth]bit // pack into SlimArray.Residuals
//   }
//
// A span stores 16*k int32 in it, where k ∈ [1, 64).
//
// `Seg.SpansBitmap` describes the layout of Span-s in a Seg.
// A "1" at i-th bit and a "1" at j-th bit means a Span stores
// `nums[i*16:j*16]`, e.g.:
//
//     100101110000......
//     <-- least significant bit
//
// In the above example:
//
//     span[0] has 16*3 nums in it.
//     span[1] has 16*2 nums in it.
//     span[2] has 16*1 nums in it.
//
// `Seg.Rank` caches the total count of "1" in all preceding Seg.SpansBitmap.
// This accelerate locating a Span in the packed field SlimArray.Polynomials .
//
// `Span.width` is the count of numbers stored in this span.
// It does not need to be stored because it can be calculated by counting the
// "0" between two "1" in `Seg.SpansBitmap`.
//
// `Span.Polynomial` stores 3 coefficients of the polynomial describing the
// overall trend of this span. I.e. the `[a, b, c]` in `y = a + bx + cx²`
//
// `Span.Config.Offset` adjust the offset to locate a residual.
// In a span we want to have that:
//		residual position = Config.Offset + (i%1024) * Config.ResidualWidth
//
// But if the preceding span has smaller residual width, the "offset" could be
// negative, e.g.: span[0] has residual of width 0 and 16 residuals,
// span[1] has residual of width 4.
// Then the "offset" of span[1] is `-16*4` in order to satisfy:
// `(-16*4) + i * 4` is the correct residual position, for i in [16, 32).
//
// `Span.Config.ResidualWidth` specifies the number of bits to
// store every residual in this span, it must be a power of 2: `2^k`.
//
// `Span.Residuals` is an array of residuals of length `Span.width`.
// Every elt in it is a `ResidualWidth`-bits integers.
//
// Compact
//
// SlimArray compact `Seg` into a dense format:
//
//   SlimArray.Bitmap = [
//     Seg[0].SpansBitmap,
//     Seg[1].SpansBitmap,
//     ... ]
//
//   SlimArray.Polynomials = [
//     Seg[0].Spans[0].Polynomials,
//     Seg[0].Spans[1].Polynomials,
//     ...
//     Seg[1].Spans[0].Polynomials,
//     Seg[1].Spans[1].Polynomials,
//     ...
//   ]
//
//   SlimArray.Configs = [
//     Seg[0].Spans[0].Config
//     Seg[0].Spans[1].Config
//     ...
//     Seg[1].Spans[0].Config
//     Seg[1].Spans[1].Config
//     ...
//   ]
//
// `SlimArray.Residuals` simply packs the residuals of every nums[i] together.
package slimarray

import (
	"fmt"
	"math"
	"math/bits"

	"github.com/openacid/low/bitmap"
	"github.com/openacid/low/size"
	"github.com/openacid/slimarray/polyfit"
)

const (
	// The smallest span is 16 numbers.
	// Two adjacent span will be merged into one if the result span costs less
	// memory.
	minSpan = int32(16)

	// Segment size. A segment consists of at most 64 spans.
	segSize = 1024

	// log(2, segSize) to speed up calc.
	segSizeShift = uint(10)
	segSizeMask  = int32(1024 - 1)

	// Degree of polynomial to describe overall trend in a span.
	// We always use a polynomial of degree 2: y = a + bx + cx²
	polyDegree = 2

	// Count of coefficients of a polynomial.
	polyCoefCnt = polyDegree + 1
)

// evalPoly2 evaluates a polynomial with degree=2.
//
// Since 0.1.1
func evalPoly2(poly []float64, x float64) float64 {
	return poly[0] + poly[1]*x + poly[2]*x*x
}

// NewU32 creates a "SlimArray" array from a slice of uint32.
//
// A NewU32() costs about 500 ns/elt.
//
// Since 0.1.1
func NewU32(nums []uint32) *SlimArray {

	pa := &SlimArray{
		N: int32(len(nums)),
	}

	for ; len(nums) > segSize; nums = nums[segSize:] {
		pa.addSeg(nums[:segSize])
	}
	if len(nums) > 0 {
		pa.addSeg(nums)
	}

	// shrink capacity to len.
	pa.Rank = append(pa.Rank[:0:0], pa.Rank...)
	pa.Bitmap = append(pa.Bitmap[:0:0], pa.Bitmap...)
	pa.Polynomials = append(pa.Polynomials[:0:0], pa.Polynomials...)
	pa.Configs = append(pa.Configs[:0:0], pa.Configs...)

	// Add another empty word to avoid panic for residual of width = 0.
	pa.Residuals = append(pa.Residuals, 0)
	pa.Residuals = append(pa.Residuals[:0:0], pa.Residuals...)

	return pa
}

// Get returns the uncompressed uint32 value.
// A Get() costs about 7 ns
//
// Since 0.1.1
func (sm *SlimArray) Get(i int32) uint32 {

	// The index of a segment
	bitmapI := i >> segSizeShift
	spansBitmap := sm.Bitmap[bitmapI]
	rank := sm.Rank[bitmapI]

	i = i & segSizeMask
	x := float64(i)

	// i>>4 is in-segment span index
	bm := spansBitmap & bitmap.Mask[i>>4]
	spanIdx := int(rank) + bits.OnesCount64(bm)

	// eval y = a + bx + cx²

	j := spanIdx * polyCoefCnt
	p := sm.Polynomials
	v := int64(p[j] + p[j+1]*x + p[j+2]*x*x)

	config := sm.Configs[spanIdx]
	residualWidth := config & 0xff
	offset := config >> 8

	// where the residual is
	resBitIdx := offset + int64(i)*residualWidth

	// extract residual from packed []uint64
	d := sm.Residuals[resBitIdx>>6]
	d = d >> uint(resBitIdx&63)

	return uint32(v + int64(d&bitmap.Mask[residualWidth]))
}

// Len returns number of elements.
//
// Since 0.1.1
func (sm *SlimArray) Len() int {
	return int(sm.N)
}

// Stat returns a map describing memory usage.
//
//    seg_cnt   :512         // segment count
//    elt_width :8           // average bits count per elt
//    span_cnt  :12          // total count of spans
//    spans/seg :7           // average span count per segment
//    mem_elts  :1048576     // memory cost for residuals
//    mem_total :1195245     // total memory cost
//    bits/elt  :9           // average memory cost per elt
//    n         :10          // total elt count
//
// Since 0.1.1
func (sm *SlimArray) Stat() map[string]int32 {
	segCnt := len(sm.Bitmap)
	totalmem := size.Of(sm)

	spanCnt := len(sm.Polynomials) / 3
	memWords := len(sm.Residuals) * 8
	widthAvg := 0
	for i := 0; i < spanCnt; i++ {
		w := sm.Configs[i] & 0xff
		widthAvg += int(w)
	}

	n := sm.Len()
	if n == 0 {
		n = 1
	}

	if spanCnt == 0 {
		spanCnt = 1
	}

	st := map[string]int32{
		"seg_cnt":   int32(segCnt),
		"elt_width": int32(widthAvg / spanCnt),
		"mem_total": int32(totalmem),
		"mem_elts":  int32(memWords),
		"bits/elt":  int32(totalmem * 8 / n),
		"spans/seg": int32((spanCnt * 1000) / (segCnt*1000 + 1)),
		"span_cnt":  int32(spanCnt),
		"n":         sm.N,
	}

	return st
}

func (sm *SlimArray) addSeg(nums []uint32) {

	bm, polynomials, configs, words := newSeg(nums, int64(len(sm.Residuals)*64))

	var r uint64
	l := len(sm.Rank)
	if l > 0 {
		r = sm.Rank[l-1] + uint64(bits.OnesCount64(sm.Bitmap[l-1]))
	} else {
		r = 0
	}

	sm.Bitmap = append(sm.Bitmap, bm)
	sm.Rank = append(sm.Rank, r)
	sm.Polynomials = append(sm.Polynomials, polynomials...)
	sm.Configs = append(sm.Configs, configs...)
	sm.Residuals = append(sm.Residuals, words...)
}

func newSeg(nums []uint32, start int64) (uint64, []float64, []int64, []uint64) {

	n := int32(len(nums))
	xs := make([]float64, n)
	ys := make([]float64, n)

	for i, v := range nums {
		xs[i] = float64(i)
		ys[i] = float64(v)
	}

	// create polynomial fit sessions for every 16 numbers
	fts := initFittings(xs, ys, minSpan)

	spans := findMinFittingsNew(xs, ys, fts)

	polynomials := make([]float64, 0)
	configs := make([]int64, 0)
	words := make([]uint64, n) // max size

	// Using a bitmap to describe which spans a polynomial spans
	segPolyBitmap := uint64(0)

	resI := int64(0)

	for _, sp := range spans {

		// every poly starts at 16*k th point
		segPolyBitmap |= 1 << uint((sp.e-1)>>4)

		width := sp.residualWidth
		if width > 0 {
			resI = resI + int64(width) - 1
			resI -= resI % int64(width)
		}

		polynomials = append(polynomials, sp.poly...)

		// We want eltIndex = stBySeg + i * residualWidth
		// min of stBySeg is -segmentSize * residualWidth = -1024 * 16;
		// Add this value to make it a positive number.
		offset := resI + start - int64(sp.s)*int64(width)
		config := offset<<8 | int64(width)
		configs = append(configs, config)

		for j := sp.s; j < sp.e; j++ {

			v := evalPoly2(sp.poly, xs[j])

			// It may overflow but the result is correct because (a+b) % p =
			// (a%p + b%p) % p
			d := uint32(int64(nums[j]) - int64(v))

			wordI := resI >> 6
			words[wordI] |= uint64(d) << uint(resI&63)

			resI += int64(width)
		}
	}

	nWords := (resI + 63) >> 6

	return segPolyBitmap, polynomials, configs, words[:nWords]
}

func initFittings(xs, ys []float64, spanSize int32) []*polyfit.Fitting {

	fts := make([]*polyfit.Fitting, 0)
	n := int32(len(xs))

	for i := int32(0); i < n; i += spanSize {
		s := i
		e := s + spanSize
		if e > n {
			e = n
		}

		ft := polyfit.NewFitting(xs[s:e], ys[s:e], polyDegree)
		fts = append(fts, ft)
	}
	return fts
}

type span struct {
	ft            *polyfit.Fitting
	origPoly      []float64
	poly          []float64
	residualWidth uint32
	mem           int

	// start and end index in original []int32
	s, e int32
}

func (sp *span) Copy() *span {
	b := &span{
		ft:       sp.ft.Copy(),
		origPoly: make([]float64, 0, len(sp.origPoly)),
		poly:     make([]float64, 0, len(sp.poly)),
		mem:      sp.mem,
		s:        sp.s,
		e:        sp.e,
	}

	b.origPoly = append(b.origPoly, sp.origPoly...)
	b.poly = append(b.poly, sp.poly...)
	return b
}

func (sp *span) String() string {
	return fmt.Sprintf("%d-%d(%d): width: %d, mem: %d, poly: %v",
		sp.s, sp.e, sp.e-sp.s, sp.residualWidth, sp.mem, sp.poly)
}

// findMinFittingsNew by merge adjacent 16-numbers span.
// If two spans has a common trend they should be described with one polynomial.
func findMinFittingsNew(xs, ys []float64, fts []*polyfit.Fitting) []*span {

	spans := make([]*span, len(fts))
	merged := make([]*span, len(fts)-1)

	var s, e int32
	s = 0
	for i, ft := range fts {

		e = s + int32(ft.N)

		sp := newSpan(xs, ys, ft, s, e)
		spans[i] = sp
		s = e
	}

	for i, sp := range spans[:len(spans)-1] {
		sp2 := spans[i+1]

		merged[i] = sp.Copy()
		mergeTwoSpan(xs, ys, merged[i], sp2)
	}

	for len(merged) > 0 {

		// find minimal merge and merge

		maxReduced := -1
		maxI := 0

		for i := 1; i < len(merged); i++ {
			a := spans[i]
			b := spans[i+1]
			mr := merged[i]
			reduced := a.mem + b.mem - mr.mem
			if maxReduced < reduced {
				maxI = i
				maxReduced = reduced
			}
		}

		// maxI -> b
		//
		// span:     a  b  c  d
		// merged:    ab bc cd
		//
		// becomes:
		//
		// span:     a   bc   d
		// merged:    abc  bcd
		//
		// a => a
		// b => nil
		// c => nil
		// d => d
		// ab + c => abc
		// bc => bc
		// b + cd => bcd

		if maxReduced > 0 {

			// a  b  c  d
			// abc bc cd
			if maxI > 0 {
				mergeTwoSpan(xs, ys, merged[maxI-1], spans[maxI+1])
			}

			// a  bcd  c  d
			// abc bc bcd
			if maxI < len(merged)-1 {
				mergeTwoSpan(xs, ys, spans[maxI], merged[maxI+1])
				merged[maxI+1] = spans[maxI]
			}

			// a  bc  d
			// abc bc bcd
			spans[maxI] = merged[maxI]
			spans = append(spans[:maxI+1], spans[maxI+2:]...)

			// a  bc  d
			// abc  bcd
			merged = append(merged[:maxI], merged[maxI+1:]...)

		} else {
			// Even the minimal merge does not reduce memory cost.
			break
		}
	}

	return spans
}

func mergeTwoSpan(xs, ys []float64, a, b *span) {
	a.ft.Merge(b.ft)
	a.e = b.e

	// policy: re-fit curve
	a.solve()
	a.updatePolyAndStat(xs, ys)

	// // policy: mean curve
	// // twice faster than re-fit, also results in twice memory cost.
	// for i, c := range b.origPoly {
	//     a.origPoly[i] = (a.origPoly[i] + c) / 2
	// }
	// a.updatePolyAndStat(xs, ys)
}

func newSpan(xs, ys []float64, ft *polyfit.Fitting, s, e int32) *span {

	sp := &span{
		ft: ft,
		s:  s,
		e:  e,
	}
	sp.solve()
	sp.updatePolyAndStat(xs, ys)

	return sp
}

func (sp *span) solve() {
	sp.origPoly = sp.ft.Solve()
}

func (sp *span) updatePolyAndStat(xs, ys []float64) {
	s, e := sp.s, sp.e
	max, min := maxMinResiduals(sp.origPoly, xs[s:e], ys[s:e])
	margin := int64(math.Ceil(max - min))

	sp.poly = append([]float64{}, sp.origPoly...)
	sp.poly[0] += min

	residualWidth := marginWidth(margin)
	if residualWidth > 32 {
		residualWidth = 32
	}
	sp.residualWidth = residualWidth

	sp.mem = memCost(sp.poly, residualWidth, int32(sp.ft.N))

}

// marginWidth calculate the minimal number of bits to store `margin`.
// The returned number of bits is a power of 2: 2^k, e.g., 0, 1, 2, 4, 8...
//
// Since 0.1.1
func marginWidth(margin int64) uint32 {
	// log(2, margin)
	width := uint32(64 - bits.LeadingZeros64(uint64(margin)))

	// align width to 2^k:

	// log(2, width-1)
	lz := 32 - uint32(bits.LeadingZeros32(width-1))

	return uint32(1) << lz
}

func memCost(poly []float64, residualWidth uint32, n int32) int {
	mm := 0
	mm += 64 * (len(poly) + 1)        // Polynomials and config
	mm += int(residualWidth) * int(n) // Residuals
	return mm
}

// maxMinResiduals finds max and min residuals along a curve.
//
// Since 0.1.1
func maxMinResiduals(poly, xs, ys []float64) (float64, float64) {

	max, min := float64(0), float64(0)

	for i, x := range xs {
		v := evalPoly2(poly, x)
		diff := ys[i] - v
		if diff > max {
			max = diff
		}
		if diff < min {
			min = diff
		}
	}

	return max, min
}
