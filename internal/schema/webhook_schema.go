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

package schema

// AddWebhookReq add webhook request
type AddWebhookReq struct {
	Name        string   `json:"name" validate:"required,gt=0,lte=100"`
	URL         string   `json:"url" validate:"required,gt=0,lte=500"`
	Method      string   `json:"method" validate:"required,oneof=POST PUT"`
	HeaderName  string   `json:"header_name" validate:"omitempty,lte=100"`
	HeaderValue string   `json:"header_value" validate:"omitempty,lte=200"`
	Events      []string `json:"events" validate:"required,min=1"`
	Enabled     bool     `json:"enabled"`
}

// UpdateWebhookReq update webhook request
type UpdateWebhookReq struct {
	ID          string   `json:"id" validate:"required"`
	Name        string   `json:"name" validate:"required,gt=0,lte=100"`
	URL         string   `json:"url" validate:"required,gt=0,lte=500"`
	Method      string   `json:"method" validate:"required,oneof=POST PUT"`
	HeaderName  string   `json:"header_name" validate:"omitempty,lte=100"`
	HeaderValue string   `json:"header_value" validate:"omitempty,lte=200"`
	Events      []string `json:"events" validate:"required,min=1"`
	Enabled     bool     `json:"enabled"`
}

// DeleteWebhookReq delete webhook request
type DeleteWebhookReq struct {
	ID string `json:"id" validate:"required"`
}

// GetWebhookResp get webhook response
type GetWebhookResp struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	URL         string   `json:"url"`
	Method      string   `json:"method"`
	HeaderName  string   `json:"header_name"`
	HeaderValue string   `json:"header_value"`
	Events      []string `json:"events"`
	Enabled     bool     `json:"enabled"`
	CreatedAt   int64    `json:"created_at"`
	UpdatedAt   int64    `json:"updated_at"`
}
