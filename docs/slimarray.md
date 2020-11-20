# slimarray
--
    import "github.com/openacid/slimarray"

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


### Uncompressed Data Structures

A SlimArray is a compacted data structure. The original data structures are
defined as follow(assumes original user data is `nums []uint32`):

    Seg struct {
      SpansBitmap   uint64      // describe span layout
      Rank         uint64      // count `1` in preceding Seg.
      Spans       []Span
    }

    Span struct {
      width         int32       // is retrieved from SpansBitmap

      Polynomial [3]double      //
      Config struct {           //
        Offset        int32     // residual offset
        ResidualWidth int32     // number of bits a residual requires
      }
      Residuals  [width][ResidualWidth]bit // pack into SlimArray.Residuals
    }

A span stores 16*k int32 in it, where k ∈ [1, 64).

`Seg.SpansBitmap` describes the layout of Span-s in a Seg. The i-th "1"
indicates where the last 16 numbers are in the i-th Span. e.g.:

    001011110000......
    <-- least significant bit

In the above example:

    span[0] has 16*3 nums in it.
    span[1] has 16*2 nums in it.
    span[2] has 16*1 nums in it.

`Seg.Rank` caches the total count of "1" in all preceding Seg.SpansBitmap. This
accelerate locating a Span in the packed field SlimArray.Polynomials .

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
satisfy: `(-16*4) + i * 4` is the correct residual position, for i in [16, 32).

`Span.Config.ResidualWidth` specifies the number of bits to store every residual
in this span, it must be a power of 2: `2^k`.

`Span.Residuals` is an array of residuals of length `Span.width`. Every elt in
it is a `ResidualWidth`-bits integers.


### Compact

SlimArray compact `Seg` into a dense format:

    SlimArray.Bitmap = [
      Seg[0].SpansBitmap,
      Seg[1].SpansBitmap,
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

## Usage

```go
var File_slimarray_proto protoreflect.FileDescriptor
```

#### type SlimArray

```go
type SlimArray struct {

	// N is the count of elts
	N    int32    `protobuf:"varint,10,opt,name=N,proto3" json:"N,omitempty"`
	Rank []uint64 `protobuf:"varint,19,rep,packed,name=Rank,proto3" json:"Rank,omitempty"`
	// Every 1024 elts segment has a 64-bit bitmap to describe the spans in it,
	// and another 64-bit rank: the count of `1` in preceding bitmaps.
	Bitmap []uint64 `protobuf:"varint,20,rep,packed,name=Bitmap,proto3" json:"Bitmap,omitempty"`
	// Polynomial and config of every span.
	// 3 doubles to represent a polynomial;
	Polynomials []float64 `protobuf:"fixed64,21,rep,packed,name=Polynomials,proto3" json:"Polynomials,omitempty"`
	// Config stores the offset of residuals in Residuals and the bit width to
	// store a residual in a span.
	Configs []int64 `protobuf:"varint,22,rep,packed,name=Configs,proto3" json:"Configs,omitempty"`
	// packed residuals for every elt.
	Residuals []uint64 `protobuf:"varint,23,rep,packed,name=Residuals,proto3" json:"Residuals,omitempty"`
}
```

SlimArray compresses a uint32 array with overall trend by describing the trend
with a polynomial, e.g., to store a sorted array is very common in practice.
Such as an block-list of IP addresses, or a series of var-length record position
on disk.

E.g. a uint32 costs only 5 bits in average in a sorted array of a million number
in range [0, 1000*1000].

In addition to the unbelievable low memory footprint, a `Get` access is also
very fast: it takes only 10 nano second in our benchmark.

SlimArray is also ready for transport since it is defined with protobuf. E.g.:

    a := slimarray.NewU32([]uint32{1, 2, 3})
    bytes, err := proto.Marshal(a)

Since 0.1.1

#### func  NewU32

```go
func NewU32(nums []uint32) *SlimArray
```
NewU32 creates a "SlimArray" array from a slice of uint32.

A NewU32() costs about 110 ns/elt.

Since 0.1.1

#### func (*SlimArray) Descriptor

```go
func (*SlimArray) Descriptor() ([]byte, []int)
```
Deprecated: Use SlimArray.ProtoReflect.Descriptor instead.

#### func (*SlimArray) Get

```go
func (sm *SlimArray) Get(i int32) uint32
```
Get returns the uncompressed uint32 value. A Get() costs about 7 ns

Since 0.1.1

#### func (*SlimArray) GetBitmap

```go
func (x *SlimArray) GetBitmap() []uint64
```

#### func (*SlimArray) GetConfigs

```go
func (x *SlimArray) GetConfigs() []int64
```

#### func (*SlimArray) GetN

```go
func (x *SlimArray) GetN() int32
```

#### func (*SlimArray) GetPolynomials

```go
func (x *SlimArray) GetPolynomials() []float64
```

#### func (*SlimArray) GetRank

```go
func (x *SlimArray) GetRank() []uint64
```

#### func (*SlimArray) GetResiduals

```go
func (x *SlimArray) GetResiduals() []uint64
```

#### func (*SlimArray) Len

```go
func (sm *SlimArray) Len() int
```
Len returns number of elements.

Since 0.1.1

#### func (*SlimArray) ProtoMessage

```go
func (*SlimArray) ProtoMessage()
```

#### func (*SlimArray) ProtoReflect

```go
func (x *SlimArray) ProtoReflect() protoreflect.Message
```

#### func (*SlimArray) Reset

```go
func (x *SlimArray) Reset()
```

#### func (*SlimArray) Slice

```go
func (sm *SlimArray) Slice(start int32, end int32, rst []uint32)
```
Slice returns a slice of uncompressed uint32, e.g., similar to foo :=
nums[start:end]. `rst` is used to store returned values, it has to have at least
`end-start` elt in it.

A Slice() costs about 3.8 ns, when retrieving 100 or more values a time.

Since 0.1.3

#### func (*SlimArray) Stat

```go
func (sm *SlimArray) Stat() map[string]int32
```
Stat returns a map describing memory usage.

    seg_cnt   :512         // segment count
    elt_width :8           // average bits count per elt
    span_cnt  :12          // total count of spans
    spans/seg :7           // average span count per segment
    mem_elts  :1048576     // memory cost for residuals
    mem_total :1195245     // total memory cost
    bits/elt  :9           // average memory cost per elt
    n         :10          // total elt count

Since 0.1.1

#### func (*SlimArray) String

```go
func (x *SlimArray) String() string
```
