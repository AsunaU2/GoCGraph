# go_cgraph_lite

A lightweight DAG (Directed Acyclic Graph) execution framework for Go, ported from C++ CGraph-lite library.

## Features

- 🚀 **Lightweight DAG execution framework**
- 🔄 **Thread-safe parameter sharing between nodes**
- ⚡ **Concurrent execution with worker pool**
- 🏗️ **Easy-to-use builder pattern for pipeline construction**
- 🔧 **Compatible with original C++ CGraph-lite interface**

## Quick Start

### Basic Example

```go
package main

import "fmt"

// Define your custom node
type MyNode struct {
    BaseGElement
}

func (n *MyNode) Run() *CStatus {
    fmt.Printf("%s: Hello World!\n", n.GetName())
    return NewCStatus()
}

func main() {
    // Create pipeline
    pipeline := Factory.Create()
    
    // Register nodes
    node := &MyNode{}
    pipeline.RegisterGElement(node, []GElement{}, "MyNode")
    
    // Execute
    pipeline.Process(1)
    
    // Cleanup
    Factory.Remove(pipeline)
}
```

### Parameter Sharing Example

See `T02_Param.go` for a complete example of sharing parameters between nodes.

## Examples

- `T01_Simple.go` - Basic node execution example
- `T02_Param.go` - Parameter sharing between nodes example

## Installation

```bash
go get github.com/AsunaU2/GoCGraph
```

## Building

```bash
go build .
```

## Running Examples

```bash
# Run simple example
go run T01_Simple.go go_cgraph_lite.go

# Run parameter example  
go run T02_Param.go go_cgraph_lite.go
```

## Core Concepts

### GElement (Node)
Basic execution unit in the DAG. Implement the `Run()` method to define your logic.

### GPipeline
Manages the entire DAG execution flow, handles dependencies and concurrent execution.

### GParam
Thread-safe parameter sharing mechanism between nodes.

## License

MIT License

## Contributing

Pull requests are welcome! For major changes, please open an issue first.
