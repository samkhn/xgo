package pq

import (
	"container/heap"
	"testing"
)

func TestPriorityQueue(t *testing.T) {
	items := map[string]int{
		"apple":  1,
		"banana": 2,
		"orange": 3,
	}
	q := make(PriorityQueue, len(items))
	i := 0
	for value, priority := range items {
		q[i] = &Item{
			value:    value,
			priority: priority,
			index:    i,
		}
		i++
	}
	heap.Init(&q)

	item := &Item{
		value:    "orange",
		priority: 1,
	}
	heap.Push(&q, item)
	q.update(item, item.value, 5)

	i = 0
	want := []int{5, 3, 2, 1}
	// Items come out in descending order
	for q.Len() > 0 {
		item := heap.Pop(&q).(*Item)
		if item.priority != want[i] {
			t.Errorf("PQ[%d] fail. Got %d. Want %d\n", i, item.priority, want[i])
		}
		i++
	}
}
