/*
Copyright LLC Newity. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package storage

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"strings"
	"time"

	"github.com/go-redis/redis/v7"
)

// Redis struct implements Distributor Storage interface for RedisDB.
type Redis struct {
	Client redis.UniversalClient
}

// NewRedis creates Redis storage implementation instance.
func NewRedis(password string, addrs []string, withTLS bool, rootCAs []string) (*Redis, error) {
	var (
		redisOpts *redis.UniversalOptions
		client    redis.UniversalClient
	)

	if withTLS {
		certPool := x509.NewCertPool()

		for _, rootCA := range rootCAs {
			cert, err := ioutil.ReadFile(rootCA)
			if err != nil {
				return nil, fmt.Errorf("failed to read root CA certificate %s", rootCA)
			}

			if ok := certPool.AppendCertsFromPEM(cert); !ok {
				return nil, fmt.Errorf(
					"failed to add root CA certificate %s to the certificate pool",
					rootCA,
				)
			}
		}

		redisOpts = &redis.UniversalOptions{
			Addrs:           addrs,
			Password:        password,
			ReadOnly:        false,
			MaxRetries:      10000,
			MaxRetryBackoff: 20 * time.Second,
			TLSConfig:       &tls.Config{RootCAs: certPool},
		}
	} else {
		redisOpts = &redis.UniversalOptions{
			Addrs:           addrs,
			Password:        password,
			ReadOnly:        false,
			MaxRetries:      10000,
			MaxRetryBackoff: 20 * time.Second,
		}
	}

	var (
		err     error
		success = make(chan struct{}, 1)
	)
	go func() {
		client = redis.NewUniversalClient(redisOpts)
		err = client.Ping().Err()
		success <- struct{}{}
	}()

	go func() {
		t := time.NewTicker(20 * time.Second)
		connHealthy := true
		for range t.C {
			if client == nil {
				logrus.Warn("Redis client is nil")
				continue
			}

			successPing := make(chan struct{})
			go func() {
				if err = client.Ping().Err(); err != nil {
					return
				}
				close(successPing)
			}()
			select {
			case <-time.After(2 * time.Second):
				logrus.Warn("Redis ping timeout")
				connHealthy = false
			case <-successPing:
				if !connHealthy {
					logrus.Info("Redis connection restored")
					connHealthy = true
				}
			}

		}
	}()

	select {
	case <-time.After(10 * time.Second):
		err = errors.New("failed connect to the Redis")
	case <-success:
	}

	return &Redis{client}, err
}

// Put saves value for for key.
func (r *Redis) Put(key string, value []string) error {
	data, err := Encode(value)
	if err != nil {
		return nil
	}

	return r.Client.Set(key, data, 0).Err()
}

// Get retrieves value of the key.
func (r *Redis) Get(key string) ([]string, error) {
	resp := r.Client.Get(key)
	if resp.Err() != nil {
		return nil, resp.Err()
	}

	data, err := resp.Bytes()
	if err != nil {
		return nil, err
	}

	return Decode(data)
}

// GetList returns data saved for key in Redis List.
func (r *Redis) GetList(key string) ([]string, error) {
	cmd := r.Client.LRange(key, 0, -1)

	var members []string

	if err := cmd.ScanSlice(&members); err != nil {
		return nil, err
	}

	return members, nil
}

// DelFromList removes item from array saved for key in Redis List.
func (r *Redis) DelFromList(listname, item string) error {
	return r.Client.LRem(listname, 0, item).Err()
}

// SetMap creates or updates Redis Hash.
func (r *Redis) SetMap(key string, m map[string]interface{}) error {
	return r.Client.HSet(key, m).Err()
}

// DelFromMap removes data stored in Redis Hash.
func (r *Redis) DelFromMap(mapname string, field string) error {
	return r.Client.HDel(mapname, field).Err()
}

// Delete removes data for this key in Redis.
func (r *Redis) Delete(key string) error {
	return r.Client.Del(key).Err()
}

// Close closes connection to Redis.
func (r *Redis) Close() error {
	return r.Client.Close()
}

// GetMapField returns value for given map field.
func (r *Redis) GetMapField(key, field string) ([]string, error) {
	resp := r.Client.HGet(key, field)
	if resp.Err() != nil {
		return nil, resp.Err()
	}

	res, err := resp.Result()
	if err != nil {
		return nil, err
	}

	return strings.Split(res, ","), nil
}
