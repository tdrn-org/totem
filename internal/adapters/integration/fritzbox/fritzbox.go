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

package fritzbox

import (
	"context"
	"log/slog"
	"net/url"

	"github.com/tdrn-org/go-tr064"
	"github.com/tdrn-org/totem/internal/adapters/integration"
	"github.com/tdrn-org/totem/internal/domain"
)

const Name = "fritzbox"

type Config interface {
	LocationID() string
	DeviceURL() *url.URL
}

func Register(host integration.Host, config Config) (integration.Service, error) {
	tr064Client := tr064.NewClient(config.DeviceURL())
	serviceID := integration.NewServiceID(Name, tr064Client.DeviceUrl.Redacted())
	s := &fritzboxService{
		config:      config,
		host:        host,
		serviceID:   serviceID,
		tr064Client: tr064Client,
		logger:      slog.With(slog.String("serviceID", serviceID)),
	}
	host.RegisterService(s)
	return s, nil
}

type fritzboxService struct {
	config      Config
	host        integration.Host
	serviceID   string
	tr064Client *tr064.Client
	logger      *slog.Logger
}

func (s *fritzboxService) Name() string {
	return Name
}

func (s *fritzboxService) ID() string {
	return s.serviceID
}

func (s *fritzboxService) Natures() []integration.Nature {
	return []integration.Nature{integration.NatureLan}
}

func (s *fritzboxService) Shutdown(ctx context.Context) error {
	return nil
}

func (s *fritzboxService) Close() error {
	return nil
}

func (s *fritzboxService) Location() *domain.Location {
	return s.host.ResolveLocationID(s.config.LocationID())
}
