package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRBuffer(t *testing.T) {
	t.Run("Push", func(t *testing.T) {
		b := NewRbuffer(3)
		assert.Equal(t, 0, b.Index)
		assert.Equal(t, []string{"", "", ""}, b.Slice)
		b.Push("a")
		assert.Equal(t, 1, b.Index)
		assert.Equal(t, []string{"a", "", ""}, b.Slice)
		b.Push("b")
		assert.Equal(t, 2, b.Index)
		assert.Equal(t, []string{"a", "b", ""}, b.Slice)
		b.Push("c")
		assert.Equal(t, 0, b.Index)
		assert.Equal(t, []string{"a", "b", "c"}, b.Slice)
		b.Push("d")
		assert.Equal(t, 1, b.Index)
		assert.Equal(t, []string{"d", "b", "c"}, b.Slice)
	})

	t.Run("ToString", func(t *testing.T) {
		b := NewRbuffer(3)
		assert.Equal(t, "", b.ToString("+"))
		b.Push("a")
		assert.Equal(t, "a", b.ToString("+"))
		b.Push("b")
		assert.Equal(t, "a+b", b.ToString("+"))
		b.Push("c")
		assert.Equal(t, "a+b+c", b.ToString("+"))
		b.Push("d")
		assert.Equal(t, "b+c+d", b.ToString("+"))
	})
}
