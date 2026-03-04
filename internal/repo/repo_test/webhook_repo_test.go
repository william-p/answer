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

package repo_test

import (
	"context"
	"testing"

	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/repo/webhook"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_webhookRepo_CRUD(t *testing.T) {
	webhookRepo := webhook.NewWebhookRepo(testDataSource)
	ctx := context.TODO()

	// Create
	w := &entity.Webhook{
		Name:        "Test Webhook",
		URL:         "https://example.com/webhook",
		Method:      "POST",
		HeaderName:  "X-Webhook-Token",
		HeaderValue: "secret123",
		Events:      `["question.create","question.update"]`,
		Enabled:     true,
	}
	err := webhookRepo.CreateWebhook(ctx, w)
	require.NoError(t, err)
	assert.NotEmpty(t, w.ID)

	// Get
	got, exist, err := webhookRepo.GetWebhook(ctx, w.ID)
	require.NoError(t, err)
	assert.True(t, exist)
	assert.Equal(t, w.Name, got.Name)
	assert.Equal(t, w.URL, got.URL)
	assert.Equal(t, w.Method, got.Method)
	assert.Equal(t, w.HeaderName, got.HeaderName)
	assert.Equal(t, w.HeaderValue, got.HeaderValue)
	assert.Equal(t, w.Events, got.Events)
	assert.True(t, got.Enabled)

	// Update
	got.Name = "Updated Webhook"
	got.URL = "https://example.com/webhook2"
	got.Method = "PUT"
	err = webhookRepo.UpdateWebhook(ctx, got)
	require.NoError(t, err)

	updated, exist, err := webhookRepo.GetWebhook(ctx, w.ID)
	require.NoError(t, err)
	assert.True(t, exist)
	assert.Equal(t, "Updated Webhook", updated.Name)
	assert.Equal(t, "https://example.com/webhook2", updated.URL)
	assert.Equal(t, "PUT", updated.Method)

	// List
	list, err := webhookRepo.ListWebhooks(ctx)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(list), 1)

	// Delete
	err = webhookRepo.DeleteWebhook(ctx, w.ID)
	require.NoError(t, err)

	_, exist, err = webhookRepo.GetWebhook(ctx, w.ID)
	require.NoError(t, err)
	assert.False(t, exist)
}

func Test_webhookRepo_GetEnabledWebhooks(t *testing.T) {
	webhookRepo := webhook.NewWebhookRepo(testDataSource)
	ctx := context.TODO()

	// Create an enabled webhook
	enabled := &entity.Webhook{
		Name:    "Enabled",
		URL:     "https://example.com/enabled",
		Method:  "POST",
		Events:  `["question.create"]`,
		Enabled: true,
	}
	err := webhookRepo.CreateWebhook(ctx, enabled)
	require.NoError(t, err)

	// Create a disabled webhook
	disabled := &entity.Webhook{
		Name:    "Disabled",
		URL:     "https://example.com/disabled",
		Method:  "POST",
		Events:  `["question.create"]`,
		Enabled: false,
	}
	err = webhookRepo.CreateWebhook(ctx, disabled)
	require.NoError(t, err)

	// GetEnabledWebhooks should only return enabled ones
	webhooks, err := webhookRepo.GetEnabledWebhooks(ctx)
	require.NoError(t, err)
	for _, wh := range webhooks {
		assert.True(t, wh.Enabled, "GetEnabledWebhooks should only return enabled webhooks")
	}

	// Cleanup
	_ = webhookRepo.DeleteWebhook(ctx, enabled.ID)
	_ = webhookRepo.DeleteWebhook(ctx, disabled.ID)
}
