package main

import (
	"fmt"
	"time"
)

// MyNode1 implements a node that sleeps for 1 second
type MyNode1 struct {
	BaseGElement
}

// Run executes the node logic
func (n *MyNode1) Run() *CStatus {
	fmt.Printf("%s: sleep 1s.\n", n.GetName())
	time.Sleep(1 * time.Second)
	return NewCStatus()
}

// MyNode2 implements a node that sleeps for 2 seconds
type MyNode2 struct {
	BaseGElement
}

func (n *MyNode2) Run() *CStatus {
	fmt.Printf("%s: sleep 2s.\n", n.GetName())
	time.Sleep(2 * time.Second)
	return NewCStatus()
}

func tutorialSimple() {
	var a, b, c, d GElement

	pipeline := Factory.Create()

	a = &MyNode1{}
	pipeline.RegisterGElement(a, []GElement{}, "nodeA")

	b = &MyNode2{}
	pipeline.RegisterGElement(b, []GElement{a}, "nodeB")

	c = &MyNode1{}
	pipeline.RegisterGElement(c, []GElement{a}, "nodeC")

	d = &MyNode2{}
	pipeline.RegisterGElement(d, []GElement{b, c}, "nodeD")

	pipeline.Process(1)
	Factory.Remove(pipeline)
}

func main() {
	tutorialSimple()
}
