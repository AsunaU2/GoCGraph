// Package gocgraph provides a lightweight DAG (Directed Acyclic Graph) execution framework

package gocgraph

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"sync"
	"sync/atomic"
)

// CStatus represents the execution result status
type CStatus struct {
	errorCode int
	errorInfo string
}

// NewCStatus creates a new successful status
func NewCStatus() *CStatus {
	return &CStatus{errorCode: 0, errorInfo: ""}
}

// NewCStatusWithError creates a new error status
func NewCStatusWithError(errorInfo string) *CStatus {
	return &CStatus{errorCode: -1, errorInfo: errorInfo}
}

// Combine combines two status objects
func (s *CStatus) Combine(other *CStatus) *CStatus {
	if s.IsOK() && !other.IsOK() {
		s.errorCode = other.errorCode
		s.errorInfo = other.errorInfo
	}
	return s
}

// IsOK checks if the status is successful
func (s *CStatus) IsOK() bool {
	return s.errorCode == 0
}

// GetCode returns the error code
func (s *CStatus) GetCode() int {
	return s.errorCode
}

// GetInfo returns the error information
func (s *CStatus) GetInfo() string {
	return s.errorInfo
}

// GParam is the base interface for shared parameters between nodes
type GParam interface {
	Setup() *CStatus
	Reset(curStatus *CStatus)
}

// BaseGParam provides a default implementation of GParam
type BaseGParam struct {
	paramSharedLock sync.RWMutex
}

// Setup provides default parameter initialization
func (p *BaseGParam) Setup() *CStatus {
	return NewCStatus()
}

// Reset provides default parameter reset
func (p *BaseGParam) Reset(curStatus *CStatus) {
	// Default implementation does nothing
}

// GetLock returns the shared lock for thread safety
func (p *BaseGParam) GetLock() *sync.RWMutex {
	return &p.paramSharedLock
}

// GParamManager manages the lifecycle of all shared parameters
type GParamManager struct {
	params map[string]GParam
	mutex  sync.RWMutex
}

// NewGParamManager creates a new parameter manager
func NewGParamManager() *GParamManager {
	return &GParamManager{
		params: make(map[string]GParam),
	}
}

// Create creates a parameter with the specified key and type
func (pm *GParamManager) Create(key string, param GParam) *CStatus {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	if existingParam, exists := pm.params[key]; exists {
		// 检查是否是同一类型，防止重复创建不同类型的同名参数
		if reflect.TypeOf(existingParam) == reflect.TypeOf(param) {
			return NewCStatus()
		}
		return NewCStatusWithError(fmt.Sprintf("create [%s] param duplicate", key))
	}

	pm.params[key] = param
	return NewCStatus()
}

// Get retrieves a parameter by key
func (pm *GParamManager) Get(key string) GParam {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	param, exists := pm.params[key]
	if !exists {
		return nil
	}
	return param
}

// Setup initializes all parameters
func (pm *GParamManager) Setup() *CStatus {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	status := NewCStatus()
	for _, param := range pm.params {
		status.Combine(param.Setup())
	}
	return status
}

// Reset resets all parameters
func (pm *GParamManager) Reset(curStatus *CStatus) {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	for _, param := range pm.params {
		param.Reset(curStatus)
	}
}

// Schedule implements a thread pool and task queue for concurrent execution
type Schedule struct {
	workers   int
	taskQueue chan func()
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
}

// NewSchedule creates a new scheduler with specified number of workers
func NewSchedule(numWorkers int) *Schedule {
	if numWorkers <= 0 {
		numWorkers = runtime.NumCPU()
	}

	ctx, cancel := context.WithCancel(context.Background())
	s := &Schedule{
		workers:   numWorkers,
		taskQueue: make(chan func(), 1000),
		ctx:       ctx,
		cancel:    cancel,
	}

	for i := 0; i < numWorkers; i++ {
		s.wg.Add(1)
		go s.worker()
	}

	return s
}

