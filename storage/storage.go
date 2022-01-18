/*
Copyright LLC Newity. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package storage

// Storage describes minimal interface for Distributor storage implementation.
// Distributor's storage provides services lists, work units list and matching table.
type Storage interface {
	GetList(key string) ([]string, error)
	SetMap(key string, m map[string]interface{}) error
	DelFromMap(mapname string, field string) error
	DelFromList(listname string, item string) error
	GetMapField(key, field string) ([]string, error)
	Close() error
}
