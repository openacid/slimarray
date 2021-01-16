package slimarray

import (
	"fmt"
	"testing"

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

var OutputSlimBytes byte

func BenchmarkSlimBytes_Get(b *testing.B) {

	cases := []int{
		1024,
		1024 * 1024,
	}

	for _, n := range cases {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {

			records := testutil.RandBytesSlice(n, 5, 10)

			sb, _ := NewBytes(records)

			b.ResetTimer()

			var s byte = 0
			for i := 0; i < b.N; i++ {
				bs := sb.Get(int32(i % 5))
				s += bs[0]
			}

			OutputSlimBytes = s
		})
	}
}
