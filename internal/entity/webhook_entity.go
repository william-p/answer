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

package entity

import "time"

// Webhook webhook configuration
type Webhook struct {
	ID          string    `xorm:"not null pk autoincr BIGINT(20) id"`
	Name        string    `xorm:"not null VARCHAR(100) name"`
	URL         string    `xorm:"not null VARCHAR(500) url"`
	Method      string    `xorm:"not null default 'POST' VARCHAR(10) method"`
	HeaderName  string    `xorm:"VARCHAR(100) header_name"`
	HeaderValue string    `xorm:"VARCHAR(200) header_value"`
	Events      string    `xorm:"not null TEXT events"`
	Enabled     bool      `xorm:"not null default 1 TINYINT(1) enabled"`
	CreatedAt   time.Time `xorm:"created not null TIMESTAMP created_at"`
	UpdatedAt   time.Time `xorm:"updated not null TIMESTAMP updated_at"`
}

// TableName webhook table name
func (w *Webhook) TableName() string {
	return "webhook"
}
