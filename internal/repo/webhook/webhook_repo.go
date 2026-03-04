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

	"github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/entity"
	"github.com/segmentfault/pacman/errors"
)

// WebhookRepo webhook repository interface
type WebhookRepo interface {
	CreateWebhook(ctx context.Context, webhook *entity.Webhook) (err error)
	UpdateWebhook(ctx context.Context, webhook *entity.Webhook) (err error)
	DeleteWebhook(ctx context.Context, id string) (err error)
	GetWebhook(ctx context.Context, id string) (webhook *entity.Webhook, exist bool, err error)
	ListWebhooks(ctx context.Context) (webhooks []*entity.Webhook, err error)
	GetEnabledWebhooks(ctx context.Context) (webhooks []*entity.Webhook, err error)
}

type webhookRepo struct {
	data *data.Data
}

// NewWebhookRepo creates a new webhook repository
func NewWebhookRepo(data *data.Data) WebhookRepo {
	return &webhookRepo{data: data}
}

// CreateWebhook creates a new webhook
func (r *webhookRepo) CreateWebhook(ctx context.Context, webhook *entity.Webhook) (err error) {
	_, err = r.data.DB.Context(ctx).Insert(webhook)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return nil
}

// UpdateWebhook updates an existing webhook
func (r *webhookRepo) UpdateWebhook(ctx context.Context, webhook *entity.Webhook) (err error) {
	_, err = r.data.DB.Context(ctx).ID(webhook.ID).AllCols().Update(webhook)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return nil
}

// DeleteWebhook deletes a webhook by id
func (r *webhookRepo) DeleteWebhook(ctx context.Context, id string) (err error) {
	_, err = r.data.DB.Context(ctx).ID(id).Delete(&entity.Webhook{})
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return nil
}

// GetWebhook gets a webhook by id
func (r *webhookRepo) GetWebhook(ctx context.Context, id string) (webhook *entity.Webhook, exist bool, err error) {
	webhook = &entity.Webhook{}
	exist, err = r.data.DB.Context(ctx).ID(id).Get(webhook)
	if err != nil {
		return nil, false, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return webhook, exist, nil
}

// ListWebhooks lists all webhooks
func (r *webhookRepo) ListWebhooks(ctx context.Context) (webhooks []*entity.Webhook, err error) {
	webhooks = make([]*entity.Webhook, 0)
	err = r.data.DB.Context(ctx).OrderBy("created_at DESC").Find(&webhooks)
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return webhooks, nil
}

// GetEnabledWebhooks gets all enabled webhooks
func (r *webhookRepo) GetEnabledWebhooks(ctx context.Context) (webhooks []*entity.Webhook, err error) {
	webhooks = make([]*entity.Webhook, 0)
	err = r.data.DB.Context(ctx).Where("enabled = ?", true).Find(&webhooks)
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return webhooks, nil
}
