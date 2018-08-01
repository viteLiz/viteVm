package vm

import (
	"fmt"
	"math/big"
)

type Stack struct {
	data []*big.Int
}

func newStack() *Stack {
	return &Stack{data: make([]*big.Int, 0, StackLimit)}
}

func (st *Stack) push(d *big.Int) {
	st.data = append(st.data, d)
}

func (st *Stack) pop() (ret *big.Int) {
	ret = st.data[len(st.data)-1]
	st.data = st.data[:len(st.data)-1]
	return
}

func (st *Stack) peek() *big.Int {
	return st.data[st.len()-1]
}

func (st *Stack) len() int {
	return len(st.data)
}

func (st *Stack) require(n int) error {
	if st.len() < n {
		return fmt.Errorf("stack underflow (%d <=> %d)", len(st.data), n)
	}
	return nil
}

// Back returns the n'th item in stack
func (st *Stack) Back(n int) *big.Int {
	return st.data[st.len()-n-1]
}

func (st *Stack) dup(pool *intPool, n int) {
	st.push(pool.get().Set(st.data[st.len()-n]))
}

func (st *Stack) swap(n int) {
	st.data[st.len()-n], st.data[st.len()-1] = st.data[st.len()-1], st.data[st.len()-n]
}

func (st *Stack) ToString() string {
	var result string
	if len(st.data) > 0 {
		for i, val := range st.data {
			if i == len(st.data)-1 {
				result += val.Text(16)
			} else {
				result += val.Text(16) + ", "
			}
		}
	}
	return result
}
