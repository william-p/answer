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
	"strings"

	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/entity"
	webhookrepo "github.com/apache/answer/internal/repo/webhook"
	"github.com/apache/answer/internal/schema"
	"github.com/segmentfault/pacman/errors"
	"github.com/segmentfault/pacman/log"
)

var allowedEvents = map[string]bool{
	"question.create":  true,
	"question.update":  true,
	"question.status":  true,
	"question.delete":  true,
	"question.hide":    true,
	"question.show":    true,
	"question.pin":     true,
	"question.unpin":   true,
	"answer.create":    true,
	"answer.update":    true,
	"answer.status":    true,
	"answer.delete":    true,
	"answer.accept":    true,
	"answer.unaccept":  true,
}

// WebhookAdminService webhook admin service
type WebhookAdminService struct {
	webhookRepo webhookrepo.WebhookRepo
}

// NewWebhookAdminService new webhook admin service
func NewWebhookAdminService(webhookRepo webhookrepo.WebhookRepo) *WebhookAdminService {
	return &WebhookAdminService{webhookRepo: webhookRepo}
}

// ListWebhooks list all webhooks
func (s *WebhookAdminService) ListWebhooks(ctx context.Context) (resp []*schema.GetWebhookResp, err error) {
	webhooks, err := s.webhookRepo.ListWebhooks(ctx)
	if err != nil {
		return nil, err
	}
	resp = make([]*schema.GetWebhookResp, 0, len(webhooks))
	for _, w := range webhooks {
		resp = append(resp, convertWebhookToResp(w))
	}
	return resp, nil
}

// GetWebhook get a webhook by id
func (s *WebhookAdminService) GetWebhook(ctx context.Context, id string) (resp *schema.GetWebhookResp, err error) {
	webhook, exist, err := s.webhookRepo.GetWebhook(ctx, id)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, errors.NotFound(reason.ObjectNotFound)
	}
	return convertWebhookToResp(webhook), nil
}

// AddWebhook add a webhook
func (s *WebhookAdminService) AddWebhook(ctx context.Context, req *schema.AddWebhookReq) (err error) {
	if err := validateEvents(req.Events); err != nil {
		return err
	}
	eventsJSON, err := json.Marshal(req.Events)
	if err != nil {
		return errors.InternalServer(reason.UnknownError).WithError(err).WithStack()
	}
	webhook := &entity.Webhook{
		Name:        req.Name,
		URL:         req.URL,
		Method:      req.Method,
		HeaderName:  req.HeaderName,
		HeaderValue: req.HeaderValue,
		Events:      string(eventsJSON),
		Enabled:     req.Enabled,
	}
	return s.webhookRepo.CreateWebhook(ctx, webhook)
}

// UpdateWebhook update a webhook
func (s *WebhookAdminService) UpdateWebhook(ctx context.Context, req *schema.UpdateWebhookReq) (err error) {
	_, exist, err := s.webhookRepo.GetWebhook(ctx, req.ID)
	if err != nil {
		return err
	}
	if !exist {
		return errors.NotFound(reason.ObjectNotFound)
	}
	if err := validateEvents(req.Events); err != nil {
		return err
	}
	eventsJSON, err := json.Marshal(req.Events)
	if err != nil {
		return errors.InternalServer(reason.UnknownError).WithError(err).WithStack()
	}
	webhook := &entity.Webhook{
		ID:          req.ID,
		Name:        req.Name,
		URL:         req.URL,
		Method:      req.Method,
		HeaderName:  req.HeaderName,
		HeaderValue: req.HeaderValue,
		Events:      string(eventsJSON),
		Enabled:     req.Enabled,
	}
	return s.webhookRepo.UpdateWebhook(ctx, webhook)
}

// DeleteWebhook delete a webhook
func (s *WebhookAdminService) DeleteWebhook(ctx context.Context, req *schema.DeleteWebhookReq) (err error) {
	_, exist, err := s.webhookRepo.GetWebhook(ctx, req.ID)
	if err != nil {
		return err
	}
	if !exist {
		return errors.NotFound(reason.ObjectNotFound)
	}
	return s.webhookRepo.DeleteWebhook(ctx, req.ID)
}

func convertWebhookToResp(w *entity.Webhook) *schema.GetWebhookResp {
	var events []string
	if err := json.Unmarshal([]byte(w.Events), &events); err != nil {
		log.Warnf("[webhook] failed to parse events JSON for webhook id=%s: %v", w.ID, err)
	}
	maskedValue := maskHeaderValue(w.HeaderValue)
	return &schema.GetWebhookResp{
		ID:          w.ID,
		Name:        w.Name,
		URL:         w.URL,
		Method:      w.Method,
		HeaderName:  w.HeaderName,
		HeaderValue: maskedValue,
		Events:      events,
		Enabled:     w.Enabled,
		CreatedAt:   w.CreatedAt.Unix(),
		UpdatedAt:   w.UpdatedAt.Unix(),
	}
}

// maskHeaderValue masks a header value for safe display, showing only the first 2
// and last 2 characters. Values of 4 characters or less are fully masked.
func maskHeaderValue(val string) string {
	if len(val) <= 4 {
		return strings.Repeat("*", len(val))
	}
	return val[:2] + strings.Repeat("*", len(val)-4) + val[len(val)-2:]
}

func validateEvents(events []string) error {
	for _, e := range events {
		if !allowedEvents[e] {
			return errors.BadRequest(reason.WebhookInvalidEventType)
		}
	}
	return nil
}
