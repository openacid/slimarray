// package polyarray uses one or more polynomial to compress and store an array of int32.
//
// The general idea
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
// Data structure
//
// PolyArray splits the entire array into segments(Seg),
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
// Uncompacted data structures
//
// A PolyArray is a compacted data structure.
// The original data structures are defined as follow(assumes original user data
// is `nums []int32`):
//
//   Seg strcut {
//     SpansBitmap   uint64      // describe span layout
//     OnesCount     uint64      // count `1` in preceding Seg.
//     Spans       []Span
//   }
//
//   Span struct {
//     width         int32       // is retrieved from SpansBitmap
//
//     Polynomial [3]double      //
//     Config strcut {           //
//       Offset        int32     // residual offset
//       ResidualWidth int32     // number of bits a residual requires
//     }
//     Residuals  [width][ResidualWidth]bit // pack into PolyArray.Residuals
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
// `Seg.OnesCount` caches the total count of "1" in all preceding Seg.SpansBitmap.
// This accelerate locating a Span in the packed field PolyArray.Polynomials .
//
// `Span.width` is the count of numbers stored in this span.
// It does not need to be stored because it can be calculated by counting the
// "0" between two "1" in `Seg.SpansBitmap`.
//
// `Span.Polynomial` stores 3 coefficients of the polynomial describing the
// overall trend of this span. I.e. the `[a, b, c]` in `y = a + bx + cx²`
//
// `Span.Config.Offset` adjust the offset to locate a residual.
//
// `Span.Config.ResidualWidth` specifies the number of bits to
// store every residual in this span, it must be a power of 2: `2^k`.
//
// With Offset = 123, ResidualWidth = 4, then the packed config is a double
// value:
// 2^14 + 4 + 123 / 2^36 = 16388.0000000017899
//
// A double value has 52 bit for Significand field, thus a double has enough
// capacity to store a Config.
//
// `Span.Residuals` is an array of residuals of length `Span.width`.
// Every elt in it is a `ResidualWidth`-bits integers.
//
// Compact
//
// PolyArray compact `Seg` into a dense format:
//
//   PolyArray.Bitmap = [
//     Seg[0].SpansBitmap,
//     Seg[0].OnesCount,
//     Seg[1].SpansBitmap,
//     Seg[1].OnesCount,
//     ... ]
//
//   PolyArray.Polynomials = [
//     Seg[0].Spans[0].Polynomials,
//     Seg[0].Spans[0].Config,
//     Seg[0].Spans[1].Polynomials,
//     Seg[0].Spans[1].Config,
//     ...
//     Seg[1].Spans[0].Polynomials,
//     Seg[1].Spans[0].Config,
//     Seg[1].Spans[1].Polynomials,
//     Seg[1].Spans[1].Config,
//     ...
//   ]
//
// `PolyArray.Residuals` simply packs the residuals of every nums[i] together.
//
// Since 0.1.1
package polyarray

import (
	"fmt"
	"math"
	"math/bits"

	"github.com/openacid/low/bitmap"
	"github.com/openacid/low/size"
	"github.com/openacid/polyarray/polyfit"
)

const (
	// The max number of bits to store a residual.
	maxResidualWidth = 16

	// The smallest span is 16 int32 numbers.
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

	// A span has 4 float64: 3 coefficients and 1 as placeholder of span config,
	// see packConfig()
	f64PerSpan      = polyCoefCnt + 1 // 3 coefficients and 1 config
	f64PerSpanShift = 2

	twoPow36 = float64(int64(1) << 36)

	// In a span we want:
	//		residual position = offset + (i%1024) * residualWidth
	// But if preceding span has smaller residual width, the "offset" could be
	// negative, e.g.: span[0] has residual of width 0 and 16 residuals,
	// span[1] has residual of width 4.
	// Then the "offset" of span[1] is -16*4 in order to satisify:
	// (-16*4) + i * 4 is the correct residual position, for i in [16, 32).
	maxNegOffset = int64(segSize) * int64(maxResidualWidth)
	// TODO: negative float seems allright?
)

// evalpoly2 evaluates a polynomial with degree=2.
//
// Since 0.1.1
func evalpoly2(poly []float64, x float64) float64 {
	return poly[0] + poly[1]*x + poly[2]*x*x
}