// worker is the worker goroutine function
func (s *Schedule) worker() {
	defer s.wg.Done()

	for {
		select {
		case task := <-s.taskQueue:
			if task != nil {
				task()
			}
		case <-s.ctx.Done():
			return
		}
	}
}

// Commit submits a task to the task queue
func (s *Schedule) Commit(task func()) {
	select {
	case s.taskQueue <- task:
	case <-s.ctx.Done():
		return
	default:
		go task()
	}
}

// Stop stops all worker goroutines
func (s *Schedule) Stop() {
	s.cancel()
	close(s.taskQueue)
	s.wg.Wait()
}

// GElement represents a node in the DAG
type GElement interface {
	Init() *CStatus
	Run() *CStatus
	Destroy() *CStatus
	GetName() string

	// Parameter management methods
	CreateGParam(key string, param GParam) *CStatus
	GetGParam(key string) GParam
}

// GNode is an alias for GElement
type GNode = GElement

// elementInfo contains internal dependency management information
type elementInfo struct {
	runBefore    []*elementInfo // 当前节点完成后需要执行的节点集合
	dependence   []*elementInfo // 当前节点依赖的节点集合
	leftDepend   int64          // 剩余依赖数量（原子变量，线程安全）
	name         string         // 节点名称
	paramManager *GParamManager // 参数管理器指针
	element      GElement       // 对应的元素接口
}

// BaseGElement provides a basic implementation that other nodes can embed
type BaseGElement struct {
	name         string
	paramManager *GParamManager
}

// Init provides default initialization
func (e *BaseGElement) Init() *CStatus {
	return NewCStatus()
}

// Destroy provides default destruction
func (e *BaseGElement) Destroy() *CStatus {
	return NewCStatus()
}

// GetName returns the node name
func (e *BaseGElement) GetName() string {
	return e.name
}

// SetName sets the node name
func (e *BaseGElement) SetName(name string) {
	e.name = name
}

// SetParamManager sets the parameter manager
func (e *BaseGElement) SetParamManager(pm *GParamManager) {
	e.paramManager = pm
}

// CreateGParam creates a parameter with the specified key and type
func (e *BaseGElement) CreateGParam(key string, param GParam) *CStatus {
	if e.paramManager == nil {
		return NewCStatusWithError("param manager is nil")
	}
	return e.paramManager.Create(key, param)
}

// GetGParam retrieves a parameter by key
func (e *BaseGElement) GetGParam(key string) GParam {
	if e.paramManager == nil {
		return nil
	}
	return e.paramManager.Get(key)
}

// GPipeline manages the entire DAG execution flow
type GPipeline struct {
	elements     []*elementInfo            // 所有图元素的内部信息
	elementMap   map[GElement]*elementInfo // 元素到内部信息的映射
	schedule     *Schedule                 // 任务调度器
	finishedSize int64                     // 已完成节点数量
	executeMutex sync.Mutex                // 执行互斥锁
	executeCond  *sync.Cond                // 执行条件变量
	status       *CStatus                  // 执行状态
	paramManager *GParamManager            // 参数管理器
}

// NewGPipeline creates a new pipeline
func NewGPipeline() *GPipeline {
	p := &GPipeline{
		elements:     make([]*elementInfo, 0),
		elementMap:   make(map[GElement]*elementInfo),
		paramManager: NewGParamManager(),
		status:       NewCStatus(),
	}
	p.executeCond = sync.NewCond(&p.executeMutex)
	return p
}

// Process executes the pipeline for the specified number of times
func (p *GPipeline) Process(times int) *CStatus {
	if times <= 0 {
		times = 1
	}

	p.init()
	for times > 0 && p.status.IsOK() {
		p.run()
		times--
	}
	p.destroy()
	return p.status
}

