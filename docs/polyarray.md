# polyarray
--
    import "github.com/openacid/polyarray"

package polyarray uses polynomial to compress and store an array of uint32. A
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

PolyArray does not have such restriction. It is more adaptive with data layout.
To achieve high compression rate, it only requires the data has a overall trend,
e.g., roughly sorted, as seen in the above 4 integers examples. Additionally, it
also accept duplicated element in the array, which a bitmap based or tree-like
data structure does not allow.


### Data Structure

PolyArray splits the entire array into segments(Seg), each of which has 1024
numbers. And then it splits every segment into several spans. Every span has its
own polynomial. A span has 16*k numbers. A segment has at most 64 spans.

            seg[0]                      seg[1]
            1024 nums                   1024 nums
    |-------+---------------+---|---------------------------|...
     span[0]    span[1]
     16 nums    32 nums      ..


### Uncompacted Data Structures

A PolyArray is a compacted data structure. The original data structures are
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
      Residuals  [width][ResidualWidth]bit // pack into PolyArray.Residuals
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
This accelerate locating a Span in the packed field PolyArray.Polynomials .

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

PolyArray compact `Seg` into a dense format:

    PolyArray.Bitmap = [
      Seg[0].SpansBitmap,
      Seg[0].OnesCount,
      Seg[1].SpansBitmap,
      Seg[1].OnesCount,
      ... ]

    PolyArray.Polynomials = [
      Seg[0].Spans[0].Polynomials,
      Seg[0].Spans[1].Polynomials,
      ...
      Seg[1].Spans[0].Polynomials,
      Seg[1].Spans[1].Polynomials,
      ...
    ]

    PolyArray.Configs = [
      Seg[0].Spans[0].Config
      Seg[0].Spans[1].Config
      ...
      Seg[1].Spans[0].Config
      Seg[1].Spans[1].Config
      ...
    ]

`PolyArray.Residuals` simply packs the residuals of every nums[i] together.

## Usage

#### type PolyArray

```go
type PolyArray struct {
	// N is the count of elts
	N int32 `protobuf:"varint,10,opt,name=N,proto3" json:"N,omitempty"`
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
	Residuals            []uint64 `protobuf:"varint,23,rep,packed,name=Residuals,proto3" json:"Residuals,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}
```

PolyArray compresses a uint32 array with overall trend by describing the trend
with a polynomial, e.g., to store a sorted array is very common in practice.
Such as an block-list of IP addresses, or a series of var-length record position
on disk.

E.g. a uint32 costs only 5 bits in average in a sorted array of a million number
in range [0, 1000*1000].

In addition to the unbelievable low memory footprint, a `Get` access is also
very fast: it takes only 10 nano second in our benchmark.

PolyArray is also ready for transport since it is defined with protobuf. E.g.:

    a := polyarray.NewPolyArray([]uint32{1, 2, 3})
    bytes, err := proto.Marshal(a)

Since 0.1.1

#### func  NewPolyArray

```go
func NewPolyArray(nums []uint32) *PolyArray
```
NewPolyArray creates a "PolyArray" array from a slice of uint32.

Since 0.1.1

#### func (*PolyArray) Descriptor

```go
func (*PolyArray) Descriptor() ([]byte, []int)
```

#### func (*PolyArray) Get

```go
func (m *PolyArray) Get(i int32) uint32
```
Get returns the uncompressed uint32 value. A Get() costs about 10 ns

Since 0.1.1

#### func (*PolyArray) GetBitmap

```go
func (m *PolyArray) GetBitmap() []uint64
```

#### func (*PolyArray) GetConfigs

```go
func (m *PolyArray) GetConfigs() []int64
```

#### func (*PolyArray) GetN

```go
func (m *PolyArray) GetN() int32
```

#### func (*PolyArray) GetPolynomials

```go
func (m *PolyArray) GetPolynomials() []float64
```

#### func (*PolyArray) GetResiduals

```go
func (m *PolyArray) GetResiduals() []uint64
```

#### func (*PolyArray) Len

```go
func (m *PolyArray) Len() int
```
Len returns number of elements.

Since 0.1.1

#### func (*PolyArray) ProtoMessage

```go
func (*PolyArray) ProtoMessage()
```

#### func (*PolyArray) Reset

```go
func (m *PolyArray) Reset()
```

#### func (*PolyArray) Stat

```go
func (m *PolyArray) Stat() map[string]int32
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

#### func (*PolyArray) String

```go
func (m *PolyArray) String() string
```

#### func (*PolyArray) XXX_DiscardUnknown

```go
func (m *PolyArray) XXX_DiscardUnknown()
```

#### func (*PolyArray) XXX_Marshal

```go
func (m *PolyArray) XXX_Marshal(b []byte, deterministic bool) ([]byte, error)
```

#### func (*PolyArray) XXX_Merge

```go
func (dst *PolyArray) XXX_Merge(src proto.Message)
```

#### func (*PolyArray) XXX_Size

```go
func (m *PolyArray) XXX_Size() int
```

#### func (*PolyArray) XXX_Unmarshal

```go
func (m *PolyArray) XXX_Unmarshal(b []byte) error
```
