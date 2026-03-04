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

package constant

// EventType event type. It is used to define the type of event. Such as object.action
type EventType string

// event object
const (
	eventQuestion = "question"
	eventAnswer   = "answer"
	eventComment  = "comment"
	eventUser     = "user"
)

// event action
const (
	eventCreate   = "create"
	eventUpdate   = "update"
	eventDelete   = "delete"
	eventStatus   = "status"
	eventVote     = "vote"
	eventAccept   = "accept"
	eventUnaccept = "unaccept"
	eventShare    = "share" // the object share link has been clicked
	eventFlag     = "flag"
	eventReact    = "react"
	eventHide     = "hide"
	eventShow     = "show"
	eventPin      = "pin"
	eventUnpin    = "unpin"
)

const (
	EventUserUpdate EventType = eventUser + "." + eventUpdate
	EventUserShare  EventType = eventUser + "." + eventShare
)

const (
	EventQuestionCreate  EventType = eventQuestion + "." + eventCreate
	EventQuestionUpdate  EventType = eventQuestion + "." + eventUpdate
	EventQuestionStatus  EventType = eventQuestion + "." + eventStatus
	EventQuestionDelete  EventType = eventQuestion + "." + eventDelete
	EventQuestionVote    EventType = eventQuestion + "." + eventVote
	EventQuestionFlag    EventType = eventQuestion + "." + eventFlag
	EventQuestionReact   EventType = eventQuestion + "." + eventReact
	EventQuestionHide    EventType = eventQuestion + "." + eventHide
	EventQuestionShow    EventType = eventQuestion + "." + eventShow
	EventQuestionPin     EventType = eventQuestion + "." + eventPin
	EventQuestionUnpin   EventType = eventQuestion + "." + eventUnpin
)

const (
	EventAnswerCreate   EventType = eventAnswer + "." + eventCreate
	EventAnswerUpdate   EventType = eventAnswer + "." + eventUpdate
	EventAnswerStatus   EventType = eventAnswer + "." + eventStatus
	EventAnswerDelete   EventType = eventAnswer + "." + eventDelete
	EventAnswerVote     EventType = eventAnswer + "." + eventVote
	EventAnswerAccept   EventType = eventAnswer + "." + eventAccept
	EventAnswerUnaccept EventType = eventAnswer + "." + eventUnaccept
	EventAnswerFlag     EventType = eventAnswer + "." + eventFlag
	EventAnswerReact    EventType = eventAnswer + "." + eventReact
)

const (
	EventCommentCreate EventType = eventComment + "." + eventCreate
	EventCommentUpdate EventType = eventComment + "." + eventUpdate
	EventCommentDelete EventType = eventComment + "." + eventDelete
	EventCommentVote   EventType = eventComment + "." + eventVote
	EventCommentFlag   EventType = eventComment + "." + eventFlag
)
