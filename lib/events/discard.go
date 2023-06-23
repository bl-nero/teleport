/*
Copyright 2018 Gravitational, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package events

import (
	"context"
	"sync/atomic"

	"github.com/gravitational/trace"
	log "github.com/sirupsen/logrus"

	apievents "github.com/gravitational/teleport/api/types/events"
	"github.com/gravitational/teleport/lib/session"
)

// DiscardAuditLog is do-nothing, discard-everything implementation
// of IAuditLog interface used for cases when audit is turned off
type DiscardAuditLog struct{}

// NewDiscardAuditLog returns a no-op audit log instance
func NewDiscardAuditLog() *DiscardAuditLog {
	return &DiscardAuditLog{}
}

func (d *DiscardAuditLog) Close() error {
	return nil
}

func (d *DiscardAuditLog) GetSessionChunk(namespace string, sid session.ID, offsetBytes, maxBytes int) ([]byte, error) {
	return make([]byte, 0), nil
}

func (d *DiscardAuditLog) GetSessionEvents(namespace string, sid session.ID, after int) ([]EventFields, error) {
	return make([]EventFields, 0), nil
}

func (d *DiscardAuditLog) SearchEvents(ctx context.Context, req SearchEventsRequest) ([]apievents.AuditEvent, string, error) {
	return make([]apievents.AuditEvent, 0), "", nil
}

func (d *DiscardAuditLog) SearchSessionEvents(ctx context.Context, req SearchSessionEventsRequest) ([]apievents.AuditEvent, string, error) {
	return make([]apievents.AuditEvent, 0), "", nil
}

func (d *DiscardAuditLog) EmitAuditEvent(ctx context.Context, event apievents.AuditEvent) error {
	return nil
}

func (d *DiscardAuditLog) StreamSessionEvents(ctx context.Context, sessionID session.ID, startIndex int64) (chan apievents.AuditEvent, chan error) {
	c, e := make(chan apievents.AuditEvent), make(chan error, 1)
	close(c)
	return c, e
}

// NewDiscardStream returns a no-op discard stream
func NewDiscardStream() *DiscardStream {
	return &DiscardStream{
		done: make(chan struct{}),
	}
}

// DiscardStream returns a stream that discards all events
type DiscardStream struct {
	completed atomic.Bool
	done      chan struct{}
}

// Write discards data
func (d *DiscardStream) Write(p []byte) (n int, err error) {
	if d.completed.Load() {
		return 0, trace.BadParameter("stream is closed")
	}

	return len(p), nil
}

// Status returns a channel that always blocks
func (*DiscardStream) Status() <-chan apievents.StreamStatus {
	return nil
}

// Done returns channel closed when streamer is closed
// should be used to detect sending errors
func (d *DiscardStream) Done() <-chan struct{} {
	return d.done
}

// Close flushes non-uploaded flight stream data without marking
// the stream completed and closes the stream instance
func (d *DiscardStream) Close(ctx context.Context) error {
	return d.Complete(ctx)
}

// Complete marks the stream as closed
func (d *DiscardStream) Complete(ctx context.Context) error {
	if !d.completed.Load() {
		close(d.done)
	}
	d.completed.Store(true)
	return nil
}

// SetupEvent does nothing
func (*DiscardStream) SetupEvent(event apievents.AuditEvent) error {
	return nil
}

// RecordEvent discards session event
func (d *DiscardStream) RecordEvent(ctx context.Context, event apievents.AuditEvent) error {
	if d.completed.Load() {
		return trace.BadParameter("stream is closed")
	}

	log.WithFields(log.Fields{
		"event_id":    event.GetID(),
		"event_type":  event.GetType(),
		"event_time":  event.GetTime(),
		"event_index": event.GetIndex(),
	}).Traceln("Discarding stream event")
	return nil
}

// NewDiscardEmitter returns a no-op discard emitter
func NewDiscardEmitter() *DiscardEmitter {
	return &DiscardEmitter{}
}

// DiscardEmitter discards all events
type DiscardEmitter struct{}

// EmitAuditEvent discards audit event
func (*DiscardEmitter) EmitAuditEvent(ctx context.Context, event apievents.AuditEvent) error {
	log.WithFields(log.Fields{
		"event_id":    event.GetID(),
		"event_type":  event.GetType(),
		"event_time":  event.GetTime(),
		"event_index": event.GetIndex(),
	}).Debugf("Discarding event")
	return nil
}

// CreateAuditStream creates a stream that discards all events
func (*DiscardEmitter) CreateAuditStream(ctx context.Context, sid session.ID) (apievents.Stream, error) {
	return NewDiscardStream(), nil
}

// ResumeAuditStream resumes a stream that discards all events
func (*DiscardEmitter) ResumeAuditStream(ctx context.Context, sid session.ID, uploadID string) (apievents.Stream, error) {
	return NewDiscardStream(), nil
}
