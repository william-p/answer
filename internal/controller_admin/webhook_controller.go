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

package controller_admin

import (
	"github.com/apache/answer/internal/base/handler"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/webhook"
	"github.com/gin-gonic/gin"
	"github.com/segmentfault/pacman/errors"
)

type WebhookController struct {
	webhookAdminService *webhook.WebhookAdminService
}

func NewWebhookController(webhookAdminService *webhook.WebhookAdminService) *WebhookController {
	return &WebhookController{webhookAdminService: webhookAdminService}
}

// ListWebhooks list all webhooks
// @Summary list all webhooks
// @Description list all webhooks
// @Tags AdminWebhook
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} handler.RespBody{data=[]schema.GetWebhookResp}
// @Router /answer/admin/api/webhooks [get]
func (wc *WebhookController) ListWebhooks(ctx *gin.Context) {
	resp, err := wc.webhookAdminService.ListWebhooks(ctx)
	handler.HandleResponse(ctx, err, resp)
}

// GetWebhook get a webhook by id
// @Summary get a webhook by id
// @Description get a webhook by id
// @Tags AdminWebhook
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id query string true "webhook id"
// @Success 200 {object} handler.RespBody{data=schema.GetWebhookResp}
// @Router /answer/admin/api/webhook [get]
func (wc *WebhookController) GetWebhook(ctx *gin.Context) {
	id := ctx.Query("id")
	if id == "" {
		handler.HandleResponse(ctx, errors.BadRequest(reason.RequestFormatError), nil)
		return
	}
	resp, err := wc.webhookAdminService.GetWebhook(ctx, id)
	handler.HandleResponse(ctx, err, resp)
}

// AddWebhook add a webhook
// @Summary add a webhook
// @Description add a webhook
// @Tags AdminWebhook
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param data body schema.AddWebhookReq true "AddWebhookReq"
// @Success 200 {object} handler.RespBody
// @Router /answer/admin/api/webhook [post]
func (wc *WebhookController) AddWebhook(ctx *gin.Context) {
	req := &schema.AddWebhookReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	err := wc.webhookAdminService.AddWebhook(ctx, req)
	handler.HandleResponse(ctx, err, nil)
}

// UpdateWebhook update a webhook
// @Summary update a webhook
// @Description update a webhook
// @Tags AdminWebhook
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param data body schema.UpdateWebhookReq true "UpdateWebhookReq"
// @Success 200 {object} handler.RespBody
// @Router /answer/admin/api/webhook [put]
func (wc *WebhookController) UpdateWebhook(ctx *gin.Context) {
	req := &schema.UpdateWebhookReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	err := wc.webhookAdminService.UpdateWebhook(ctx, req)
	handler.HandleResponse(ctx, err, nil)
}

// DeleteWebhook delete a webhook
// @Summary delete a webhook
// @Description delete a webhook
// @Tags AdminWebhook
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param data body schema.DeleteWebhookReq true "DeleteWebhookReq"
// @Success 200 {object} handler.RespBody
// @Router /answer/admin/api/webhook [delete]
func (wc *WebhookController) DeleteWebhook(ctx *gin.Context) {
	req := &schema.DeleteWebhookReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	err := wc.webhookAdminService.DeleteWebhook(ctx, req)
	handler.HandleResponse(ctx, err, nil)
}
