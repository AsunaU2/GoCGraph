package main

import (
	"fmt"
	"time"

	gocgraph "github.com/AsunaU2/GoCGraph"
)

// MyNode1 implements a node that sleeps for 1 second
type MyNode1 struct {
	gocgraph.BaseGElement
}

// Run executes the node logic
func (n *MyNode1) Run() *gocgraph.CStatus {
	fmt.Printf("%s: sleep 1s.\n", n.GetName())
	time.Sleep(1 * time.Second)
	return gocgraph.NewCStatus()
}

// MyNode2 implements a node that sleeps for 2 seconds
type MyNode2 struct {
	gocgraph.BaseGElement
}

func (n *MyNode2) Run() *gocgraph.CStatus {
	fmt.Printf("%s: sleep 2s.\n", n.GetName())
	time.Sleep(2 * time.Second)
	return gocgraph.NewCStatus()
}

func tutorialSimple() {
	var a, b, c, d gocgraph.GElement

	pipeline := gocgraph.Factory.Create()

	a = &MyNode1{}
	pipeline.RegisterGElement(a, []gocgraph.GElement{}, "nodeA")

	b = &MyNode2{}
	pipeline.RegisterGElement(b, []gocgraph.GElement{a}, "nodeB")

	c = &MyNode1{}
	pipeline.RegisterGElement(c, []gocgraph.GElement{a}, "nodeC")

	d = &MyNode2{}
	pipeline.RegisterGElement(d, []gocgraph.GElement{b, c}, "nodeD")

	pipeline.Process(1)
	gocgraph.Factory.Remove(pipeline)
}

func main() {
	tutorialSimple()
}
