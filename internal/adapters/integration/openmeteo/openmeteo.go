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

package openmeteo

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"

	openmeteoapi "github.com/tdrn-org/go-open-meteo"
	"github.com/tdrn-org/totem/config"
	"github.com/tdrn-org/totem/internal/adapters/integration"
)

const Name = "openmeteo"

type Config interface {
	APIURL() *url.URL
	TemperatureUnit() config.TemperatureUnit
	PrecipitationUnit() config.PrecipitationUnit
	WindSpeedUnit() config.WindSpeedUnit
}

func Register(host integration.Host, config Config) (integration.Service, error) {
	apiURL := config.APIURL()
	apiClient, err := openmeteoapi.NewClient(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create API client for URL %q (cause: %w)", apiURL.Redacted(), err)
	}
	serviceID := integration.NewServiceID(Name, apiURL.Redacted())
	s := &openmeteoService{
		config:    config,
		serviceID: serviceID,
		apiClient: apiClient,
		logger:    slog.With(slog.String("serviceID", serviceID)),
	}
	host.RegisterService(s)
	return s, nil
}

type openmeteoService struct {
	config    Config
	serviceID string
	apiClient *openmeteoapi.Client
	logger    *slog.Logger
}

func (s *openmeteoService) Name() string {
	return Name
}

func (s *openmeteoService) ID() string {
	return s.serviceID
}

func (s *openmeteoService) Natures() []integration.Nature {
	return []integration.Nature{integration.NatureWeather}
}

func (s *openmeteoService) Shutdown(ctx context.Context) error {
	return nil
}

func (s *openmeteoService) Close() error {
	return nil
}
