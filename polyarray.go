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
	maxResidualWidth = 16
	minSpan          = int32(16)

	segSize      = 1024
	segSizeShift = uint(10)
	segSizeMask  = int32(1024 - 1)

	polyDegree      = 2
	polyCoefCnt     = polyDegree + 1
	f64PerSpan      = polyCoefCnt + 1 // 3 coefficients and 1 config
	f64PerSpanShift = 2
	twoPow36        = float64(int64(1) << 36)

	// In a segment we want:
	//		residual position = offset + (i%1024) * residualWidth
	// But if preceding span has smaller residual width, the "offset" could be
	// negative, e.g.: span[0] has residual of width 0 and 16 residuals,
	// span[1] has residual of width 4.
	// Then the "offset" of span[1] is -16*4 in order to satisify:
	// (-16*4) + i * 4 is the correct residual position, for i in [16, 32).
	maxNegOffset = int64(segSize) * int64(maxResidualWidth)
)

// evalpoly2 evaluates a polynomial with degree=2.
//
// Since 0.5.2
func evalpoly2(poly []float64, x float64) float64 {
	return poly[0] + poly[1]*x + poly[2]*x*x
}

// NewPolyArray creates a "PolyArray" array from a slice of int32.
// A "PolyArray" array uses polynomial curves to compress data.
//
// It is very efficient to store a serias integers with a overall trend, such as
// a sorted array.
//
// Since 0.5.2
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
	pa.Residuals = append(pa.Residuals[:0:0], pa.Residuals...)

	return pa
}

// Get returns the uncompressed int32 value.
// A Get() costs about 11 ns
//
// Since 0.5.2
func (m *PolyArray) Get(i int32) int32 {

	bitmapI := (i >> segSizeShift) << 1
	polyBitmap, rank := m.Bitmap[bitmapI], m.Bitmap[bitmapI|1]

	i = i & segSizeMask
	x := float64(i)

	bm := polyBitmap & bitmap.Mask[i>>4]
	polyI := int(rank) + bits.OnesCount64(bm)

	// evalpoly2(poly, x)

	j := polyI << f64PerSpanShift
	p := m.Polynomials
	v := int32(p[j] + p[j+1]*x + p[j+2]*x*x)

	// read the config of this polynomial:
	// residualWidth: how many bits a residual needs.
	// offset.

	config := p[j+3]
	residualWidth := int64(config)
	offset := int64((config-float64(residualWidth))*twoPow36) - maxNegOffset

	// where the residual is
	ibit := offset + int64(i)*residualWidth

	d := m.Residuals[ibit>>6]
	d = d >> uint(ibit&63)

	return v + int32(d&bitmap.Mask[residualWidth])
}

// Len returns number of elements.
//
// Since 0.5.2
func (m *PolyArray) Len() int {
	return int(m.N)
}

// Stat returns a map describing memory usage.
//
//    elt_width :8
//    seg_cnt   :512
//    polys/seg :7
//    mem_elts  :1048576
//    mem_total :1195245
//    bits/elt  :9
//
// Since 0.5.2
func (m *PolyArray) Stat() map[string]int32 {
	nseg := len(m.Bitmap) / 2
	totalmem := size.Of(m)

	polyCnt := len(m.Polynomials) >> 2
	memWords := len(m.Residuals) * 8
	widthAvg := 0
	for i := 0; i < polyCnt; i++ {
		// get the last float64
		_, w := unpackConfig(m.Polynomials[i*f64PerSpan+f64PerSpan-1])
		widthAvg += int(w)
	}

	n := m.Len()
	if n == 0 {
		n = 1
	}

	if polyCnt == 0 {
		polyCnt = 1
	}

	st := map[string]int32{
		"seg_cnt":   int32(nseg),
		"elt_width": int32(widthAvg / polyCnt),
		"mem_total": int32(totalmem),
		"mem_elts":  int32(memWords),
		"bits/elt":  int32(totalmem * 8 / n),
		"polys/seg": int32((polyCnt * 1000) / (nseg*1000 + 1)),
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
	s, e          int32
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

		sp := estimatePoly(xs, ys, ft, s, e)
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
	sp := estimatePoly(xs, ys, ft, a.s, b.e)
	return sp
}

func estimatePoly(xs, ys []float64, ft *polyfit.Fitting, s, e int32) span {

	poly := ft.Solve()
	max, min := maxminResiduals(poly, xs[s:e], ys[s:e])
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
	for _, width := range []uint32{0, 1, 2, 4, 8, 16} {
		if int32(1)<<width > margin {
			return width
		}
	}

	panic(fmt.Sprintf("margin is too large: %d >= 2^16", margin))
}

func memCost(poly []float64, residualWidth uint32, n int32) int {
	mm := 0
	mm += 64 * (len(poly) + 1)        // Polynomials and config
	mm += int(residualWidth) * int(n) // Residuals
	return mm
}

// maxminResiduals finds max and min residuals along a curve.
//
// Since 0.5.2
func maxminResiduals(poly, xs, ys []float64) (float64, float64) {

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