// RegisterGElement registers a graph element with dependencies
func (p *GPipeline) RegisterGElement(element GElement, depends []GElement, name string) *CStatus {
	// Set up the element with name and parameter manager
	if setter, ok := element.(interface{ SetName(string) }); ok {
		setter.SetName(name)
	}
	if setter, ok := element.(interface{ SetParamManager(*GParamManager) }); ok {
		setter.SetParamManager(p.paramManager)
	}

	// Create element info
	info := &elementInfo{
		name:         name,
		paramManager: p.paramManager,
		element:      element,
		runBefore:    make([]*elementInfo, 0),
		dependence:   make([]*elementInfo, 0),
	}

	// Set up dependencies
	for _, dep := range depends {
		if depInfo, exists := p.elementMap[dep]; exists {
			info.dependence = append(info.dependence, depInfo)
			depInfo.runBefore = append(depInfo.runBefore, info)
		} else {
			return NewCStatusWithError("dependency element not found")
		}
	}

	atomic.StoreInt64(&info.leftDepend, int64(len(info.dependence)))

	// Add to pipeline
	p.elements = append(p.elements, info)
	p.elementMap[element] = info

	return NewCStatus()
}

// init initializes all elements and scheduler
func (p *GPipeline) init() {
	p.status = NewCStatus()
	for _, info := range p.elements {
		p.status.Combine(info.element.Init())
	}
	p.schedule = NewSchedule(8) // 创建调度器，默认8个工作线程
}

// run executes one complete DAG execution
func (p *GPipeline) run() {
	p.setup()      // 设置
	p.executeAll() // 执行所有节点
	p.reset()      // 重置
}

// destroy cleans up all elements and scheduler
func (p *GPipeline) destroy() {
	for _, info := range p.elements {
		p.status.Combine(info.element.Destroy())
	}
	if p.schedule != nil {
		p.schedule.Stop()
		p.schedule = nil
	}
}

// executeAll starts execution from nodes with no dependencies
func (p *GPipeline) executeAll() {
	for _, info := range p.elements {
		if len(info.dependence) == 0 {
			p.schedule.Commit(func() {
				p.execute(info)
			})
		}
	}
}

// execute executes a single node

func (p *GPipeline) execute(info *elementInfo) {
	if !p.status.IsOK() {
		return
	}

	p.status.Combine(info.element.Run())

	for _, cur := range info.runBefore {
		if atomic.AddInt64(&cur.leftDepend, -1) <= 0 {
			p.schedule.Commit(func() {
				p.execute(cur)
			})
		}
	}

	p.executeMutex.Lock()
	atomic.AddInt64(&p.finishedSize, 1)
	if atomic.LoadInt64(&p.finishedSize) >= int64(len(p.elements)) || !p.status.IsOK() {
		p.executeCond.Signal()
	}
	p.executeMutex.Unlock()
}

// setup prepares for execution
func (p *GPipeline) setup() {
	atomic.StoreInt64(&p.finishedSize, 0)
	for _, info := range p.elements {
		atomic.StoreInt64(&info.leftDepend, int64(len(info.dependence)))
	}
	p.status.Combine(p.paramManager.Setup())
}

// reset cleans up after execution
func (p *GPipeline) reset() {
	p.executeMutex.Lock()
	for atomic.LoadInt64(&p.finishedSize) < int64(len(p.elements)) && p.status.IsOK() {
		p.executeCond.Wait()
	}
	p.executeMutex.Unlock()
	p.paramManager.Reset(p.status)
}

// GPipelineFactory provides factory methods for pipeline creation and destruction
type GPipelineFactory struct{}

// Create creates a new pipeline instance
func (f *GPipelineFactory) Create() *GPipeline {
	return NewGPipeline()
}

// Remove destroys a pipeline instance
func (f *GPipelineFactory) Remove(pipeline *GPipeline) *CStatus {
	if pipeline != nil {
		pipeline.destroy()
	}
	return NewCStatus()
}

// Global factory instance for convenience
var Factory = &GPipelineFactory{}
