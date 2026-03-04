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

import { useState, useEffect, FC } from 'react';
import { Modal, Button, Form } from 'react-bootstrap';
import { useTranslation } from 'react-i18next';

import type {
  WebhookItem,
  AddOrEditWebhookParams,
} from '@/services/admin/webhooks';
import { addWebhook, updateWebhook } from '@/services';

const AVAILABLE_EVENTS = {
  question: [
    'create',
    'update',
    'status',
    'delete',
    'hide',
    'show',
    'pin',
    'unpin',
  ],
  answer: ['create', 'update', 'status', 'delete', 'accept', 'unaccept'],
};

interface Props {
  visible: boolean;
  data: WebhookItem | null;
  onClose: (visible: boolean, item: WebhookItem | null) => void;
  callback: () => void;
}

const AddOrEditModal: FC<Props> = ({ visible, data, onClose, callback }) => {
  const { t } = useTranslation('translation', {
    keyPrefix: 'admin.webhooks',
  });
  const isEdit = !!data;

  const [formData, setFormData] = useState<AddOrEditWebhookParams>({
    name: '',
    url: '',
    method: 'POST',
    header_name: '',
    header_value: '',
    events: [],
    enabled: true,
  });

  useEffect(() => {
    if (data) {
      setFormData({
        id: data.id,
        name: data.name,
        url: data.url,
        method: data.method,
        header_name: data.header_name || '',
        header_value: data.header_value || '',
        events: data.events || [],
        enabled: data.enabled,
      });
    } else {
      setFormData({
        name: '',
        url: '',
        method: 'POST',
        header_name: '',
        header_value: '',
        events: [],
        enabled: true,
      });
    }
  }, [data, visible]);

  const handleEventToggle = (event: string) => {
    setFormData((prev) => ({
      ...prev,
      events: prev.events.includes(event)
        ? prev.events.filter((e) => e !== event)
        : [...prev.events, event],
    }));
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    try {
      if (isEdit) {
        await updateWebhook(formData);
      } else {
        await addWebhook(formData);
      }
      callback();
    } catch {
      // error handled by request util
    }
  };

  return (
    <Modal show={visible} onHide={() => onClose(false, null)} size="lg">
      <Modal.Header closeButton>
        <Modal.Title>
          {isEdit
            ? t('add_or_edit_modal.edit_title')
            : t('add_or_edit_modal.add_title')}
        </Modal.Title>
      </Modal.Header>
      <Form onSubmit={handleSubmit}>
        <Modal.Body>
          <Form.Group className="mb-3">
            <Form.Label>{t('add_or_edit_modal.name')}</Form.Label>
            <Form.Control
              type="text"
              required
              value={formData.name}
              onChange={(e) =>
                setFormData({ ...formData, name: e.target.value })
              }
              placeholder={t('add_or_edit_modal.name_placeholder')}
            />
          </Form.Group>

          <Form.Group className="mb-3">
            <Form.Label>{t('add_or_edit_modal.url')}</Form.Label>
            <Form.Control
              type="url"
              required
              value={formData.url}
              onChange={(e) =>
                setFormData({ ...formData, url: e.target.value })
              }
              placeholder="https://example.com/webhook"
            />
          </Form.Group>

          <Form.Group className="mb-3">
            <Form.Label>{t('add_or_edit_modal.method')}</Form.Label>
            <Form.Select
              value={formData.method}
              onChange={(e) =>
                setFormData({ ...formData, method: e.target.value })
              }>
              <option value="POST">POST</option>
              <option value="PUT">PUT</option>
            </Form.Select>
          </Form.Group>

          <Form.Group className="mb-3">
            <Form.Label>{t('add_or_edit_modal.header_name')}</Form.Label>
            <Form.Control
              type="text"
              value={formData.header_name}
              onChange={(e) =>
                setFormData({ ...formData, header_name: e.target.value })
              }
              placeholder="X-Webhook-Token"
            />
            <Form.Text className="text-muted">
              {t('add_or_edit_modal.header_hint')}
            </Form.Text>
          </Form.Group>

          <Form.Group className="mb-3">
            <Form.Label>{t('add_or_edit_modal.header_value')}</Form.Label>
            <Form.Control
              type="text"
              value={formData.header_value}
              onChange={(e) =>
                setFormData({ ...formData, header_value: e.target.value })
              }
              placeholder={t('add_or_edit_modal.header_value_placeholder')}
            />
          </Form.Group>

          <Form.Group className="mb-3">
            <Form.Label>{t('add_or_edit_modal.events')}</Form.Label>
            {Object.entries(AVAILABLE_EVENTS).map(([category, actions]) => (
              <div key={category} className="mb-3">
                <div className="text-muted text-capitalize fw-semibold small mb-2">
                  {category}
                </div>
                <div className="ms-3">
                  {actions.map((action) => {
                    const eventName = `${category}.${action}`;
                    return (
                      <Form.Check
                        key={eventName}
                        type="checkbox"
                        label={action}
                        checked={formData.events.includes(eventName)}
                        onChange={() => handleEventToggle(eventName)}
                      />
                    );
                  })}
                </div>
              </div>
            ))}
          </Form.Group>

          <Form.Group className="mb-3">
            <Form.Check
              type="switch"
              label={t('add_or_edit_modal.enabled')}
              checked={formData.enabled}
              onChange={(e) =>
                setFormData({ ...formData, enabled: e.target.checked })
              }
            />
          </Form.Group>
        </Modal.Body>
        <Modal.Footer>
          <Button variant="secondary" onClick={() => onClose(false, null)}>
            {t('add_or_edit_modal.cancel')}
          </Button>
          <Button variant="primary" type="submit">
            {isEdit
              ? t('add_or_edit_modal.save')
              : t('add_or_edit_modal.create')}
          </Button>
        </Modal.Footer>
      </Form>
    </Modal>
  );
};

export default AddOrEditModal;
