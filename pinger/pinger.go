/*
Copyright LLC Newity. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package pinger

import "context"

// Pinger is responsible for checking services liveness.
type Pinger interface {
	// Init makes any needed by implementation initialization work
	Init(urls ...string) error
	// Ping makes a ping call to the service to check it's liveness
	Ping(ctx context.Context, url string) error
}
