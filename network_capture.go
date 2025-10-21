package main

import (
	"context"
	"encoding/json"
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

type NetworkResponse struct {
	ID          string            `json:"id"`
	URL         string            `json:"url"`
	Status      int               `json:"status"`
	StatusText  string            `json:"statusText"`
	Headers     map[string]string `json:"headers"`
	Body        string            `json:"body,omitempty"`
	Size        int               `json:"size"`
	Duration    int64             `json:"duration"`
	ContentType string            `json:"contentType"`
}

type NetworkEntry struct {
	ID         string          `json:"id"`
	URL        string          `json:"url"`
	Method     string          `json:"method"`
	Status     int             `json:"status"`
	StatusText string          `json:"statusText"`
	Type       string          `json:"type"`
	Size       int             `json:"size"`
	Time       int64           `json:"time"`
	Duration   int64           `json:"duration"`
	Domain     string          `json:"domain"`
	Path       string          `json:"path"`
	Request    NetworkRequest  `json:"request"`
	Response   NetworkResponse `json:"response"`
}

type NetworkCapture struct {
	ctx       context.Context
	requests  []NetworkRequest
	responses []NetworkResponse
	entries   []NetworkEntry
	mu        sync.RWMutex
}

func NewNetworkCapture() *NetworkCapture {
	return &NetworkCapture{
		requests:  make([]NetworkRequest, 0),
		responses: make([]NetworkResponse, 0),
		entries:   make([]NetworkEntry, 0),
	}
}

func (nc *NetworkCapture) RecordRequest(reqJSON string) error {
	var req NetworkRequest
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
	var resp NetworkResponse
	if err := json.Unmarshal([]byte(respJSON), &resp); err != nil {
		return err
	}

	nc.mu.Lock()
	nc.responses = append(nc.responses, resp)
	nc.mu.Unlock()

	return nil
}

func (nc *NetworkCapture) GetRequests() []NetworkRequest {
	nc.mu.RLock()
	defer nc.mu.RUnlock()
	return nc.requests
}

func (nc *NetworkCapture) GetResponses() []NetworkResponse {
	nc.mu.RLock()
	defer nc.mu.RUnlock()
	return nc.responses
}

func (nc *NetworkCapture) Clear() {
	nc.mu.Lock()
	nc.requests = make([]NetworkRequest, 0)
	nc.responses = make([]NetworkResponse, 0)
	nc.entries = make([]NetworkEntry, 0)
	nc.mu.Unlock()
}

func (nc *NetworkCapture) GetEntries() []NetworkEntry {
	nc.mu.RLock()
	defer nc.mu.RUnlock()
	return nc.entries
}

func (nc *NetworkCapture) AddEntry(entry NetworkEntry) {
	nc.mu.Lock()
	nc.entries = append(nc.entries, entry)
	nc.mu.Unlock()
}

func (nc *NetworkCapture) GetInjectionScript() string {
	return `
(function() {
	const capturedRequests = new Map();
	
	// 拦截 Fetch
	const originalFetch = window.fetch;
	window.fetch = function(...args) {
		const startTime = Date.now();
		const reqId = 'req_' + Date.now() + '_' + Math.random();
		const url = typeof args[0] === 'string' ? args[0] : args[0].url;
		const options = args[1] || {};
		
		const headers = {};
		if (options.headers) {
			if (options.headers instanceof Headers) {
				options.headers.forEach((v, k) => headers[k] = v);
			} else {
				Object.assign(headers, options.headers);
			}
		}
		
		window.RecordRequest(JSON.stringify({
			id: reqId,
			url: url,
			method: options.method || 'GET',
			headers: headers,
			body: options.body ? String(options.body).substring(0, 1000) : '',
			type: 'fetch',
			time: startTime
		}));

		return originalFetch.apply(this, args).then(response => {
			const clonedResponse = response.clone();
			clonedResponse.text().then(body => {
				const respHeaders = {};
				response.headers.forEach((v, k) => respHeaders[k] = v);
				
				window.RecordResponse(JSON.stringify({
					id: reqId,
					url: url,
					status: response.status,
					headers: respHeaders,
					body: body.substring(0, 5000),
					size: body.length,
					duration: Date.now() - startTime
				}));
			}).catch(() => {});
			return response;
		});
	};

	// 拦截 XMLHttpRequest
	const originalXHROpen = XMLHttpRequest.prototype.open;
	const originalXHRSend = XMLHttpRequest.prototype.send;
	const originalXHRSetRequestHeader = XMLHttpRequest.prototype.setRequestHeader;

	XMLHttpRequest.prototype.open = function(method, url, ...rest) {
		this._captureURL = url;
		this._captureMethod = method;
		this._captureHeaders = {};
		this._captureStartTime = Date.now();
		this._captureId = 'xhr_' + Date.now() + '_' + Math.random();
		return originalXHROpen.apply(this, [method, url, ...rest]);
	};

	XMLHttpRequest.prototype.setRequestHeader = function(header, value) {
		if (this._captureHeaders) {
			this._captureHeaders[header] = value;
		}
		return originalXHRSetRequestHeader.apply(this, arguments);
	};

	XMLHttpRequest.prototype.send = function(body) {
		const xhr = this;
		
		window.RecordRequest(JSON.stringify({
			id: xhr._captureId,
			url: xhr._captureURL,
			method: xhr._captureMethod,
			headers: xhr._captureHeaders || {},
			body: body ? String(body).substring(0, 1000) : '',
			type: 'xhr',
			time: xhr._captureStartTime
		}));

		const originalOnReadyStateChange = xhr.onreadystatechange;
		xhr.onreadystatechange = function() {
			if (xhr.readyState === 4) {
				const respHeaders = {};
				const headerStr = xhr.getAllResponseHeaders();
				if (headerStr) {
					headerStr.split('\r\n').forEach(line => {
						const parts = line.split(': ');
						if (parts.length === 2) {
							respHeaders[parts[0]] = parts[1];
						}
					});
				}
				
				window.RecordResponse(JSON.stringify({
					id: xhr._captureId,
					url: xhr._captureURL,
					status: xhr.status,
					headers: respHeaders,
					body: xhr.responseText ? xhr.responseText.substring(0, 5000) : '',
					size: xhr.responseText ? xhr.responseText.length : 0,
					duration: Date.now() - xhr._captureStartTime
				}));
			}
			if (originalOnReadyStateChange) {
				originalOnReadyStateChange.apply(this, arguments);
			}
		};

		return originalXHRSend.apply(this, arguments);
	};

	console.log('[Network Capture] Initialized');
})();
`
}

func generateID() string {
	return time.Now().Format("20060102150405") + "_" + string(time.Now().UnixNano()%1000)
}
