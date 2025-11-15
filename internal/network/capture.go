package network

import (
	"context"
	"encoding/json"
	"fmt"
	"netcaptor/internal/types"
	"sync"
	"time"
)

type NetworkRequest struct {
	ID      string            `json:"id"`
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body,omitempty"`
	Type    string            `json:"type"`
	Time    int64             `json:"time"`
	Domain  string            `json:"domain"`
	Path    string            `json:"path"`
}

type NetworkCapture struct {
	ctx        context.Context
	requests   []types.NetworkRequest
	responses  []types.NetworkResponse
	entries    []types.NetworkEntry
	mu         sync.RWMutex
	maxEntries int
}

func NewNetworkCapture() *NetworkCapture {
	return &NetworkCapture{
		requests:   make([]types.NetworkRequest, 0),
		responses:  make([]types.NetworkResponse, 0),
		entries:    make([]types.NetworkEntry, 0),
		maxEntries: 30, // 默认保存30条
	}
}

func (nc *NetworkCapture) RecordRequest(reqJSON string) error {
	var req types.NetworkRequest
	if err := json.Unmarshal([]byte(reqJSON), &req); err != nil {
		return err
	}

	if req.ID == "" {
		req.ID = generateID()
	}

	nc.mu.Lock()
	nc.requests = append(nc.requests, req)
	nc.mu.Unlock()

	return nil
}

func (nc *NetworkCapture) RecordResponse(respJSON string) error {
	var resp types.NetworkResponse
	if err := json.Unmarshal([]byte(respJSON), &resp); err != nil {
		return err
	}

	nc.mu.Lock()
	nc.responses = append(nc.responses, resp)
	nc.mu.Unlock()

	return nil
}

// AddRequest 实现 types.CaptureHandler 接口
func (nc *NetworkCapture) AddRequest(req *types.NetworkRequest) {
	nc.mu.Lock()
	defer nc.mu.Unlock()
	nc.requests = append(nc.requests, *req)
	if len(nc.requests) > nc.maxEntries {
		nc.requests = nc.requests[len(nc.requests)-nc.maxEntries:]
	}
}

// GetRequests 实现 types.CaptureHandler 接口
func (nc *NetworkCapture) GetRequests() []types.NetworkRequest {
	nc.mu.Lock()
	defer nc.mu.Unlock()
	return nc.requests
}

// AddResponse 实现 types.CaptureHandler 接口
func (nc *NetworkCapture) AddResponse(resp *types.NetworkResponse) {
	nc.mu.Lock()
	defer nc.mu.Unlock()
	nc.responses = append(nc.responses, *resp)
	if len(nc.responses) > nc.maxEntries {
		nc.responses = nc.responses[len(nc.responses)-nc.maxEntries:]
	}
}

// AddEntry 实现 types.CaptureHandler 接口
func (nc *NetworkCapture) AddEntry(entry *types.NetworkEntry) {
	nc.mu.Lock()
	defer nc.mu.Unlock()
	nc.entries = append(nc.entries, *entry)
	if len(nc.entries) > nc.maxEntries {
		nc.entries = nc.entries[len(nc.entries)-nc.maxEntries:]
	}
}

// GetEntries 实现 types.CaptureHandler 接口
func (nc *NetworkCapture) GetEntries() []types.NetworkEntry {
	nc.mu.Lock()
	defer nc.mu.Unlock()
	return nc.entries
}

// GetResponses 获取响应列表
func (nc *NetworkCapture) GetResponses() []types.NetworkResponse {
	nc.mu.Lock()
	defer nc.mu.Unlock()
	return nc.responses
}

// Clear 清空捕获数据
func (nc *NetworkCapture) Clear() {
	nc.mu.Lock()
	defer nc.mu.Unlock()
	nc.requests = []types.NetworkRequest{}
	nc.responses = []types.NetworkResponse{}
	nc.entries = []types.NetworkEntry{}
}

// GetMaxEntries 获取最大条目数
func (nc *NetworkCapture) GetMaxEntries() int {
	nc.mu.RLock()
	defer nc.mu.RUnlock()
	return nc.maxEntries
}

// SetMaxEntries 设置最大条目数
func (nc *NetworkCapture) SetMaxEntries(max int) {
	nc.mu.Lock()
	defer nc.mu.Unlock()
	nc.maxEntries = max
}

// GetInjectionScript 实现 types.CaptureHandler 接口
func (nc *NetworkCapture) GetInjectionScript() string {
	return `
		// Network capture injection script
		console.log('Network capture enabled');
	`
}

func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
