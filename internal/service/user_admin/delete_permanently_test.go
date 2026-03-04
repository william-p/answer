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

package user_admin

import (
	"context"
	"testing"

	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
	answercommon "github.com/apache/answer/internal/service/answer_common"
	questioncommon "github.com/apache/answer/internal/service/question_common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- mock: QuestionRepo (embed interface, override only DeletePermanentlyQuestions) ---

type mockQuestionRepo struct {
	questioncommon.QuestionRepo
	deletedQuestions []*entity.Question
	err              error
}

func (m *mockQuestionRepo) DeletePermanentlyQuestions(ctx context.Context) ([]*entity.Question, error) {
	return m.deletedQuestions, m.err
}

// --- mock: AnswerRepo (embed interface, override only DeletePermanentlyAnswers) ---

type mockAnswerRepo struct {
	answercommon.AnswerRepo
	deletedAnswers []*entity.Answer
	err            error
}

func (m *mockAnswerRepo) DeletePermanentlyAnswers(ctx context.Context) ([]*entity.Answer, error) {
	return m.deletedAnswers, m.err
}

// --- mock: eventqueue.Service ---

type mockEventQueue struct {
	events []*schema.EventMsg
}

func (m *mockEventQueue) Send(ctx context.Context, msg *schema.EventMsg) {
	m.events = append(m.events, msg)
}
func (m *mockEventQueue) RegisterHandler(handler func(ctx context.Context, msg *schema.EventMsg) error) {
}
func (m *mockEventQueue) Close() {}

// --- helper ---

func newTestService(
	questionRepo questioncommon.QuestionRepo,
	answerRepo answercommon.AnswerRepo,
	eventQueue *mockEventQueue,
) *UserAdminService {
	return &UserAdminService{
		questionCommonRepo: questionRepo,
		answerCommonRepo:   answerRepo,
		eventQueueService:  eventQueue,
	}
}

// --- tests ---

func TestDeletePermanently_Questions_EmitsOneEventPerItem(t *testing.T) {
	questions := []*entity.Question{
		{ID: "10000000000000001", UserID: "10000000000000010"},
		{ID: "10000000000000002", UserID: "10000000000000020"},
		{ID: "10000000000000003", UserID: "10000000000000030"},
	}
	eventQueue := &mockEventQueue{}
	svc := newTestService(
		&mockQuestionRepo{deletedQuestions: questions},
		&mockAnswerRepo{},
		eventQueue,
	)

	err := svc.DeletePermanently(context.Background(), &schema.DeletePermanentlyReq{
		Type:   constant.DeletePermanentlyQuestions,
		UserID: "10000000000000040",
	})
	require.NoError(t, err)

	require.Len(t, eventQueue.events, 3, "should emit one event per deleted question")
	for i, evt := range eventQueue.events {
		assert.Equal(t, constant.EventQuestionDelete, evt.EventType)
		assert.Equal(t, "10000000000000040", evt.UserID)
		assert.Equal(t, questions[i].ID, evt.QuestionID)
		assert.Equal(t, questions[i].UserID, evt.QuestionUserID)
		assert.Equal(t, questions[i].ID, evt.TriggerObjectID)
	}
}

func TestDeletePermanently_Answers_EmitsOneEventPerItem(t *testing.T) {
	answers := []*entity.Answer{
		{ID: "10000000000000004", UserID: "10000000000000050", QuestionID: "10000000000000001"},
		{ID: "10000000000000005", UserID: "10000000000000060", QuestionID: "10000000000000001"},
	}
	eventQueue := &mockEventQueue{}
	svc := newTestService(
		&mockQuestionRepo{},
		&mockAnswerRepo{deletedAnswers: answers},
		eventQueue,
	)

	err := svc.DeletePermanently(context.Background(), &schema.DeletePermanentlyReq{
		Type:   constant.DeletePermanentlyAnswers,
		UserID: "10000000000000040",
	})
	require.NoError(t, err)

	require.Len(t, eventQueue.events, 2, "should emit one event per deleted answer")
	for i, evt := range eventQueue.events {
		assert.Equal(t, constant.EventAnswerDelete, evt.EventType)
		assert.Equal(t, "10000000000000040", evt.UserID)
		assert.Equal(t, answers[i].ID, evt.AnswerID)
		assert.Equal(t, answers[i].UserID, evt.AnswerUserID)
		assert.Equal(t, answers[i].ID, evt.TriggerObjectID)
	}
}

func TestDeletePermanently_Questions_NoItems_NoEvents(t *testing.T) {
	eventQueue := &mockEventQueue{}
	svc := newTestService(
		&mockQuestionRepo{deletedQuestions: nil},
		&mockAnswerRepo{},
		eventQueue,
	)

	err := svc.DeletePermanently(context.Background(), &schema.DeletePermanentlyReq{
		Type:   constant.DeletePermanentlyQuestions,
		UserID: "10000000000000040",
	})
	require.NoError(t, err)

	assert.Empty(t, eventQueue.events, "should emit no events when nothing is deleted")
}

func TestDeletePermanently_Answers_NoItems_NoEvents(t *testing.T) {
	eventQueue := &mockEventQueue{}
	svc := newTestService(
		&mockQuestionRepo{},
		&mockAnswerRepo{deletedAnswers: nil},
		eventQueue,
	)

	err := svc.DeletePermanently(context.Background(), &schema.DeletePermanentlyReq{
		Type:   constant.DeletePermanentlyAnswers,
		UserID: "10000000000000040",
	})
	require.NoError(t, err)

	assert.Empty(t, eventQueue.events, "should emit no events when nothing is deleted")
}
