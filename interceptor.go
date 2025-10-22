package main

import (
	"bytes"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"
)

type InterceptRule struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Enabled         bool   `json:"enabled"`
	URLPattern      string `json:"urlPattern"`
	ActionType      string `json:"actionType"` // "findReplace", "redirect", "responseReplace"
	FindText        string `json:"findText,omitempty"`
	ReplaceText     string `json:"replaceText,omitempty"`
	UseRegex        bool   `json:"useRegex,omitempty"`
	ReplaceAll      bool   `json:"replaceAll,omitempty"`
	ResponseContent string `json:"responseContent,omitempty"`
	RedirectURL     string `json:"redirectUrl,omitempty"`
	WebhookURL      string `json:"webhookUrl,omitempty"`
	WebhookEnabled  bool   `json:"webhookEnabled,omitempty"`

	compiledURLRegex  *regexp.Regexp
	compiledFindRegex *regexp.Regexp
}

type Interceptor struct {
	rules []InterceptRule
	mu    sync.RWMutex
}

func NewInterceptor() *Interceptor {
	return &Interceptor{
		rules: []InterceptRule{},
	}
}

func (i *Interceptor) SetRules(rules []InterceptRule) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	for idx := range rules {
		if rules[idx].ActionType == "findReplace" && rules[idx].UseRegex && rules[idx].FindText != "" {
			compiled, err := regexp.Compile(rules[idx].FindText)
			if err != nil {
				return err
			}
			rules[idx].compiledFindRegex = compiled
		}
	}

	i.rules = rules
	return nil
}

func (i *Interceptor) GetRules() []InterceptRule {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.rules
}

func (i *Interceptor) matchURL(url string, rule InterceptRule) bool {
	if !rule.Enabled {
		return false
	}
	return matchWildcard(url, rule.URLPattern)
}

func matchWildcard(text, pattern string) bool {
	// Convert wildcard pattern to regex
	// * matches any characters
	regexPattern := "^" + regexp.QuoteMeta(pattern) + "$"
	regexPattern = strings.ReplaceAll(regexPattern, "\\*", ".*")

	matched, _ := regexp.MatchString(regexPattern, text)
	return matched
}

func (i *Interceptor) InterceptResponse(resp *http.Response, req *http.Request) error {
	i.mu.RLock()
	defer i.mu.RUnlock()

	url := req.URL.String()

	for _, rule := range i.rules {
		if !i.matchURL(url, rule) {
			continue
		}

		// Read original body
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		resp.Body.Close()

		bodyStr := string(bodyBytes)
		modified := false

		// Apply transformation
		if rule.ActionType == "redirect" && rule.RedirectURL != "" {
			resp.StatusCode = 302
			resp.Status = "302 Found"
			resp.Header.Set("Location", rule.RedirectURL)
			resp.Body = io.NopCloser(bytes.NewReader([]byte{}))
			resp.ContentLength = 0
			modified = true
		} else if rule.ActionType == "responseReplace" && rule.ResponseContent != "" {
			bodyStr = rule.ResponseContent
			modified = true
		} else if rule.ActionType == "findReplace" && rule.FindText != "" {
			if rule.UseRegex && rule.compiledFindRegex != nil {
				if rule.ReplaceAll {
					bodyStr = rule.compiledFindRegex.ReplaceAllString(bodyStr, rule.ReplaceText)
				} else {
					bodyStr = rule.compiledFindRegex.ReplaceAllStringFunc(bodyStr, func(s string) string {
						return rule.ReplaceText
					})
				}
			} else {
				if rule.ReplaceAll {
					bodyStr = strings.ReplaceAll(bodyStr, rule.FindText, rule.ReplaceText)
				} else {
					bodyStr = strings.Replace(bodyStr, rule.FindText, rule.ReplaceText, 1)
				}
			}
			modified = true
		}

		// Update response
		if modified && rule.ActionType != "redirect" {
			newBody := []byte(bodyStr)
			resp.Body = io.NopCloser(bytes.NewReader(newBody))
			resp.ContentLength = int64(len(newBody))
			resp.Header.Del("Content-Encoding")
		} else if !modified {
			resp.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		}

		// Webhook notification (async)
		if rule.WebhookEnabled && rule.WebhookURL != "" {
			go sendWebhook(rule.WebhookURL, url, bodyStr)
		}

		break // Only apply first matching rule
	}

	return nil
}

func sendWebhook(webhookURL, requestURL, body string) {
	// Simple webhook implementation
	payload := map[string]string{
		"url":  requestURL,
		"body": body,
	}

	// This is a simplified version - in production you'd want proper error handling
	_ = payload
	// http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
}
