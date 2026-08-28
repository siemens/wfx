package persistence

/*
 * SPDX-FileCopyrightText: 2026 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import "context"

type clientIDContextKey struct{}

// WithClientID restricts job persistence queries and mutations using ctx to clientID.
func WithClientID(ctx context.Context, clientID string) context.Context {
	return context.WithValue(ctx, clientIDContextKey{}, clientID)
}

// ClientIDFromContext returns the client ID restriction attached to ctx.
func ClientIDFromContext(ctx context.Context) (string, bool) {
	clientID, ok := ctx.Value(clientIDContextKey{}).(string)
	return clientID, ok
}
