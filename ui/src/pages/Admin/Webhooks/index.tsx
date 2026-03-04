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

import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Button, Table, Badge, Form } from 'react-bootstrap';

import dayjs from 'dayjs';

import { Empty } from '@/components';
import { useQueryWebhooks, deleteWebhook, updateWebhook } from '@/services';
import type { WebhookItem } from '@/services/admin/webhooks';
import { useToast } from '@/hooks';

import Action from './components/Action';
import AddOrEditModal from './components/AddOrEditModal';

const Webhooks = () => {
  const { t } = useTranslation('translation', {
    keyPrefix: 'admin.webhooks',
  });
  const Toast = useToast();

  const [showModal, setShowModal] = useState<{
    visible: boolean;
    item: WebhookItem | null;
  }>({
    visible: false,
    item: null,
  });

  const {
    data: webhookList,
    isLoading,
    mutate: refreshList,
  } = useQueryWebhooks();

  const handleAddModalState = (visible: boolean, item: WebhookItem | null) => {
    setShowModal({ visible, item });
  };

  const handleCallback = () => {
    handleAddModalState(false, null);
    refreshList();
    Toast.onShow({
      msg: t('update', { keyPrefix: 'toast' }),
      variant: 'success',
    });
  };

  const handleDelete = (webhook: WebhookItem) => {
    if (window.confirm(t('delete_confirm'))) {
      deleteWebhook(webhook.id).then(() => {
        refreshList();
        Toast.onShow({
          msg: t('webhook_deleted', { keyPrefix: 'messages' }),
          variant: 'success',
        });
      });
    }
  };

  const handleToggleEnabled = (webhook: WebhookItem) => {
    updateWebhook({
      ...webhook,
      enabled: !webhook.enabled,
    }).then(() => {
      refreshList();
    });
  };

  return (
    <div>
      <h3 className="mb-4">{t('title')}</h3>
      <p className="text-muted mb-3">{t('desc')}</p>
      <Button
        variant="outline-primary mb-3"
        size="sm"
        onClick={() => handleAddModalState(true, null)}>
        {t('add_webhook')}
      </Button>
      <Table responsive="md">
        <thead>
          <tr>
            <th style={{ width: '20%' }}>{t('name')}</th>
            <th style={{ width: '25%' }}>{t('url')}</th>
            <th style={{ width: '8%' }}>{t('method')}</th>
            <th style={{ width: '20%' }}>{t('events')}</th>
            <th style={{ width: '10%' }}>{t('status')}</th>
            <th style={{ width: '12%' }}>{t('created')}</th>
            <th className="text-end" style={{ width: '5%' }}>
              {t('action')}
            </th>
          </tr>
        </thead>
        <tbody className="align-middle">
          {webhookList?.map((webhook) => (
            <tr key={webhook.id}>
              <td>{webhook.name}</td>
              <td className="text-break small">{webhook.url}</td>
              <td>
                <Badge bg="info" text="dark">
                  {webhook.method}
                </Badge>
              </td>
              <td>
                {(() => {
                  const grouped: Record<string, string[]> = {};
                  webhook.events?.forEach((event) => {
                    const [category, action] = event.split('.');
                    if (!grouped[category]) {
                      grouped[category] = [];
                    }
                    grouped[category].push(action);
                  });
                  return Object.entries(grouped).map(([category, actions]) => (
                    <div key={category} className="mb-2">
                      <div className="text-muted text-capitalize fw-semibold small mb-1">
                        {category}
                      </div>
                      <div>
                        {actions.map((action) => (
                          <Badge
                            key={action}
                            bg="info"
                            className="me-1 mb-1"
                            text="dark">
                            {action}
                          </Badge>
                        ))}
                      </div>
                    </div>
                  ));
                })()}
              </td>
              <td>
                <Form.Check
                  type="switch"
                  checked={webhook.enabled}
                  onChange={() => handleToggleEnabled(webhook)}
                />
              </td>
              <td className="small">
                {dayjs
                  .unix(webhook.created_at)
                  .tz()
                  .format(t('long_date_with_time', { keyPrefix: 'dates' }))}
              </td>
              <td className="text-end">
                <Action
                  showModal={() => handleAddModalState(true, webhook)}
                  onDelete={() => handleDelete(webhook)}
                />
              </td>
            </tr>
          ))}
        </tbody>
      </Table>
      {Number(webhookList?.length) <= 0 && !isLoading && <Empty />}

      <AddOrEditModal
        data={showModal.item}
        visible={showModal.visible}
        onClose={handleAddModalState}
        callback={handleCallback}
      />
    </div>
  );
};

export default Webhooks;
