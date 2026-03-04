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
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockEventQueueService implements eventqueue.Service for testing
type mockEventQueueService struct {
	handler func(ctx context.Context, msg *schema.EventMsg) error
}

func (m *mockEventQueueService) Send(ctx context.Context, msg *schema.EventMsg) {}
func (m *mockEventQueueService) RegisterHandler(handler func(ctx context.Context, msg *schema.EventMsg) error) {
	m.handler = handler
}
func (m *mockEventQueueService) Close() {}

// mockWebhookRepo implements webhook.WebhookRepo for testing
type mockWebhookRepo struct {
	webhooks []*entity.Webhook
}

func (m *mockWebhookRepo) CreateWebhook(ctx context.Context, webhook *entity.Webhook) error {
	return nil
}
func (m *mockWebhookRepo) UpdateWebhook(ctx context.Context, webhook *entity.Webhook) error {
	return nil
}
func (m *mockWebhookRepo) DeleteWebhook(ctx context.Context, id string) error { return nil }
func (m *mockWebhookRepo) GetWebhook(ctx context.Context, id string) (*entity.Webhook, bool, error) {
	return nil, false, nil
}
func (m *mockWebhookRepo) ListWebhooks(ctx context.Context) ([]*entity.Webhook, error) {
	return m.webhooks, nil
}
func (m *mockWebhookRepo) GetEnabledWebhooks(ctx context.Context) ([]*entity.Webhook, error) {
	return m.webhooks, nil
}

func TestWebhookEventHandler_FilterEvents(t *testing.T) {
	var mu sync.Mutex
	var receivedPayloads []WebhookPayload

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var p WebhookPayload
		_ = json.Unmarshal(body, &p)
		mu.Lock()
		receivedPayloads = append(receivedPayloads, p)
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	repo := &mockWebhookRepo{
		webhooks: []*entity.Webhook{
			{
				ID:      "1",
				Name:    "Create Only",
				URL:     server.URL,
				Method:  "POST",
				Events:  `["question.create"]`,
				Enabled: true,
			},
			{
				ID:      "2",
				Name:    "All Question Events",
				URL:     server.URL,
				Method:  "POST",
				Events:  `["question.create","question.update","question.delete","question.accept"]`,
				Enabled: true,
			},
		},
	}

	mockQueue := &mockEventQueueService{}
	handler := NewWebhookEventHandler(repo, NewWebhookService(), mockQueue)

	// Send a question.create event - both webhooks should receive it
	msg := schema.NewEvent(constant.EventQuestionCreate, "user1").
		TID("q1").QID("q1", "user1")
	err := handler.Handle(context.Background(), msg)
	require.NoError(t, err)

	handler.Wait()

	mu.Lock()
	assert.Len(t, receivedPayloads, 2, "Both webhooks should receive question.create")
	for _, p := range receivedPayloads {
		assert.Equal(t, "question.create", p.Event)
	}
	receivedPayloads = nil
	mu.Unlock()

	// Send a question.update event - only webhook #2 should receive it
	msg2 := schema.NewEvent(constant.EventQuestionUpdate, "user1").
		TID("q1").QID("q1", "user1")
	err = handler.Handle(context.Background(), msg2)
	require.NoError(t, err)

	handler.Wait()

	mu.Lock()
	assert.Len(t, receivedPayloads, 1, "Only webhook #2 should receive question.update")
	if len(receivedPayloads) > 0 {
		assert.Equal(t, "question.update", receivedPayloads[0].Event)
	}
	mu.Unlock()
}

func TestWebhookEventHandler_AcceptUnacceptEvents(t *testing.T) {
	var mu sync.Mutex
	var receivedPayloads []WebhookPayload

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var p WebhookPayload
		_ = json.Unmarshal(body, &p)
		mu.Lock()
		receivedPayloads = append(receivedPayloads, p)
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	repo := &mockWebhookRepo{
		webhooks: []*entity.Webhook{
			{
				ID:      "accept-only",
				Name:    "Accept Only",
				URL:     server.URL,
				Method:  "POST",
				Events:  `["answer.accept"]`,
				Enabled: true,
			},
			{
				ID:      "unaccept-only",
				Name:    "Unaccept Only",
				URL:     server.URL,
				Method:  "POST",
				Events:  `["answer.unaccept"]`,
				Enabled: true,
			},
			{
				ID:      "both",
				Name:    "Accept and Unaccept",
				URL:     server.URL,
				Method:  "POST",
				Events:  `["answer.accept","answer.unaccept"]`,
				Enabled: true,
			},
		},
	}

	mockQueue := &mockEventQueueService{}
	handler := NewWebhookEventHandler(repo, NewWebhookService(), mockQueue)

	// Test answer.accept event
	acceptMsg := schema.NewEvent(constant.EventAnswerAccept, "user1").
		TID("a1").QID("q1", "quser1").AID("a1", "auser1")
	err := handler.Handle(context.Background(), acceptMsg)
	require.NoError(t, err)

	handler.Wait()

	mu.Lock()
	assert.Len(t, receivedPayloads, 2, "accept-only and both webhooks should receive answer.accept")
	for _, p := range receivedPayloads {
		assert.Equal(t, "answer.accept", p.Event)
		assert.Contains(t, p.Data, "question_id")
		assert.Contains(t, p.Data, "answer_id")
	}
	receivedPayloads = nil
	mu.Unlock()

	// Test answer.unaccept event
	unacceptMsg := schema.NewEvent(constant.EventAnswerUnaccept, "user1").
		TID("a1").QID("q1", "quser1").AID("a1", "auser1")
	err = handler.Handle(context.Background(), unacceptMsg)
	require.NoError(t, err)

	handler.Wait()

	mu.Lock()
	assert.Len(t, receivedPayloads, 2, "unaccept-only and both webhooks should receive answer.unaccept")
	for _, p := range receivedPayloads {
		assert.Equal(t, "answer.unaccept", p.Event)
		assert.Contains(t, p.Data, "question_id")
		assert.Contains(t, p.Data, "answer_id")
	}
	mu.Unlock()
}

