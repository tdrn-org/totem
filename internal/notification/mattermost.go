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

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/tdrn-org/totem/config"
)

type mattermostParams interface {
	GetAgent() string
}

type mattermostTransitionParams struct {
	Agent    string
	User     string
	Location string
}

func (p *mattermostTransitionParams) GetAgent() string {
	return p.Agent
}

type mattermostScheduleParams struct {
	Agent string
}

func (p *mattermostScheduleParams) GetAgent() string {
	return p.Agent
}

type mattermostPulseParams struct {
	Agent   string
	Message string
}

func (p *mattermostPulseParams) GetAgent() string {
	return p.Agent
}

type mattermostConfig struct {
	cfg *config.NotifyConfig
}

func (c *mattermostConfig) GetServerURL() (string, error) {
	return c.cfg.MattermostURL, nil
}

func (c *mattermostConfig) GetToken() (string, error) {
	return c.cfg.MattermostToken, nil
}

func (c *mattermostConfig) ResolveChannelID(ctx context.Context, client *model.Client4, params mattermostParams) (string, error) {
	user, _, err := client.GetMe(ctx, "")
	if err != nil {
		return "", fmt.Errorf("failed to retrieve User (cause: %w)", err)
	}
	agent, _, err := client.GetUserByUsername(ctx, params.GetAgent(), "")
	if err != nil {
		return "", fmt.Errorf("failed to retrieve Agent '%s' (cause: %w)", params.GetAgent(), err)
	}
	channel, _, err := client.CreateDirectChannel(ctx, user.Id, agent.Id)
	if err != nil {
		return "", fmt.Errorf("failed to create direct channel to Agent '%s' (cause: %w)", params.GetAgent(), err)
	}
	return channel.Id, nil
}
