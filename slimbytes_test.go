package slimarray

import (
	"fmt"
	"testing"

	"github.com/openacid/low/size"
	"github.com/openacid/testutil"
	"github.com/stretchr/testify/require"
)

func TestSlimBytes_Get(t *testing.T) {

	ta := require.New(t)

	sb, err := NewBytes([][]byte{
		[]byte("foo"), []byte("bar"), []byte(""),
		[]byte("hello"), []byte("xp"), []byte("seeyou"),
	})
	ta.NoError(err)

	cases := []struct {
		i    int32
		want string
	}{
		{0, "foo"},
		{1, "bar"},
		{2, ""},
		{3, "hello"},
		{4, "xp"},
		{5, "seeyou"},
	}

	for i, c := range cases {
		got := sb.Get(c.i)
		ta.Equal(c.want, string(got), "%d-th: case: %+v", i+1, c)
	}

	ta.Panics(func() { sb.Get(6) })
	ta.Panics(func() { sb.Get(-1) })
}

func TestSlimBytes_Get_recordLens(t *testing.T) {

	ta := require.New(t)

	cases := []int{
		5,
		10,
		20,
	}

	n := 1024 * 1024

	for _, recLen := range cases {
		t.Run(fmt.Sprintf("recLen:[%d,%d)", recLen, recLen*2),
			func(t *testing.T) {

				records := testutil.RandBytesSlice(n, recLen, recLen*2)
				sb, err := NewBytes(records)
				ta.NoError(err)
				for i, rec := range records {
					ta.Equal(rec, sb.Get(int32(i)))
				}
			})

	}
}

func TestSlimBytes_memoryOverhead(t *testing.T) {

	ta := require.New(t)

	cases := []int{
		5,
		10,
		20,
		40,
	}

	n := 1024 * 1024

	for _, recLen := range cases {
		t.Run(fmt.Sprintf("recLen:[%d,%d)", recLen, recLen*2),
			func(t *testing.T) {

				records := testutil.RandBytesSlice(n, recLen, recLen*2)
				sb, err := NewBytes(records)
				ta.NoError(err)

				payload := 0
				for _, r := range records {
					payload += len(r)
				}

				total := size.Of(sb)
				overhead := (total - payload) * 100 / payload

				// fmt.Println(size.Stat(sb, 10, 2))
				// fmt.Println(payload)
				// fmt.Println(total)
				// fmt.Println(overhead)

				ta.Less(overhead, 13, "memory overhead is 1 slimarray elt / record, thus it should be smaller than 13%")

			})
	}
}

var OutputSlimBytes byte

func BenchmarkSlimBytes_Get(b *testing.B) {

	cases := []int{
		4,
		1024,
		1024 * 1024,
	}

	for _, n := range cases {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {

			records := testutil.RandBytesSlice(n, 5, 10)

			sb, _ := NewBytes(records)

			mask := n - 1

			b.ResetTimer()

			var s byte = 0
			for i := 0; i < b.N; i++ {
				bs := sb.Get(int32(i & mask))
				s += bs[0]
			}

			OutputSlimBytes = s
		})
	}
}
