/*
 * Copyright 2026 Holger de Carne
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package huebridge

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/tdrn-org/go-hue"
	"github.com/tdrn-org/totem/internal/adapters/integration"
)

const Name = "huebridge"

type Config interface {
	LocationID() string
	Address() string
	BridgeID() string
}

func Register(host integration.Host, config Config) (integration.Service, error) {
	address := config.Address()
	bridgeLocator, err := hue.NewAddressBridgeLocator(address)
	if err != nil {
		return nil, fmt.Errorf("failed to create bridge locator for address '%s' (cause: %w)", config.Address(), err)
	}
	serviceID := integration.NewServiceID(Name, address)
	s := &hueService{
		config:        config,
		serviceID:     serviceID,
		bridgeLocator: bridgeLocator,
		logger:        slog.With(slog.String("serviceID", serviceID)),
	}
	host.RegisterService(s)
	return s, nil
}

type hueService struct {
	config        Config
	serviceID     string
	bridgeLocator hue.BridgeLocator
	logger        *slog.Logger
}

func (s *hueService) Name() string {
	return Name
}

func (s *hueService) ID() string {
	return s.serviceID
}

func (s *hueService) Shutdown(ctx context.Context) error {
	return nil
}

func (s *hueService) Close() error {
	return nil
}
