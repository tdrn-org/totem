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

package mcp

import (
	"log/slog"
	"net/http"
	"net/url"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tdrn-org/totem/internal/adapters/integration"
	"github.com/tdrn-org/totem/internal/application"
	"github.com/tdrn-org/totem/internal/buildinfo"
	"github.com/tdrn-org/totem/internal/data"
	"github.com/tdrn-org/totem/internal/domain"
)

// Runtime provides the MCP handler access to the server's services and configuration.
type Runtime interface {
	BaseURL() *url.URL
	Logger() *slog.Logger
	Services() map[string]integration.Service
	DataStore() *data.Store
	PresenceTracker() *application.PresenceTracker
	ResolveLocationID(locationID string) *domain.Location
}

// filterServices filters the runtime's services by nature and type-asserts them to T.
// Returns only services that declare the given nature and implement T.
func filterServices[T any](services map[string]integration.Service, nature integration.Nature, logger *slog.Logger) []T {
	var result []T
	for _, service := range services {
		for _, n := range service.Natures() {
			if n != nature {
				continue
			}
			typed, ok := any(service).(T)
			if !ok {
				logger.Warn("service declares nature but does not implement expected interface",
					slog.String("serviceID", service.ID()),
					slog.String("nature", string(nature)))
				continue
			}
			result = append(result, typed)
		}
	}
	return result
}

func NewHandler(runtime Runtime) http.Handler {
	impl := &mcp.Implementation{
		Name:       buildinfo.Cmd(),
		Version:    buildinfo.Version(),
		WebsiteURL: runtime.BaseURL().String(),
	}
	options := &mcp.ServerOptions{
		Logger: runtime.Logger(),
	}
	server := mcp.NewServer(impl, options)

	// Register tools from services based on their natures
	registerSmarthomeTools(server, runtime)
	registerMediaTools(server, runtime)
	registerPresenceTools(server, runtime)

	getServer := func(_ *http.Request) *mcp.Server {
		return server
	}
	return mcp.NewStreamableHTTPHandler(getServer, nil)
}
