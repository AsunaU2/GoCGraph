package main

import (
	"fmt"

	gocgraph "github.com/AsunaU2/GoCGraph"
)

const ParamKey = "param_key"

// MyParam implements the GParam interface for parameter sharing
type MyParam struct {
	gocgraph.BaseGParam
	Val  int `json:"val"`
	Loop int `json:"loop"`
}

// Reset resets the parameter values
func (p *MyParam) Reset(curStatus *gocgraph.CStatus) {
	p.Val = 0
}

// MyReadParamNode reads shared parameters
type MyReadParamNode struct {
	gocgraph.BaseGElement
}

// Init initializes the node and creates shared parameters
func (n *MyReadParamNode) Init() *gocgraph.CStatus {
	return n.CreateGParam(ParamKey, &MyParam{})
}

// Run executes the read logic
func (n *MyReadParamNode) Run() *gocgraph.CStatus {
	param := n.GetGParam(ParamKey)
	if param == nil {
		return gocgraph.NewCStatusWithError("param not found")
	}

	myParam, ok := param.(*MyParam)
	if !ok {
		return gocgraph.NewCStatusWithError("param type mismatch")
	}

	// Lock for thread safety
	myParam.GetLock().RLock()
	fmt.Printf("%s: val = %d, loop = %d\n", n.GetName(), myParam.Val, myParam.Loop)
	myParam.GetLock().RUnlock()

	return gocgraph.NewCStatus()
}

// MyWriteParamNode writes to shared parameters
type MyWriteParamNode struct {
	gocgraph.BaseGElement
}

// Run executes the write logic
func (n *MyWriteParamNode) Run() *gocgraph.CStatus {
	param := n.GetGParam(ParamKey)
	if param == nil {
		return gocgraph.NewCStatusWithError("param not found")
	}

	myParam, ok := param.(*MyParam)
	if !ok {
		return gocgraph.NewCStatusWithError("param type mismatch")
	}

	// Lock for thread safety
	myParam.GetLock().Lock()
	myParam.Val++
	myParam.Loop++
	fmt.Printf("%s: val = %d, loop = %d\n", n.GetName(), myParam.Val, myParam.Loop)
	myParam.GetLock().Unlock()

	return gocgraph.NewCStatus()
}

// tutorialParam demonstrates parameter sharing between nodes
func tutorialParam() {
	var w, r gocgraph.GElement

	pipeline := gocgraph.Factory.Create()

	w = &MyWriteParamNode{}
	pipeline.RegisterGElement(w, []gocgraph.GElement{}, "WriteNode")

	r = &MyReadParamNode{}
	pipeline.RegisterGElement(r, []gocgraph.GElement{w}, "ReadNode")

	pipeline.Process(5) // Run 5 times like the C++ version
	gocgraph.Factory.Remove(pipeline)
}

func main() {
	tutorialParam()
}
