/*
Copyright LLC Newity. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"github.com/scientificideas/distributor/storage"
)

type Option func(d *Distributor) error

// WithTransport creates a network-related configuration for Distributor.
func WithTransport(tr *Transport) Option {
	return func(d *Distributor) error {
		d.transport = tr

		return nil
	}
}

// WithRedis injects Redis storage client into the Distributor.
func WithRedis(password string, addresses []string, withTLS bool, redisRootCACerts []string) Option {
	return func(d *Distributor) error {
		stor, err := storage.NewRedis(password, addresses, withTLS, redisRootCACerts)
		if err != nil {
			return err
		}

		d.Storage = stor

		return nil
	}
}

// WithStorage injects any Storage implementation into the Distributor.
func WithStorage(stor storage.Storage) Option {
	return func(d *Distributor) error {
		d.Storage = stor

		return nil
	}
}
