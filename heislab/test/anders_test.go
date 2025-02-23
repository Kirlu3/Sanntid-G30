package test

import (
	"fmt"
	"strconv"
	"testing"
)


func TestNetwork(t *testing.T) {
	network()
}

func TestFoo(t *testing.T) {
	foo()
}

func foo() {
	var a []string
	s := "one"
	num, err := strconv.Atoi(s)
	fmt.Printf("num: %v\n", num)
	fmt.Printf("err: %v\n", err)

	for i, val := range a {
		fmt.Printf("a[i]: %v\n", a[i])
		fmt.Printf("val: %v\n", val)
	}

	
}