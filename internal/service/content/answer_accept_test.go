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

package content

import (
	"context"
	"testing"

	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/activity"
	answercommon "github.com/apache/answer/internal/service/answer_common"
	"github.com/apache/answer/internal/service/config"
	questioncommon "github.com/apache/answer/internal/service/question_common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- mock: QuestionRepo (only methods used by AcceptAnswer path) ---

type mockQuestionRepo struct {
	questioncommon.QuestionRepo
	questions map[string]*entity.Question
}

func (m *mockQuestionRepo) GetQuestion(ctx context.Context, id string) (*entity.Question, bool, error) {
	q, ok := m.questions[id]
	if !ok {
		return nil, false, nil
	}
	cp := *q
	return &cp, true, nil
}
func (m *mockQuestionRepo) UpdateAccepted(ctx context.Context, question *entity.Question) error {
	if q, ok := m.questions[question.ID]; ok {
		q.AcceptedAnswerID = question.AcceptedAnswerID
	}
	return nil
}

// --- mock: AnswerRepo (only methods used by AcceptAnswer path) ---

type mockAnswerRepo struct {
	answercommon.AnswerRepo
	answers map[string]*entity.Answer
}

func (m *mockAnswerRepo) GetByID(ctx context.Context, answerID string) (*entity.Answer, bool, error) {
	a, ok := m.answers[answerID]
	if !ok {
		return nil, false, nil
	}
	cp := *a
	return &cp, true, nil
}
func (m *mockAnswerRepo) UpdateAcceptedStatus(ctx context.Context, acceptedAnswerID string, questionID string) error {
	return nil
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

// --- mock: AnswerActivityRepo (for updateAnswerRank) ---

type mockAnswerActivityRepo struct{}

func (m *mockAnswerActivityRepo) SaveAcceptAnswerActivity(ctx context.Context, op *schema.AcceptAnswerOperationInfo) error {
	return nil
}
func (m *mockAnswerActivityRepo) SaveCancelAcceptAnswerActivity(ctx context.Context, op *schema.AcceptAnswerOperationInfo) error {
	return nil
}

// --- mock: ConfigRepo (for AnswerActivityService → configService) ---

type mockConfigRepo struct{}

func (m *mockConfigRepo) GetConfigByID(ctx context.Context, id int) (*entity.Config, error) {
	return &entity.Config{ID: id, Key: "mock", Value: "0"}, nil
}
func (m *mockConfigRepo) GetConfigByKey(ctx context.Context, key string) (*entity.Config, error) {
	return &entity.Config{ID: 1, Key: key, Value: "0"}, nil
}
func (m *mockConfigRepo) GetConfigByKeyFromDB(ctx context.Context, key string) (*entity.Config, error) {
	return &entity.Config{ID: 1, Key: key, Value: "0"}, nil
}
func (m *mockConfigRepo) UpdateConfig(ctx context.Context, key, value string) error {
	return nil
}

// --- helpers ---

func newTestAnswerService(
	questionRepo questioncommon.QuestionRepo,
	answerRepo answercommon.AnswerRepo,
	eventQueue *mockEventQueue,
) *AnswerService {
	qc := questioncommon.NewQuestionCommon(
		questionRepo, answerRepo,
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
	)

	configSvc := config.NewConfigService(&mockConfigRepo{})
	activitySvc := activity.NewAnswerActivityService(&mockAnswerActivityRepo{}, configSvc)

	return &AnswerService{
		questionRepo:          questionRepo,
		answerRepo:            answerRepo,
		questionCommon:        qc,
		answerActivityService: activitySvc,
		eventQueueService:     eventQueue,
	}
}

const (
	testQuestionID = "10000000000000001"
	testAnswerID1  = "10000000000000002"
	testAnswerID2  = "10000000000000003"
	testQUserID    = "10000000000000010"
	testAUserID1   = "10000000000000020"
	testAUserID2   = "10000000000000030"
	testUserID     = "10000000000000040"
)

func TestAcceptAnswer_EmitsAcceptEvent(t *testing.T) {
	questionRepo := &mockQuestionRepo{
		questions: map[string]*entity.Question{
			testQuestionID: {
				ID:               testQuestionID,
				UserID:           testQUserID,
				AcceptedAnswerID: "0",
			},
		},
	}
	answerRepo := &mockAnswerRepo{
		answers: map[string]*entity.Answer{
			testAnswerID1: {ID: testAnswerID1, QuestionID: testQuestionID, UserID: testAUserID1},
		},
	}
	eventQueue := &mockEventQueue{}

	svc := newTestAnswerService(questionRepo, answerRepo, eventQueue)

	err := svc.AcceptAnswer(context.Background(), &schema.AcceptAnswerReq{
		QuestionID: testQuestionID,
		AnswerID:   testAnswerID1,
		UserID:     testUserID,
	})
	require.NoError(t, err)

	require.Len(t, eventQueue.events, 1, "should emit exactly one event")
	evt := eventQueue.events[0]
	assert.Equal(t, constant.EventAnswerAccept, evt.EventType)
	assert.Equal(t, testUserID, evt.UserID)
	assert.Equal(t, testQuestionID, evt.QuestionID)
	assert.Equal(t, testAnswerID1, evt.AnswerID)
	assert.Equal(t, testAUserID1, evt.AnswerUserID)
}

func TestAcceptAnswer_EmitsUnacceptEvent(t *testing.T) {
	questionRepo := &mockQuestionRepo{
		questions: map[string]*entity.Question{
			testQuestionID: {
				ID:               testQuestionID,
				UserID:           testQUserID,
				AcceptedAnswerID: testAnswerID1, // currently accepted
			},
		},
	}
	answerRepo := &mockAnswerRepo{
		answers: map[string]*entity.Answer{
			testAnswerID1: {ID: testAnswerID1, QuestionID: testQuestionID, UserID: testAUserID1},
		},
	}
	eventQueue := &mockEventQueue{}

	svc := newTestAnswerService(questionRepo, answerRepo, eventQueue)

	// Unaccept: AnswerID="0" means remove acceptance
	err := svc.AcceptAnswer(context.Background(), &schema.AcceptAnswerReq{
		QuestionID: testQuestionID,
		AnswerID:   "0",
		UserID:     testUserID,
	})
	require.NoError(t, err)

	require.Len(t, eventQueue.events, 1, "should emit exactly one event")
	evt := eventQueue.events[0]
	assert.Equal(t, constant.EventAnswerUnaccept, evt.EventType)
	assert.Equal(t, testUserID, evt.UserID)
	assert.Equal(t, testQuestionID, evt.QuestionID)
	assert.Equal(t, testAnswerID1, evt.AnswerID)
	assert.Equal(t, testAUserID1, evt.AnswerUserID)
}

func TestAcceptAnswer_NoEventWhenAlreadyAccepted(t *testing.T) {
	questionRepo := &mockQuestionRepo{
		questions: map[string]*entity.Question{
			testQuestionID: {
				ID:               testQuestionID,
				UserID:           testQUserID,
				AcceptedAnswerID: testAnswerID1, // already accepted
			},
		},
	}
	answerRepo := &mockAnswerRepo{
		answers: map[string]*entity.Answer{
			testAnswerID1: {ID: testAnswerID1, QuestionID: testQuestionID, UserID: testAUserID1},
		},
	}
	eventQueue := &mockEventQueue{}

	svc := newTestAnswerService(questionRepo, answerRepo, eventQueue)

	// Accept the same answer that is already accepted → no-op
	err := svc.AcceptAnswer(context.Background(), &schema.AcceptAnswerReq{
		QuestionID: testQuestionID,
		AnswerID:   testAnswerID1,
		UserID:     testUserID,
	})
	require.NoError(t, err)

	assert.Empty(t, eventQueue.events, "should not emit any event when answer is already accepted")
}

func TestAcceptAnswer_SwitchAcceptedEmitsAcceptEvent(t *testing.T) {
	questionRepo := &mockQuestionRepo{
		questions: map[string]*entity.Question{
			testQuestionID: {
				ID:               testQuestionID,
				UserID:           testQUserID,
				AcceptedAnswerID: testAnswerID1, // currently accepted
			},
		},
	}
	answerRepo := &mockAnswerRepo{
		answers: map[string]*entity.Answer{
			testAnswerID1: {ID: testAnswerID1, QuestionID: testQuestionID, UserID: testAUserID1},
			testAnswerID2: {ID: testAnswerID2, QuestionID: testQuestionID, UserID: testAUserID2},
		},
	}
	eventQueue := &mockEventQueue{}

	svc := newTestAnswerService(questionRepo, answerRepo, eventQueue)

	// Switch from answer1 to answer2
	err := svc.AcceptAnswer(context.Background(), &schema.AcceptAnswerReq{
		QuestionID: testQuestionID,
		AnswerID:   testAnswerID2,
		UserID:     testUserID,
	})
	require.NoError(t, err)

	// Should emit accept for the new answer (not unaccept for the old one)
	require.Len(t, eventQueue.events, 1, "should emit exactly one event for the new accepted answer")
	evt := eventQueue.events[0]
	assert.Equal(t, constant.EventAnswerAccept, evt.EventType)
	assert.Equal(t, testAnswerID2, evt.AnswerID)
	assert.Equal(t, testAUserID2, evt.AnswerUserID)
}
