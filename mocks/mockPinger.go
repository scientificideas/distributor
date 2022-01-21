/*
Copyright Scientific Ideas 2022. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"context"
	"fmt"
	"strings"
)

type MockPinger struct{}

func NewMockPinger() *MockPinger {
	return &MockPinger{}
}

func (p *MockPinger) Init(_ ...string) error {
	return nil
}

func (p *MockPinger) Ping(_ context.Context, url string) error {
	if strings.Contains(url, "bad") {
		return fmt.Errorf("bad request")
	}

	return nil
}
