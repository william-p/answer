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
	"sync"
	"time"

	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/entity"
	webhookrepo "github.com/apache/answer/internal/repo/webhook"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/eventqueue"
	"github.com/segmentfault/pacman/log"
)

const maxConcurrentSends = 10

// WebhookEventHandler handles events and dispatches them to configured webhooks.
// It limits concurrency to maxConcurrentSends goroutines and tracks in-flight
// sends with a WaitGroup so they can be drained on shutdown.
type WebhookEventHandler struct {
	webhookRepo    webhookrepo.WebhookRepo
	webhookService *WebhookService
	wg             sync.WaitGroup
	sem            chan struct{}
}

// NewWebhookEventHandler creates a new WebhookEventHandler and registers it on the event queue
func NewWebhookEventHandler(
	webhookRepo webhookrepo.WebhookRepo,
	webhookService *WebhookService,
	eventQueueService eventqueue.Service,
) *WebhookEventHandler {
	h := &WebhookEventHandler{
		webhookRepo:    webhookRepo,
		webhookService: webhookService,
		sem:            make(chan struct{}, maxConcurrentSends),
	}
	eventQueueService.RegisterHandler(h.Handle)
	return h
}

// Wait blocks until all in-flight webhook sends have completed.
func (h *WebhookEventHandler) Wait() {
	h.wg.Wait()
}

// Handle processes an event message and dispatches to matching webhooks
func (h *WebhookEventHandler) Handle(ctx context.Context, msg *schema.EventMsg) error {
	log.Debugf("[webhook] received event: type=%s, trigger_object_id=%s", msg.EventType, msg.TriggerObjectID)
	webhooks, err := h.webhookRepo.GetEnabledWebhooks(ctx)
	if err != nil {
		log.Errorf("[webhook] failed to get enabled webhooks: %v", err)
		return err
	}
	log.Debugf("[webhook] found %d enabled webhooks", len(webhooks))

	payload := h.buildPayload(msg)

	for _, webhook := range webhooks {
		log.Debugf("[webhook] checking webhook id=%s name=%s events=%v", webhook.ID, webhook.Name, webhook.Events)
		events := parseWebhookEvents(webhook)
		if !shouldSendEvent(events, msg.EventType) {
			log.Debugf("[webhook] skipping webhook id=%s (event type %s not in configured events)", webhook.ID, msg.EventType)
			continue
		}
		log.Infof("[webhook] dispatching event %s to webhook id=%s name=%s url=%s", msg.EventType, webhook.ID, webhook.Name, webhook.URL)
		// wg.Add must precede the goroutine launch to ensure Wait() observes it.
		// The semaphore (sem) limits at most maxConcurrentSends in-flight HTTP requests.
		h.wg.Add(1)
		wh := webhook
		go func() {
			defer h.wg.Done()
			delay := initialRetryDelay
			for attempt := range maxRetries {
				h.sem <- struct{}{}
				err := h.webhookService.SendOnce(wh, payload)
				<-h.sem
				if err == nil {
					log.Debugf("[webhook] successfully sent webhook id=%s", wh.ID)
					return
				}
				if attempt == maxRetries-1 {
					log.Errorf("[webhook] failed to send webhook id=%s after %d attempts: %v", wh.ID, maxRetries, err)
					return
				}
				log.Warnf("[webhook] webhook id=%s attempt %d/%d failed: %v, retrying in %v", wh.ID, attempt+1, maxRetries, err, delay)
				time.Sleep(delay)
				delay *= 2
			}
		}()
	}
	return nil
}

// parseWebhookEvents deserializes the JSON event list from a webhook entity into
// a set for O(1) lookup. Returns an empty map on parse failure.
func parseWebhookEvents(wh *entity.Webhook) map[string]struct{} {
	var list []string
	if err := json.Unmarshal([]byte(wh.Events), &list); err != nil {
		log.Errorf("[webhook] failed to parse events for webhook id=%s name=%s: %v", wh.ID, wh.Name, err)
		return nil
	}
	set := make(map[string]struct{}, len(list))
	for _, e := range list {
		set[e] = struct{}{}
	}
	return set
}

// shouldSendEvent checks if a pre-parsed event set contains the given event type
func shouldSendEvent(events map[string]struct{}, eventType constant.EventType) bool {
	_, ok := events[string(eventType)]
	return ok
}

// buildPayload constructs a WebhookPayload from an EventMsg
func (h *WebhookEventHandler) buildPayload(msg *schema.EventMsg) *WebhookPayload {
	data := map[string]any{
		"user_id": msg.UserID,
	}
	if msg.QuestionID != "" {
		data["question_id"] = msg.QuestionID
	}
	if msg.QuestionUserID != "" {
		data["question_user_id"] = msg.QuestionUserID
	}
	if msg.AnswerID != "" {
		data["answer_id"] = msg.AnswerID
	}
	if msg.AnswerUserID != "" {
		data["answer_user_id"] = msg.AnswerUserID
	}
	if msg.CommentID != "" {
		data["comment_id"] = msg.CommentID
	}
	if msg.TriggerObjectID != "" {
		data["trigger_object_id"] = msg.TriggerObjectID
	}
	if msg.FromStatusLabel != "" || msg.ToStatusLabel != "" {
		data["from_status"] = map[string]any{
			"code":  msg.FromStatus,
			"label": msg.FromStatusLabel,
		}
		data["to_status"] = map[string]any{
			"code":  msg.ToStatus,
			"label": msg.ToStatusLabel,
		}
	}

	return &WebhookPayload{
		Event:     string(msg.EventType),
		Timestamp: time.Now().Unix(),
		Data:      data,
	}
}
