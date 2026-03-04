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

import (
	"fmt"
	"testing"

	"github.com/apache/answer/internal/base/validator"
	"github.com/segmentfault/pacman/i18n"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWebhookSchemaValidation tests that webhook schemas can be validated
// without causing nil pointer panics in the i18n validator system.
// This is a regression test for the bug where using 'url' and 'dive' validate tags
// caused panics because they weren't registered in the custom i18n validator.
func TestWebhookSchemaValidation(t *testing.T) {
	// Get the default validator (uses default language)
	v := validator.GetValidatorByLang(i18n.LanguageEnglish)
	require.NotNil(t, v, "validator should not be nil")

	tests := []struct {
		name      string
		req       any
		wantError bool
	}{
		{
			name: "valid AddWebhookReq",
			req: &AddWebhookReq{
				Name:   "Test Webhook",
				URL:    "https://example.com/webhook",
				Method: "POST",
				Events: []string{"question.create"},
			},
			wantError: false,
		},
		{
			name: "AddWebhookReq with empty name",
			req: &AddWebhookReq{
				Name:   "",
				URL:    "https://example.com/webhook",
				Method: "POST",
				Events: []string{"question.create"},
			},
			wantError: true,
		},
		{
			name: "AddWebhookReq with empty URL",
			req: &AddWebhookReq{
				Name:   "Test",
				URL:    "",
				Method: "POST",
				Events: []string{"question.create"},
			},
			wantError: true,
		},
		{
			name: "AddWebhookReq with invalid method",
			req: &AddWebhookReq{
				Name:   "Test",
				URL:    "https://example.com/webhook",
				Method: "GET",
				Events: []string{"question.create"},
			},
			wantError: true,
		},
		{
			name: "AddWebhookReq with empty events",
			req: &AddWebhookReq{
				Name:   "Test",
				URL:    "https://example.com/webhook",
				Method: "POST",
				Events: []string{},
			},
			wantError: true,
		},
		{
			name: "valid UpdateWebhookReq",
			req: &UpdateWebhookReq{
				ID:     "123",
				Name:   "Updated Webhook",
				URL:    "https://example.com/webhook",
				Method: "PUT",
				Events: []string{"question.update", "question.delete"},
			},
			wantError: false,
		},
		{
			name: "UpdateWebhookReq with missing ID",
			req: &UpdateWebhookReq{
				ID:     "",
				Name:   "Test",
				URL:    "https://example.com/webhook",
				Method: "POST",
				Events: []string{"question.create"},
			},
			wantError: true,
		},
		{
			name: "valid DeleteWebhookReq",
			req: &DeleteWebhookReq{
				ID: "123",
			},
			wantError: false,
		},
		{
			name: "DeleteWebhookReq with missing ID",
			req: &DeleteWebhookReq{
				ID: "",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This should not panic - that's the main assertion
			errFields, err := v.Check(tt.req)

			if tt.wantError {
				// We expect validation errors
				assert.True(t, len(errFields) > 0 || err != nil,
					"expected validation errors but got none")
			} else {
				// We expect no validation errors
				assert.Empty(t, errFields, "expected no validation errors")
				assert.NoError(t, err, "expected no error")
			}
		})
	}
}

// TestWebhookSchemaValidation_NoPanic specifically tests that validation
// doesn't panic even with edge cases
func TestWebhookSchemaValidation_NoPanic(t *testing.T) {
	v := validator.GetValidatorByLang(i18n.LanguageEnglish)
	require.NotNil(t, v)

	// Test with various edge cases that might trigger panics
	edgeCases := []any{
		&AddWebhookReq{
			Name:        "Test",
			URL:         "not-a-url", // Invalid URL format but shouldn't panic
			Method:      "POST",
			Events:      []string{"question.create"},
			HeaderName:  "X-Token",
			HeaderValue: "secret",
		},
		&AddWebhookReq{
			Name:   "Test",
			URL:    "https://example.com/webhook",
			Method: "POST",
			Events: []string{"invalid.event"}, // Invalid event but shouldn't panic
		},
		&UpdateWebhookReq{
			ID:     "123",
			Name:   "Test",
			URL:    "https://example.com/webhook",
			Method: "POST",
			Events: []string{"question.create", "question.update", "question.delete", "question.accept"},
		},
	}

	for i, req := range edgeCases {
		t.Run(fmt.Sprintf("edge_case_%d", i), func(t *testing.T) {
			// The main assertion is that this doesn't panic
			assert.NotPanics(t, func() {
				_, _ = v.Check(req)
			}, "validation should not panic for edge case %d", i)
		})
	}
}
