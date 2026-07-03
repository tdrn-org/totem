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
	"sync"
	"time"

	"github.com/tdrn-org/go-hue"
	"github.com/tdrn-org/totem/internal/adapters/integration"
	"github.com/tdrn-org/totem/internal/domain"
)

const Name = "huebridge"

type Config interface {
	LocationID() string
	Address() string
	BridgeID() string
	Login() string
	Timeout() time.Duration
}

func Register(host integration.Host, config Config) (integration.Service, error) {
	address := config.Address()
	bridgeLocator, err := hue.NewAddressBridgeLocator(address)
	if err != nil {
		return nil, fmt.Errorf("failed to create bridge locator for address '%s' (cause: %w)", config.Address(), err)
	}
	bridgeAuthenticator := hue.NewLocalBridgeAuthenticator(config.Login())
	serviceID := integration.NewServiceID(Name, address)
	s := &huebridgeService{
		config:              config,
		host:                host,
		serviceID:           serviceID,
		bridgeLocator:       bridgeLocator,
		bridgeAuthenticator: bridgeAuthenticator,
		timeout:             config.Timeout(),
		logger:              slog.With(slog.String("serviceID", serviceID)),
	}
	host.RegisterService(s)
	return s, nil
}

type huebridgeService struct {
	config              Config
	host                integration.Host
	serviceID           string
	bridgeLocator       hue.BridgeLocator
	bridgeAuthenticator hue.BridgeAuthenticator
	timeout             time.Duration
	cachedBridgeClient  hue.BridgeClient
	logger              *slog.Logger
	mutex               sync.Mutex
}

func (s *huebridgeService) Name() string {
	return Name
}

func (s *huebridgeService) ID() string {
	return s.serviceID
}

func (s *huebridgeService) Natures() []integration.Nature {
	return []integration.Nature{integration.NatureSmarthome}
}

func (s *huebridgeService) Shutdown(ctx context.Context) error {
	return nil
}

func (s *huebridgeService) Close() error {
	return nil
}

func (s *huebridgeService) Location() *domain.Location {
	return s.host.ResolveLocationID(s.config.LocationID())
}

func (s *huebridgeService) resolveBridgeClient() (hue.BridgeClient, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.cachedBridgeClient != nil {
		return s.cachedBridgeClient, nil
	}
	bridgeID := s.config.BridgeID()
	s.logger.Info("resolving client", slog.String("bridgeID", bridgeID))
	timeout := s.config.Timeout()
	bridge, err := s.bridgeLocator.Lookup(bridgeID, timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup Hue Bridge '%s' (cause: %w)", bridgeID, err)
	}
	bridgeClient, err := bridge.NewClient(s.bridgeAuthenticator, timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to create Hue Bridge client '%s' (cause: %w)", bridgeID, err)
	}
	s.cachedBridgeClient = bridgeClient
	return s.cachedBridgeClient, nil
}

func (s *huebridgeService) taintBridgeClient() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.cachedBridgeClient = nil
}
