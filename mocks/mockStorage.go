/*
Copyright Scientific Ideas 2022. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"strings"
)

type MockStorage struct {
	Lists     map[string][]string
	HashTable map[string]string
}

func (m *MockStorage) GetList(key string) ([]string, error) {
	return m.Lists[key], nil
}

func (m *MockStorage) SetMap(_ string, value map[string]interface{}) error {
	for k, v := range value {
		m.HashTable[k] = v.(string)
	}

	return nil
}

func (m *MockStorage) DelFromMap(_ string, field string) error {
	delete(m.HashTable, field)

	return nil
}

func (m *MockStorage) DelFromList(listname string, item string) error {
	var newSlice []string

	for _, value := range m.Lists[listname] {
		if value != item {
			newSlice = append(newSlice, value)
		}
	}

	m.Lists[listname] = newSlice

	return nil
}

func (m *MockStorage) GetMapField(_, field string) ([]string, error) {
	return strings.Split(m.HashTable[field], ","), nil
}

func (m *MockStorage) Close() error {
	return nil
}