func TestWebhookEventHandler_StatusEvent(t *testing.T) {
	var mu sync.Mutex
	var receivedPayloads []WebhookPayload

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var p WebhookPayload
		_ = json.Unmarshal(body, &p)
		mu.Lock()
		receivedPayloads = append(receivedPayloads, p)
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	repo := &mockWebhookRepo{
		webhooks: []*entity.Webhook{
			{
				ID:      "status-wh",
				Name:    "Status Events",
				URL:     server.URL,
				Method:  "POST",
				Events:  `["question.status"]`,
				Enabled: true,
			},
		},
	}

	mockQueue := &mockEventQueueService{}
	handler := NewWebhookEventHandler(repo, NewWebhookService(), mockQueue)

	// Send a question.status event with status transition (available -> deleted)
	msg := schema.NewEvent(constant.EventQuestionStatus, "user1").
		TID("q1").QID("q1", "user1").
		StatusChange(
			entity.QuestionStatusAvailable,
			entity.QuestionStatusDeleted,
			entity.AdminQuestionSearchStatusIntToString,
		)
	err := handler.Handle(context.Background(), msg)
	require.NoError(t, err)

	handler.Wait()

	mu.Lock()
	require.Len(t, receivedPayloads, 1)
	p := receivedPayloads[0]
	assert.Equal(t, "question.status", p.Event)
	assert.Contains(t, p.Data, "from_status")
	assert.Contains(t, p.Data, "to_status")
	assert.NotContains(t, p.Data, "transition")

	fromStatus := p.Data["from_status"].(map[string]any)
	assert.Equal(t, "available", fromStatus["label"])

	toStatus := p.Data["to_status"].(map[string]any)
	assert.Equal(t, "deleted", toStatus["label"])
	mu.Unlock()
}

func TestWebhookEventHandler_RetryThenSuccess(t *testing.T) {
	var callCount atomic.Int32

	// Server returns 500 on the first 2 attempts, then 200.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := callCount.Add(1)
		if n <= 2 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	repo := &mockWebhookRepo{
		webhooks: []*entity.Webhook{
			{
				ID:      "retry-wh",
				Name:    "Retry Webhook",
				URL:     server.URL,
				Method:  "POST",
				Events:  `["question.create"]`,
				Enabled: true,
			},
		},
	}

	mockQueue := &mockEventQueueService{}
	handler := NewWebhookEventHandler(repo, NewWebhookService(), mockQueue)

	msg := schema.NewEvent(constant.EventQuestionCreate, "user1").
		TID("q1").QID("q1", "user1")
	err := handler.Handle(context.Background(), msg)
	require.NoError(t, err)

	handler.Wait()

	assert.Equal(t, int32(3), callCount.Load(), "should have been called exactly 3 times (2 failures + 1 success)")
}

func TestWebhookEventHandler_RetryExhausted(t *testing.T) {
	var callCount atomic.Int32

	// Server always returns 502.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount.Add(1)
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer server.Close()

	repo := &mockWebhookRepo{
		webhooks: []*entity.Webhook{
			{
				ID:      "exhaust-wh",
				Name:    "Exhaust Webhook",
				URL:     server.URL,
				Method:  "POST",
				Events:  `["question.create"]`,
				Enabled: true,
			},
		},
	}

	mockQueue := &mockEventQueueService{}
	handler := NewWebhookEventHandler(repo, NewWebhookService(), mockQueue)

	msg := schema.NewEvent(constant.EventQuestionCreate, "user1").
		TID("q1").QID("q1", "user1")
	err := handler.Handle(context.Background(), msg)
	require.NoError(t, err)

	handler.Wait()

	assert.Equal(t, int32(maxRetries), callCount.Load(), "should have been called exactly maxRetries times")
}
