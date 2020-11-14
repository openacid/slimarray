package slimarray uses polynomial to compress and store an array of uint32. A
uint32 costs only 5 bits in a sorted array of a million number in range [0,
1000*1000].


### The General Idea

We use a polynomial y = a + bx + cx² to describe the overall trend of the
numbers. And for every number i we add a residual to fit the gap between y(i)
and nums[i]. E.g. If there are 4 numbers: 0, 15, 33, 50 The polynomial and
residuals are:

    y = 16x
    0, -1, 1, 2

In this case the residuals require 3 bits for each of them. To retrieve the
numbers, we evaluate y(i) and add the residual to it:

    get(0) = y(0) + 0 = 16 * 0 + 0 = 0
    get(1) = y(1) - 1 = 16 * 1 - 1 = 15
    get(2) = y(2) + 1 = 16 * 2 + 1 = 33
    get(3) = y(3) + 2 = 16 * 3 + 2 = 50


### What It Is And What It Is Not

Another space efficient data structure to store uint32 array is trie or prefix
tree or radix tree. It is possible to use bitmap-based btree like structure to
reduce space(very likely in such case it provides higher compression rate). But
it requires the array to be sorted.

SlimArray does not have such restriction. It is more adaptive with data layout.
To achieve high compression rate, it only requires the data has a overall trend,
e.g., roughly sorted, as seen in the above 4 integers examples. Additionally, it
also accept duplicated element in the array, which a bitmap based or tree-like
data structure does not allow.


### Data Structure

SlimArray splits the entire array into segments(Seg), each of which has 1024
numbers. And then it splits every segment into several spans. Every span has its
own polynomial. A span has 16*k numbers. A segment has at most 64 spans.

            seg[0]                      seg[1]
            1024 nums                   1024 nums
    |-------+---------------+---|---------------------------|...
     span[0]    span[1]
     16 nums    32 nums      ..


### Uncompacted Data Structures

A SlimArray is a compacted data structure. The original data structures are
defined as follow(assumes original user data is `nums []uint32`):

    Seg strcut {
      SpansBitmap   uint64      // describe span layout
      OnesCount     uint64      // count `1` in preceding Seg.
      Spans       []Span
    }

    Span struct {
      width         int32       // is retrieved from SpansBitmap

      Polynomial [3]double      //
      Config strcut {           //
        Offset        int32     // residual offset
        ResidualWidth int32     // number of bits a residual requires
      }
      Residuals  [width][ResidualWidth]bit // pack into SlimArray.Residuals
    }

A span stores 16*k int32 in it, where k ∈ [1, 64).

`Seg.SpansBitmap` describes the layout of Span-s in a Seg. A "1" at i-th bit and
a "1" at j-th bit means a Span stores `nums[i*16:j*16]`, e.g.:

    100101110000......
    <-- least significant bit

In the above example:

    span[0] has 16*3 nums in it.
    span[1] has 16*2 nums in it.
    span[2] has 16*1 nums in it.

`Seg.OnesCount` caches the total count of "1" in all preceding Seg.SpansBitmap.
This accelerate locating a Span in the packed field SlimArray.Polynomials .

`Span.width` is the count of numbers stored in this span. It does not need to be
stored because it can be calculated by counting the "0" between two "1" in
`Seg.SpansBitmap`.

`Span.Polynomial` stores 3 coefficients of the polynomial describing the overall
trend of this span. I.e. the `[a, b, c]` in `y = a + bx + cx²`

`Span.Config.Offset` adjust the offset to locate a residual. In a span we want
to have that:

    residual position = Config.Offset + (i%1024) * Config.ResidualWidth

But if the preceding span has smaller residual width, the "offset" could be
negative, e.g.: span[0] has residual of width 0 and 16 residuals, span[1] has
residual of width 4. Then the "offset" of span[1] is `-16*4` in order to
satisify: `(-16*4) + i * 4` is the correct residual position, for i in [16, 32).

`Span.Config.ResidualWidth` specifies the number of bits to store every residual
in this span, it must be a power of 2: `2^k`.

`Span.Residuals` is an array of residuals of length `Span.width`. Every elt in
it is a `ResidualWidth`-bits integers.


### Compact

SlimArray compact `Seg` into a dense format:

    SlimArray.Bitmap = [
      Seg[0].SpansBitmap,
      Seg[0].OnesCount,
      Seg[1].SpansBitmap,
      Seg[1].OnesCount,
      ... ]

    SlimArray.Polynomials = [
      Seg[0].Spans[0].Polynomials,
      Seg[0].Spans[1].Polynomials,
      ...
      Seg[1].Spans[0].Polynomials,
      Seg[1].Spans[1].Polynomials,
      ...
    ]

    SlimArray.Configs = [
      Seg[0].Spans[0].Config
      Seg[0].Spans[1].Config
      ...
      Seg[1].Spans[0].Config
      Seg[1].Spans[1].Config
      ...
    ]

`SlimArray.Residuals` simply packs the residuals of every nums[i] together.

