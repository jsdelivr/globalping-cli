package view

import (
	"strings"
	"time"

	"github.com/jsdelivr/globalping-go"
)

type HistoryBuffer struct {
	Index int
	Slice []*HistoryItem
}

type HistoryItem struct {
	Id          string
	Status      globalping.MeasurementStatus
	ProbeStatus []globalping.TestStatus
	StartedAt   time.Time
}

func NewHistoryBuffer(size int) *HistoryBuffer {
	return &HistoryBuffer{
		Index: 0,
		Slice: make([]*HistoryItem, size),
	}
}

func (h *HistoryBuffer) Find(id string) *HistoryItem {
	i := h.Index - 1

	for {
		if i < 0 {
			i = len(h.Slice) - 1
		}

		if h.Slice[i] != nil && h.Slice[i].Id == id {
			return h.Slice[i]
		}

		if i == h.Index {
			break
		}

		i--
	}

	return nil
}

func (h *HistoryBuffer) Push(m *HistoryItem) {
	h.Slice[h.Index] = m
	h.Index = (h.Index + 1) % len(h.Slice)
}

func (h *HistoryBuffer) Last() *HistoryItem {
	i := h.Index - 1

	if i < 0 {
		i = len(h.Slice) - 1
	}

	return h.Slice[i]
}

func (h *HistoryBuffer) Capacity() int {
	return len(h.Slice)
}

func (h *HistoryBuffer) ToString(sep string) string {
	var output strings.Builder

	i := h.Index
	isFirst := true

	for {
		if h.Slice[i] != nil {
			if isFirst {
				isFirst = false
				output.WriteString(h.Slice[i].Id)
			} else {
				output.WriteString(sep + h.Slice[i].Id)
			}
		}

		i = (i + 1) % len(h.Slice)

		if i == h.Index {
			break
		}
	}

	return output.String()
}
