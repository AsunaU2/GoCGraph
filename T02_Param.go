//go:build T02
// +build T02

package main

import (
	"fmt"
)

const ParamKey = "param_key"

// MyParam implements the GParam interface for parameter sharing
type MyParam struct {
	BaseGParam
	Val  int `json:"val"`
	Loop int `json:"loop"`
}

// Reset resets the parameter values
func (p *MyParam) Reset(curStatus *CStatus) {
	p.Val = 0
}

// MyReadParamNode reads shared parameters
type MyReadParamNode struct {
	BaseGElement
}

// Init initializes the node and creates shared parameters
func (n *MyReadParamNode) Init() *CStatus {
	return n.CreateGParam(ParamKey, &MyParam{})
}

// Run executes the read logic
func (n *MyReadParamNode) Run() *CStatus {
	param := n.GetGParam(ParamKey)
	if param == nil {
		return NewCStatusWithError("param not found")
	}

	myParam, ok := param.(*MyParam)
	if !ok {
		return NewCStatusWithError("param type mismatch")
	}

	// Lock for thread safety
	myParam.GetLock().RLock()
	fmt.Printf("%s: val = %d, loop = %d\n", n.GetName(), myParam.Val, myParam.Loop)
	myParam.GetLock().RUnlock()

	return NewCStatus()
}

// MyWriteParamNode writes to shared parameters
type MyWriteParamNode struct {
	BaseGElement
}

// Run executes the write logic
func (n *MyWriteParamNode) Run() *CStatus {
	param := n.GetGParam(ParamKey)
	if param == nil {
		return NewCStatusWithError("param not found")
	}

	myParam, ok := param.(*MyParam)
	if !ok {
		return NewCStatusWithError("param type mismatch")
	}

	// Lock for thread safety
	myParam.GetLock().Lock()
	myParam.Val++
	myParam.Loop++
	fmt.Printf("%s: val = %d, loop = %d\n", n.GetName(), myParam.Val, myParam.Loop)
	myParam.GetLock().Unlock()

	return NewCStatus()
}

// tutorialParam demonstrates parameter sharing between nodes
func tutorialParam() {
	var w, r GElement

	pipeline := Factory.Create()

	w = &MyWriteParamNode{}
	pipeline.RegisterGElement(w, []GElement{}, "WriteNode")

	r = &MyReadParamNode{}
	pipeline.RegisterGElement(r, []GElement{w}, "ReadNode")

	pipeline.Process(5) // Run 5 times like the C++ version
	Factory.Remove(pipeline)
}

func main() {
	tutorialParam()
}
