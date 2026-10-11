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

	"github.com/tdrn-org/totem/config"
)

func Send(ctx context.Context, cfg *config.NotifyConfig, message string) error {
	messageCfg, ok := cfg.Messages[message]
	if !ok {
		return fmt.Errorf("unknown message %q", message)
	}
	mattermostPayloads, err := newPayloadFactory(cfg)
	if err != nil {
		return err
	}
	payload := mattermostPayloads.NewPayload(messageCfg.Message)
	params := &mattermostAgentParams{
		Agent: messageCfg.Agent,
	}
	err = payload.Send(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to send message %q (cause: %w)", message, err)
	}
	return nil
}
