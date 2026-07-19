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

package notification

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/tdrn-org/go-notify/mattermost"
	"github.com/tdrn-org/totem/config"
	"github.com/tdrn-org/totem/internal/data"
	"github.com/tdrn-org/totem/internal/domain/pulse"
)

type Runtime interface {
	DataStore() *data.Store
}

type Service struct {
	runtime            Runtime
	cfg                *config.NotifyConfig
	mattermostPayloads *mattermost.PayloadFactory[mattermostParams]
	logger             *slog.Logger
}

func NewService(runtime Runtime, cfg *config.NotifyConfig) (*Service, error) {
	config := &mattermostConfig{
		cfg: cfg,
	}
	mattermostPayloads, err := mattermost.NewPayloadFactory(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Mattermost payload factory (cause: %w)", err)
	}
	service := &Service{
		runtime:            runtime,
		cfg:                cfg,
		mattermostPayloads: mattermostPayloads,
		logger:             slog.Default(),
	}
	return service, nil
}

func (s *Service) NotifyTransition(ctx context.Context, log *data.UserPresenceLog) {
	notification := s.cfg.DepartureNotification
	if log.Active {
		notification = s.cfg.ArrivalNotification
	}
	payload := s.mattermostPayloads.NewPayload(notification)
	for _, profile := range s.cfg.Profiles {
		if !profile.Presence {
			continue
		}
		params := &mattermostTransitionParams{
			Agent:    profile.Agent,
			User:     log.User.Name,
			Location: log.Location.Name,
		}
		err := payload.Send(ctx, params)
		if err != nil {
			s.logger.Warn("failed to send transition notification", slog.Any("err", err))
		}
	}
}

func (s *Service) PostPulse(ctx context.Context, pulse *pulse.Pulse) {
	payload := s.mattermostPayloads.NewPayload(s.cfg.PulseNotification)
	params := &mattermostPulseParams{
		Agent:   pulse.Agent,
		Message: pulse.Message,
	}
	err := payload.Send(ctx, params)
	if err != nil {
		s.logger.Warn("failed to send transition notification", slog.Any("err", err))
	}
}

func (s *Service) Pulse(ctx context.Context) {
}
