// Package auditsink sends a copy of each audit record somewhere this server cannot reach back into.
//
// The records themselves are rows in this server's own database (see service/auditservice), which is the problem this
// package exists for: they are evidence about the server, kept by the server. Whoever reaches the database can edit or
// delete them, and the retention sweep can delete them by design, so on their own they are an operations log rather
// than something to show an auditor.
//
// A line already sent elsewhere cannot be taken back. That defends against the ordinary case - somebody who gets
// access later, or an insider covering their tracks afterwards - and not against somebody who blocks the sink first
// and then acts. The second case is why a configured sink is a required one by default: an action whose record cannot
// reach the sink does not happen, which is the same stance service/auditservice already takes about the database.
//
// What this does not do is buffer. Point it at a collector on the machine - rsyslog, syslog-ng, vector, fluent-bit -
// and let that forward to wherever the records are kept. A collector is what has the disk queue, the retries and the
// credentials for the far end; a second implementation of all three in here would be worse at it and would be another
// thing holding audit records on this server.
package auditsink

import (
	"context"
	"errors"
)

var (
	// ErrNotConfigured is a sink named without what it needs to reach anything.
	ErrNotConfigured = errors.New("audit sink is not configured")
	// ErrSinkUnavailable is a record that could not be delivered. Whether that stops the action being recorded is the
	// caller's to decide - see config.Audit.Required.
	ErrSinkUnavailable = errors.New("audit sink cannot be reached")
)

// Sink takes one record at a time.
//
// It is deliberately told nothing about what a record is: the body is already JSON by the time it arrives, formatted
// where the audit entry's own fields are known. That keeps this package to transport, which is all that differs
// between one sink and the next.
type Sink interface {
	// Name is what the sink is, for a message that says which one failed.
	Name() string
	// Send delivers one record. msgID says what kind of action it was, short enough to be a field of its own in most
	// log formats; payload is the record as JSON.
	Send(ctx context.Context, msgID string, payload []byte) error
	// Close gives up whatever the sink holds. Sending after it is closed reconnects, so this is for shutdown rather
	// than for a pause.
	Close() error
}

// nop is the sink an installation that configured none gets, so that the service has no branch for whether it has one.
type nop struct{}

// Nop returns a sink that accepts everything and keeps nothing.
func Nop() Sink { return nop{} }

func (nop) Name() string { return "none" }

func (nop) Send(context.Context, string, []byte) error { return nil }

func (nop) Close() error { return nil }
