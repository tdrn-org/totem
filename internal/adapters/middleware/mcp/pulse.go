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
	"context"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tdrn-org/totem/internal/domain/pulse"
)

func registerPulseTools(server *mcp.Server, runtime Runtime) {
	// echo_pulse: schedule a self-addressed delayed notification
	mcp.AddTool(server, &mcp.Tool{
		Name:        "echo_pulse",
		Description: "Schedules a self-addressed delayed notification. The message will be delivered to the agent's Mattermost DM at the specified time. At most one pending pulse exists per agent — creating a new one overwrites any existing pending pulse.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"schedule": map[string]any{"type": "string", "description": "When to deliver the message, as a Go duration string (e.g. '30m', '2h', '1h30m')."},
				"message":  map[string]any{"type": "string", "description": "The message to deliver. Write it as a self-addressed prompt that your future self will understand (e.g. 'Ist Holger im WLAN? Wenn nein, wie lange schon?')."},
				"agent":    map[string]any{"type": "string", "description": "The agent requesting the pulse (default: 'judy')."},
			},
			"required": []string{"schedule", "message"},
		},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input struct {
		Schedule string `json:"schedule"`
		Message  string `json:"message"`
		Agent    string `json:"agent"`
	}) (*mcp.CallToolResult, any, error) {
		duration, err := time.ParseDuration(input.Schedule)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid schedule %q: %w", input.Schedule, err)
		}
		agent := input.Agent
		if agent == "" {
			agent = "judy"
		}
		now := time.Now()
		p := &pulse.Pulse{
			Agent:       agent,
			Message:     input.Message,
			ScheduledAt: now.Add(duration),
			CreatedAt:   now,
		}
		if err := runtime.DataStore().SavePulse(ctx, p); err != nil {
			return nil, nil, fmt.Errorf("saving pulse: %w", err)
		}
		text := fmt.Sprintf("Pulse scheduled for %s (in %s). Message: %q",
			p.ScheduledAt.Format(time.RFC3339), duration, input.Message)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: text}},
		}, map[string]any{"pulse": p}, nil
	})

	// echo_pulse_cancel: cancel the pending pulse
	mcp.AddTool(server, &mcp.Tool{
		Name:        "echo_pulse_cancel",
		Description: "Cancels the pending echo pulse for an agent, if one exists.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"agent": map[string]any{"type": "string", "description": "The agent whose pulse to cancel (default: 'judy')."},
			},
		},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input struct {
		Agent string `json:"agent"`
	}) (*mcp.CallToolResult, any, error) {
		agent := input.Agent
		if agent == "" {
			agent = "judy"
		}
		pulses, err := runtime.DataStore().ListAgentPulses(ctx, agent)
		if err != nil {
			return nil, nil, fmt.Errorf("listing pulses: %w", err)
		}
		if len(pulses) == 0 {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("No pending pulse for agent %q.", agent)}},
			}, nil, nil
		}
		for _, p := range pulses {
			if err := runtime.DataStore().DeletePulse(ctx, p.ID); err != nil {
				return nil, nil, fmt.Errorf("deleting pulse %q: %w", p.ID, err)
			}
		}
		text := fmt.Sprintf("Cancelled %d pulse(s) for agent %q.", len(pulses), agent)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: text}},
		}, nil, nil
	})
}