// NewPolyArray creates a "PolyArray" array from a slice of int32.
//
// Since 0.1.1
func NewPolyArray(nums []int32) *PolyArray {

	pa := &PolyArray{
		N: int32(len(nums)),
	}

	for ; len(nums) > segSize; nums = nums[segSize:] {
		pa.addSeg(nums[:segSize])
	}
	if len(nums) > 0 {
		pa.addSeg(nums)
	}

	// Add another empty word to avoid panic for residual of width = 0.
	pa.Residuals = append(pa.Residuals, 0)
	// shrink capacity to len.
	pa.Residuals = append(pa.Residuals[:0:0], pa.Residuals...)

	return pa
}

// Get returns the uncompressed int32 value.
// A Get() costs about 11 ns
//
// Since 0.1.1
func (m *PolyArray) Get(i int32) int32 {

	// The index of a segment
	bitmapI := (i >> segSizeShift) << 1
	spansBitmap, rank := m.Bitmap[bitmapI], m.Bitmap[bitmapI|1]

	i = i & segSizeMask
	x := float64(i)

	// i>>4 is in-segment span index
	bm := spansBitmap & bitmap.Mask[i>>4]
	spanIdx := int(rank) + bits.OnesCount64(bm)

	// eval y = a + bx + cx²

	j := spanIdx << f64PerSpanShift
	p := m.Polynomials
	v := int32(p[j] + p[j+1]*x + p[j+2]*x*x)

	// read the config of this Span:
	// residualWidth: how many bits a residual needs.
	// offset.

	config := p[j+3]
	residualWidth := int64(config)
	offset := int64((config-float64(residualWidth))*twoPow36) - maxNegOffset

	// where the residual is
	resBitIdx := offset + int64(i)*residualWidth

	// extract residual from packed []uint64
	d := m.Residuals[resBitIdx>>6]
	d = d >> uint(resBitIdx&63)

	return v + int32(d&bitmap.Mask[residualWidth])
}

// Len returns number of elements.
//
// Since 0.1.1
func (m *PolyArray) Len() int {
	return int(m.N)
}

// Stat returns a map describing memory usage.
//
//    elt_width :8
//    seg_cnt   :512
//    spans/seg :7
//    mem_elts  :1048576
//    mem_total :1195245
//    bits/elt  :9
//
// Since 0.1.1
func (m *PolyArray) Stat() map[string]int32 {
	nseg := len(m.Bitmap) / 2
	totalmem := size.Of(m)

	spanCnt := len(m.Polynomials) >> 2
	memWords := len(m.Residuals) * 8
	widthAvg := 0
	for i := 0; i < spanCnt; i++ {
		// get the last float64
		_, w := unpackConfig(m.Polynomials[i*f64PerSpan+f64PerSpan-1])
		widthAvg += int(w)
	}

	n := m.Len()
	if n == 0 {
		n = 1
	}

	if spanCnt == 0 {
		spanCnt = 1
	}

	st := map[string]int32{
		"seg_cnt":   int32(nseg),
		"elt_width": int32(widthAvg / spanCnt),
		"mem_total": int32(totalmem),
		"mem_elts":  int32(memWords),
		"bits/elt":  int32(totalmem * 8 / n),
		"spans/seg": int32((spanCnt * 1000) / (nseg*1000 + 1)),
	}

	return st
}

func packConfig(offset int64, residualWidth int64) float64 {
	return float64(residualWidth) + float64(offset)/twoPow36
}

func unpackConfig(config float64) (int64, int64) {
	residualWidth := int64(config)
	offset := int64((config - float64(residualWidth)) * twoPow36)
	return offset, residualWidth
}

func (m *PolyArray) addSeg(nums []int32) {

	bm, polys, words := newSeg(nums, int64(len(m.Residuals)*64))

	var r uint64
	if len(m.Bitmap) > 0 {
		l := len(m.Bitmap)
		r = m.Bitmap[l-1] + uint64(bits.OnesCount64(m.Bitmap[l-2]))
	} else {
		r = 0
	}

	m.Bitmap = append(m.Bitmap, bm, r)
	m.Polynomials = append(m.Polynomials, polys...)
	m.Residuals = append(m.Residuals, words...)
}

