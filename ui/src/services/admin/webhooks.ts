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

import useSWR from 'swr';

import request from '@/utils/request';

export interface WebhookItem {
  id: string;
  name: string;
  url: string;
  method: string;
  header_name: string;
  header_value: string;
  events: string[];
  enabled: boolean;
  created_at: number;
  updated_at: number;
}

export interface AddOrEditWebhookParams {
  id?: string;
  name: string;
  url: string;
  method: string;
  header_name?: string;
  header_value?: string;
  events: string[];
  enabled: boolean;
}

export const useQueryWebhooks = () => {
  const apiUrl = `/answer/admin/api/webhooks`;
  const { data, error, mutate } = useSWR<WebhookItem[], Error>(
    apiUrl,
    request.instance.get,
  );
  return {
    data,
    isLoading: !data && !error,
    error,
    mutate,
  };
};

export const addWebhook = (params: AddOrEditWebhookParams) => {
  return request.post('/answer/admin/api/webhook', params);
};

export const updateWebhook = (params: AddOrEditWebhookParams) => {
  return request.put('/answer/admin/api/webhook', params);
};

export const deleteWebhook = (id: string) => {
  return request.delete('/answer/admin/api/webhook', { id });
};
