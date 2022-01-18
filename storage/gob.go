/*
Copyright LLC Newity. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package storage

import (
	"bytes"
	"encoding/gob"
)

// Encode serializes string slice to the gob-encoded data.
func Encode(data []string) ([]byte, error) {
	var bytebuffer bytes.Buffer
	e := gob.NewEncoder(&bytebuffer)

	if err := e.Encode(data); err != nil {
		return nil, err
	}

	return bytebuffer.Bytes(), nil
}

// Encode deserializes gob-encoded data to the string slice.
func Decode(data []byte) ([]string, error) {
	var decoded []string

	bytebuffer := bytes.NewBuffer(data)
	d := gob.NewDecoder(bytebuffer)

	if err := d.Decode(&decoded); err != nil {
		return nil, err
	}

	return decoded, nil
}