func newSeg(nums []int32, start int64) (uint64, []float64, []uint64) {

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

	polys := make([]float64, 0)
	words := make([]uint64, n) // max size

	// Using a bitmap to describe which spans a polynomial spans
	segPolyBitmap := uint64(0)

	resI := int64(0)

	for _, sp := range spans {

		// every poly starts at 16*k th point
		segPolyBitmap |= (1 << uint((sp.e-1)>>4))

		width := sp.residualWidth
		margin := int32((1 << width) - 1)
		if width > 0 {
			resI = (resI + int64(width) - 1)
			resI -= resI % int64(width)
		}

		if resI+start >= int64(1)<<35 {
			panic(fmt.Sprintf("wordStart is too large:%d, should < 2^35", resI+start))
		}

		polys = append(polys, sp.poly...)

		// We want eltIndex = stBySeg + i * residualWidth
		// min of stBySeg is -segmentSize * residualWidth = -1024 * 16;
		// Add this value to make it a positive number.
		offset := resI + start - int64(sp.s)*int64(width)
		config := packConfig(offset+maxNegOffset, int64(width))
		polys = append(polys, config)

		for j := sp.s; j < sp.e; j++ {

			v := evalpoly2(sp.poly, xs[j])

			d := nums[j] - int32(v)
			if d > margin || d < 0 {
				panic(fmt.Sprintf("d=%d must smaller than %d and > 0", d, margin))
			}

			wordI := resI >> 6
			words[wordI] |= uint64(d) << uint(resI&63)

			resI += int64(width)
		}
	}

	nWords := (resI + 63) >> 6

	return segPolyBitmap, polys, words[:nWords]
}

func initFittings(xs, ys []float64, polysize int32) []*polyfit.Fitting {

	fts := make([]*polyfit.Fitting, 0)
	n := int32(len(xs))

	for i := int32(0); i < n; i += polysize {
		s := i
		e := s + polysize
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
	poly          []float64
	residualWidth uint32
	mem           int

	// start and end index in original []int32
	s, e int32
}

func (sp span) String() string {
	return fmt.Sprintf("%d-%d(%d): width: %d, mem: %d, poly: %v",
		sp.s, sp.e, sp.e-sp.s, sp.residualWidth, sp.mem, sp.poly)
}

// findMinFittingsNew by merge adjacent 16-numbers span.
// If two spans has a common trend they should be described with one polynomial.
func findMinFittingsNew(xs, ys []float64, fts []*polyfit.Fitting) []span {

	if len(fts) == 0 {
		return []span{}
	}

	spans := make([]span, len(fts))
	merged := make([]span, len(fts)-1)

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
		merged[i] = mergeTwoSpan(xs, ys, sp, sp2)
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
			if maxReduced < a.mem+b.mem-mr.mem {
				maxI = i
				maxReduced = reduced
			}
		}

		if maxReduced > 0 {

			spans[maxI] = mergeTwoSpan(xs, ys, spans[maxI], spans[maxI+1])
			spans = append(spans[:maxI+1], spans[maxI+2:]...)
			if maxI > 0 {
				merged[maxI-1] = mergeTwoSpan(xs, ys, spans[maxI-1], spans[maxI])
			}

			merged = append(merged[:maxI], merged[maxI+1:]...)

			if maxI < len(spans)-1 {
				merged[maxI] = mergeTwoSpan(xs, ys, spans[maxI], spans[maxI+1])
			}
		} else {
			// Even the minimal merge does not reduce memory cost.
			break
		}
	}

	return spans
}

func mergeTwoSpan(xs, ys []float64, a, b span) span {
	ft := mergeTwoFitting(a.ft, b.ft)
	sp := newSpan(xs, ys, ft, a.s, b.e)
	return sp
}

func newSpan(xs, ys []float64, ft *polyfit.Fitting, s, e int32) span {

	poly := ft.Solve()
	max, min := maxMinResiduals(poly, xs[s:e], ys[s:e])
	margin := int32(math.Ceil(max - min))
	poly[0] += min

	residualWidth := marginWidth(margin)
	mem := memCost(poly, residualWidth, int32(ft.N))

	return span{
		ft:            ft,
		poly:          poly,
		residualWidth: residualWidth,
		mem:           mem,
		s:             s,
		e:             e,
	}
}

func mergeTwoFitting(a, b *polyfit.Fitting) *polyfit.Fitting {

	f := polyfit.NewFitting(nil, nil, a.Degree)
	f.Merge(a)
	f.Merge(b)

	return f
}

func marginWidth(margin int32) uint32 {
	if margin >= 65536 {
		panic(fmt.Sprintf("margin is too large: %d >= 2^16", margin))
	}

	// log(2, margin)
	width := uint32(32 - bits.LeadingZeros32(uint32(margin)))

	// align width to 2^k:

	// log(2, width-1)
	lz := uint32(32 - uint32(bits.LeadingZeros32(width-1)))

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
		v := evalpoly2(poly, x)
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
