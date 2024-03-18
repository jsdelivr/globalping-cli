package view

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_HistoryBuffer(t *testing.T) {
	b := NewHistoryBuffer(3)

	assert.Equal(t, 0, b.Index)
	assert.Equal(t, []*HistoryItem{nil, nil, nil}, b.Slice)
	assert.Equal(t, b.ToString("+"), "")

	b.Push(&HistoryItem{Id: "a"})
	assert.Equal(t, 1, b.Index)
	assert.Equal(t, []*HistoryItem{{Id: "a"}, nil, nil}, b.Slice)
	assert.Equal(t, &HistoryItem{Id: "a"}, b.Find("a"))

	b.Push(&HistoryItem{Id: "b"})
	assert.Equal(t, 2, b.Index)
	assert.Equal(t, []*HistoryItem{
		{Id: "a"},
		{Id: "b"},
		nil,
	}, b.Slice)
	assert.Equal(t, b.ToString("+"), "a+b")
	assert.Equal(t, &HistoryItem{Id: "b"}, b.Find("b"))
	assert.Equal(t, &HistoryItem{Id: "a"}, b.Find("a"))

	b.Push(&HistoryItem{Id: "c"})
	assert.Equal(t, 0, b.Index)
	assert.Equal(t, []*HistoryItem{
		{Id: "a"},
		{Id: "b"},
		{Id: "c"},
	}, b.Slice)
	assert.Equal(t, b.ToString("+"), "a+b+c")

	b.Push(&HistoryItem{Id: "d"})
	assert.Equal(t, 1, b.Index)
	assert.Equal(t, []*HistoryItem{
		{Id: "d"},
		{Id: "b"},
		{Id: "c"},
	}, b.Slice)
	assert.Equal(t, b.ToString("+"), "b+c+d")
	assert.Nil(t, b.Find("a"))
}
