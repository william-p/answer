/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package webhook

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/apache/answer/internal/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebhookService_SendOnce_Success(t *testing.T) {
	var mu sync.Mutex
	var receivedMethod string
	var receivedBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		receivedMethod = r.Method
		receivedBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	svc := NewWebhookService()

	webhook := &entity.Webhook{
		URL:    server.URL,
		Method: "POST",
	}

	payload := &WebhookPayload{
		Event:     "question.create",
		Timestamp: 1234567890,
		Data: map[string]any{
			"question_id": "123",
			"title":       "Test Question",
		},
	}

	err := svc.SendOnce(webhook, payload)
	require.NoError(t, err)

	mu.Lock()
	assert.Equal(t, "POST", receivedMethod)

	var got WebhookPayload
	err = json.Unmarshal(receivedBody, &got)
	require.NoError(t, err)
	assert.Equal(t, "question.create", got.Event)
	assert.Equal(t, int64(1234567890), got.Timestamp)
	assert.Equal(t, "123", got.Data["question_id"])
	mu.Unlock()

	// Test PUT method
	webhook.Method = "PUT"
	err = svc.SendOnce(webhook, payload)
	require.NoError(t, err)

	mu.Lock()
	assert.Equal(t, "PUT", receivedMethod)
	mu.Unlock()
}

func TestWebhookService_SendOnce_WithHeader(t *testing.T) {
	var mu sync.Mutex
	var receivedHeaderName string
	var receivedHeaderValue string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		receivedHeaderName = "X-Webhook-Token"
		receivedHeaderValue = r.Header.Get("X-Webhook-Token")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	svc := NewWebhookService()

	webhook := &entity.Webhook{
		URL:         server.URL,
		Method:      "POST",
		HeaderName:  "X-Webhook-Token",
		HeaderValue: "my-secret-token",
	}

	payload := &WebhookPayload{
		Event:     "question.create",
		Timestamp: 1234567890,
		Data:      map[string]any{},
	}

	err := svc.SendOnce(webhook, payload)
	require.NoError(t, err)

	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, "X-Webhook-Token", receivedHeaderName)
	assert.Equal(t, "my-secret-token", receivedHeaderValue)
}

func TestWebhookService_SendOnce_NonSuccessStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer server.Close()

	svc := NewWebhookService()
	wh := &entity.Webhook{ID: "fail-test", URL: server.URL, Method: "POST"}
	payload := &WebhookPayload{Event: "question.create", Timestamp: 1, Data: map[string]any{}}

	err := svc.SendOnce(wh, payload)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "502")
}
