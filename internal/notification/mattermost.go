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
	"net/url"
	"time"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/tdrn-org/go-notify/mattermost"
	"github.com/tdrn-org/totem/config"
)

type mattermostParams interface {
	GetAgent() string
	GetCooldown() time.Duration
}

type mattermostTransitionParams struct {
	Agent    string
	User     string
	Location string
}

func (p *mattermostTransitionParams) GetAgent() string {
	return p.Agent
}

func (p *mattermostTransitionParams) GetCooldown() time.Duration {
	return 0
}

type mattermostAgentParams struct {
	Agent    string
	Cooldown time.Duration
}

func (p *mattermostAgentParams) GetAgent() string {
	return p.Agent
}

func (p *mattermostAgentParams) GetCooldown() time.Duration {
	return p.Cooldown
}

type mattermostMessageParams struct {
	Agent   string
	Message string
}

func (p *mattermostMessageParams) GetAgent() string {
	return p.Agent
}

func (p *mattermostMessageParams) GetCooldown() time.Duration {
	return 0
}

type mattermostConfig struct {
	cfg *config.NotifyConfig
}

func (c *mattermostConfig) GetServerURL() (*url.URL, error) {
	return c.cfg.MattermostURL.URL, nil
}

func (c *mattermostConfig) GetToken() (string, error) {
	return c.cfg.MattermostToken, nil
}

func (c *mattermostConfig) ResolveChannelID(ctx context.Context, client *model.Client4, params mattermostParams) (string, error) {
	user, err := c.getClientUser(ctx, client)
	if err != nil {
		return "", err
	}
	agent, _, err := client.GetUserByUsername(ctx, params.GetAgent(), "")
	if err != nil {
		return "", fmt.Errorf("failed to retrieve Agent %q (cause: %w)", params.GetAgent(), err)
	}
	channel, _, err := client.CreateDirectChannel(ctx, user.Id, agent.Id)
	if err != nil {
		return "", fmt.Errorf("failed to create direct channel to Agent %q (cause: %w)", params.GetAgent(), err)
	}
	return channel.Id, nil
}

func (c *mattermostConfig) AuthorizeSend(ctx context.Context, client *model.Client4, channelID string, params mattermostParams) (bool, error) {
	cooldown := params.GetCooldown()
	if cooldown <= 0 {
		return true, nil
	}
	user, err := c.getClientUser(ctx, client)
	if err != nil {
		return false, err
	}
	earliest := time.Now().Add(-cooldown)
	page := 0
	for {
		postList, _, err := client.GetPostsForChannel(ctx, channelID, page, 50, "", false, false)
		if err != nil {
			return false, fmt.Errorf("failed to retrieve channel posts (cause: %w)", err)
		}
		if len(postList.Order) == 0 {
			return true, nil
		}
		for _, postID := range postList.Order {
			post := postList.Posts[postID]
			postCreatedAt := time.UnixMilli(post.CreateAt)
			if postCreatedAt.Before(earliest) {
				return true, nil
			}
			if post.UserId == user.Id {
				return false, nil
			}
		}
		page++
	}
}

func (c *mattermostConfig) getClientUser(ctx context.Context, client *model.Client4) (*model.User, error) {
	user, _, err := client.GetMe(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve User (cause: %w)", err)
	}
	return user, nil
}

func newPayloadFactory(cfg *config.NotifyConfig) (*mattermost.PayloadFactory[mattermostParams], error) {
	if cfg.MattermostURL.URL == nil {
		return nil, fmt.Errorf("missing Mattermost URL configuration")
	}
	config := &mattermostConfig{
		cfg: cfg,
	}
	mattermostPayloads, err := mattermost.NewPayloadFactory(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Mattermost payload factory (cause: %w)", err)
	}
	return mattermostPayloads, nil
}
