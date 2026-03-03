package array_vs_list

import (
	"fmt"
)

type ArrayDS2 struct {
	data []int
}

func NewArrayDS2(capacity int) *ArrayDS2 {
	return &ArrayDS2{
		data: make([]int, 0, capacity),
	}
}

func (a *ArrayDS2) Append(val int) {
	a.data = append(a.data, val)

}

func (a *ArrayDS2) Insert(index, val int) error {
	if index < 0 || index >= len(a.data) {
		return fmt.Errorf("index out of range: 0-%d", index)
	}
	a.data = append(a.data[:index], a.data[index+1:]...)
	return nil
}
func (a *ArrayDS2) Delete(index int) error {
	if index < 0 || index >= len(a.data) {
		return fmt.Errorf("index out of range: 0-%d", index)
	}
	a.data = append(a.data[:index], a.data[index+1:]...)
	return nil
}

func (a *ArrayDS2) Get(index int) (int, error) {
	if index < 0 || index >= len(a.data) {
		return 0, fmt.Errorf("index out of range: 0-%d", index)
	}

	return a.data[index], nil
}

func (a *ArrayDS2) Len() int {
	return len(a.data)
}
