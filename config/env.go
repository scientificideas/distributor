/*
Copyright LLC Newity. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package config

import "github.com/caarlos0/env/v6"

func envVarsToConfig() (*Config, error) {
	cfg := new(Config)
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
