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
	"testing"

	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// adminMockWebhookRepo implements WebhookRepo for admin service tests
type adminMockWebhookRepo struct {
	webhooks map[string]*entity.Webhook
	updated  map[string]*entity.Webhook
	deleted  []string
}

func newAdminMockRepo() *adminMockWebhookRepo {
	return &adminMockWebhookRepo{
		webhooks: make(map[string]*entity.Webhook),
		updated:  make(map[string]*entity.Webhook),
	}
}

func (m *adminMockWebhookRepo) CreateWebhook(ctx context.Context, webhook *entity.Webhook) error {
	m.webhooks[webhook.ID] = webhook
	return nil
}
func (m *adminMockWebhookRepo) UpdateWebhook(ctx context.Context, webhook *entity.Webhook) error {
	m.updated[webhook.ID] = webhook
	return nil
}
func (m *adminMockWebhookRepo) DeleteWebhook(ctx context.Context, id string) error {
	m.deleted = append(m.deleted, id)
	return nil
}
func (m *adminMockWebhookRepo) GetWebhook(ctx context.Context, id string) (*entity.Webhook, bool, error) {
	w, ok := m.webhooks[id]
	if !ok {
		return nil, false, nil
	}
	return w, true, nil
}
func (m *adminMockWebhookRepo) ListWebhooks(ctx context.Context) ([]*entity.Webhook, error) {
	list := make([]*entity.Webhook, 0, len(m.webhooks))
	for _, w := range m.webhooks {
		list = append(list, w)
	}
	return list, nil
}
func (m *adminMockWebhookRepo) GetEnabledWebhooks(ctx context.Context) ([]*entity.Webhook, error) {
	return nil, nil
}

func TestWebhookAdminService_UpdateWebhook_NotFound(t *testing.T) {
	repo := newAdminMockRepo()
	svc := NewWebhookAdminService(repo)

	err := svc.UpdateWebhook(context.Background(), &schema.UpdateWebhookReq{
		ID:     "nonexistent",
		Name:   "Test",
		URL:    "https://example.com",
		Method: "POST",
		Events: []string{"question.create"},
	})
	require.Error(t, err, "should return error when webhook does not exist")
	assert.Empty(t, repo.updated, "repo.UpdateWebhook should not have been called")
}

func TestWebhookAdminService_UpdateWebhook_Exists(t *testing.T) {
	repo := newAdminMockRepo()
	repo.webhooks["1"] = &entity.Webhook{
		ID: "1", Name: "Old", URL: "https://old.com", Method: "POST",
		Events: `["question.create"]`, Enabled: true,
	}
	svc := NewWebhookAdminService(repo)

	err := svc.UpdateWebhook(context.Background(), &schema.UpdateWebhookReq{
		ID:     "1",
		Name:   "Updated",
		URL:    "https://new.com",
		Method: "PUT",
		Events: []string{"question.update"},
	})
	require.NoError(t, err)
	assert.Contains(t, repo.updated, "1", "repo.UpdateWebhook should have been called")
}

func TestWebhookAdminService_DeleteWebhook_NotFound(t *testing.T) {
	repo := newAdminMockRepo()
	svc := NewWebhookAdminService(repo)

	err := svc.DeleteWebhook(context.Background(), &schema.DeleteWebhookReq{
		ID: "nonexistent",
	})
	require.Error(t, err, "should return error when webhook does not exist")
	assert.Empty(t, repo.deleted, "repo.DeleteWebhook should not have been called")
}

func TestWebhookAdminService_DeleteWebhook_Exists(t *testing.T) {
	repo := newAdminMockRepo()
	repo.webhooks["1"] = &entity.Webhook{ID: "1", Name: "Test"}
	svc := NewWebhookAdminService(repo)

	err := svc.DeleteWebhook(context.Background(), &schema.DeleteWebhookReq{
		ID: "1",
	})
	require.NoError(t, err)
	assert.Contains(t, repo.deleted, "1")
}

func TestWebhookAdminService_GetWebhook_NotFound(t *testing.T) {
	repo := newAdminMockRepo()
	svc := NewWebhookAdminService(repo)

	resp, err := svc.GetWebhook(context.Background(), "nonexistent")
	require.Error(t, err, "should return error when webhook does not exist")
	assert.Nil(t, resp)
}

func TestWebhookAdminService_GetWebhook_Exists(t *testing.T) {
	repo := newAdminMockRepo()
	repo.webhooks["1"] = &entity.Webhook{
		ID: "1", Name: "Test", URL: "https://example.com", Method: "POST",
		Events: `["question.create"]`, Enabled: true,
	}
	svc := NewWebhookAdminService(repo)

	resp, err := svc.GetWebhook(context.Background(), "1")
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "1", resp.ID)
	assert.Equal(t, "Test", resp.Name)
}
