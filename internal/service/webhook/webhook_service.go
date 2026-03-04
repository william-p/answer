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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/apache/answer/internal/entity"
	"github.com/segmentfault/pacman/log"
)

const (
	defaultTimeout    = 5 * time.Second
	maxRetries        = 3
	initialRetryDelay = 1 * time.Second
)

// WebhookPayload represents the JSON payload sent to webhook endpoints
type WebhookPayload struct {
	Event     string         `json:"event"`
	Timestamp int64          `json:"timestamp"`
	Data      map[string]any `json:"data"`
}

// WebhookService handles sending HTTP requests to webhook endpoints
type WebhookService struct {
	client *http.Client
}

// NewWebhookService creates a new WebhookService
func NewWebhookService() *WebhookService {
	return &WebhookService{
		client: &http.Client{
			Timeout: defaultTimeout,
		},
	}
}

// SendOnce performs a single HTTP request to the webhook endpoint.
// It returns nil on success (2xx status) or an error on failure.
// Callers are responsible for retry logic.
func (s *WebhookService) SendOnce(webhook *entity.Webhook, payload *WebhookPayload) error {
	log.Debugf("[webhook] preparing to send webhook id=%s to %s", webhook.ID, webhook.URL)
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}
	log.Debugf("[webhook] payload: %s", string(body))

	req, err := http.NewRequest(webhook.Method, webhook.URL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create webhook request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	if webhook.HeaderName != "" && webhook.HeaderValue != "" {
		req.Header.Set(webhook.HeaderName, webhook.HeaderValue)
	}

	log.Debugf("[webhook] sending HTTP %s request to %s", webhook.Method, webhook.URL)
	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("webhook request failed: %w", err)
	}
	statusCode := resp.StatusCode
	// Drain and close body immediately to reuse the connection.
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
	log.Debugf("[webhook] received HTTP %d response from %s", statusCode, webhook.URL)

	if statusCode >= 200 && statusCode < 300 {
		log.Infof("[webhook] webhook id=%s sent successfully (status %d)", webhook.ID, statusCode)
		return nil
	}
	return fmt.Errorf("webhook returned status %d", statusCode)
}
