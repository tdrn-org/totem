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

package jellyfin

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"

	jellyfinapi "github.com/sj14/jellyfin-go/api"
	"github.com/tdrn-org/totem/internal/adapters/integration"
)

const Name = "jellyfin"

type Config interface {
	JellyfinURL() *url.URL
	APIToken() string
}

func Register(host integration.Host, config Config) (integration.Service, error) {
	url := config.JellyfinURL()
	apiClientConfig := &jellyfinapi.Configuration{
		Servers:       jellyfinapi.ServerConfigurations{{URL: url.String()}},
		DefaultHeader: map[string]string{"Authorization": fmt.Sprintf("MediaBrowser Token=\"%s\"", config.APIToken())},
	}
	apiClient := jellyfinapi.NewAPIClient(apiClientConfig)
	serviceID := integration.NewServiceID(Name, url.Redacted())
	s := &jellyfinService{
		config:            config,
		serviceID:         serviceID,
		jellyfinapiClient: apiClient,
		logger:            slog.With(slog.String("serviceID", serviceID)),
	}
	host.Register(s)
	return s, nil
}

type jellyfinService struct {
	config            Config
	serviceID         string
	jellyfinapiClient *jellyfinapi.APIClient
	logger            *slog.Logger
}

func (s *jellyfinService) Name() string {
	return Name
}

func (s *jellyfinService) ID() string {
	return s.serviceID
}

func (s *jellyfinService) Shutdown(ctx context.Context) error {
	return nil
}

func (s *jellyfinService) Close() error {
	return nil
}
