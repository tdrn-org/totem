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

package totem

import (
	"log/slog"

	"github.com/tdrn-org/totem/internal/adapters/integration"
	"github.com/tdrn-org/totem/internal/domain"
)

func (s *Server) integrationHost() integration.Host {
	return &integrationHost{s: s}
}

func (s *Server) registerService(service integration.Service) {
	serviceID := service.ID()
	s.logger.Info("registering service", slog.String("serviceID", serviceID))
	_, exists := s.services[serviceID]
	if exists {
		s.logger.Info("re-registering service", slog.String("serviceID", serviceID))
	}
	s.services[serviceID] = service
}

type integrationHost struct {
	s *Server
}

func (h *integrationHost) RegisterService(service integration.Service) {
	h.s.registerService(service)
}

func (h *integrationHost) ResolveLocationID(locationID string) *domain.Location {
	location, _ := h.s.resolveLocationID(locationID)
	return location
}
