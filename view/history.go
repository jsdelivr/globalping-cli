package view

import "time"

type HistoryBuffer struct {
	Index int
	Slice []*HistoryItem
}

type HistoryItem struct {
	Id           string
	LinesPrinted int
	StartedAt    time.Time
}

func NewHistoryBuffer(size int) *HistoryBuffer {
	return &HistoryBuffer{
		Index: 0,
		Slice: make([]*HistoryItem, size),
	}
}

func (q *HistoryBuffer) Find(id string) *HistoryItem {
	i := q.Index - 1
	for {
		if i < 0 {
			i = len(q.Slice) - 1
		}
		if q.Slice[i] != nil && q.Slice[i].Id == id {
			return q.Slice[i]
		}
		if i == q.Index {
			break
		}
		i--
	}
	return nil
}

func (q *HistoryBuffer) Push(m *HistoryItem) {
	q.Slice[q.Index] = m
	q.Index = (q.Index + 1) % len(q.Slice)
}

func (q *HistoryBuffer) Capacity() int {
	return len(q.Slice)
}

func (q *HistoryBuffer) ToString(sep string) string {
	s := ""
	i := q.Index
	isFirst := true
	for {
		if q.Slice[i] != nil {
			if isFirst {
				isFirst = false
				s += q.Slice[i].Id
			} else {
				s += sep + q.Slice[i].Id
			}
		}
		i = (i + 1) % len(q.Slice)
		if i == q.Index {
			break
		}
	}
	return s
}
