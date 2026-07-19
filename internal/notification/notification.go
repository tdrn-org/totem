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

	"github.com/tdrn-org/go-notify/mattermost"
	"github.com/tdrn-org/totem/internal/data"
)

type Runtime interface {
	DataStore() *data.Store
}

type Service struct {
	runtime            Runtime
	mattermostPayloads *mattermost.PayloadFactory[*mattermostParams]
}

func NewService(runtime Runtime, mattermostURL, mattermostToken string) (*Service, error) {
	config := &mattermostConfig{
		serverURL: mattermostURL,
		token:     mattermostToken,
	}
	mattermostPayloads, err := mattermost.NewPayloadFactory(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Mattermost payload factory (cause: %w)", err)
	}
	service := &Service{
		runtime:            runtime,
		mattermostPayloads: mattermostPayloads,
	}
	return service, nil
}

func (s *Service) NotifyTransition(ctx context.Context, log *data.UserPresenceLog) {

}

func (s *Service) PostPulse(ctx context.Context) error {
	return nil
}

func (s *Service) Pulse(ctx context.Context) {
}
