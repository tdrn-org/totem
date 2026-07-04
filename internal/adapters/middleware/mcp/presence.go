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

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tdrn-org/totem/internal/domain"
	"github.com/tdrn-org/totem/internal/domain/presence"
)

func registerPresenceTools(server *mcp.Server, runtime Runtime) {
	presenceTracker := runtime.PresenceTracker()
	if presenceTracker == nil {
		return
	}

	// checkUser: find a user's current and last known location
	mcp.AddTool(server, &mcp.Tool{
		Name:        "checkUser",
		Description: "Determine the current location of a user by checking whether any of their registered devices are active on a known LAN. Returns the location if found, along with the last seen location and timestamp.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"userID": map[string]any{"type": "string", "description": "The user's unique identifier."},
			},
			"required": []string{"userID"},
		},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input struct {
		UserID string `json:"userID"`
	}) (*mcp.CallToolResult, any, error) {
		user, err := runtime.DataStore().GetUser(ctx, input.UserID)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get user '%s': %w", input.UserID, err)
		}
		if user == nil {
			return nil, nil, fmt.Errorf("user '%s' not found", input.UserID)
		}
		presence, err := presenceTracker.CheckUser(ctx, user)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to check user presence: %w", err)
		}
		text := formatUserPresenceText(user, presence)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: text}},
		}, map[string]any{"user": user, "presence": presence}, nil
	})

	// checkLocation: list users and triggered motion sensors at a location
	mcp.AddTool(server, &mcp.Tool{
		Name:        "checkLocation",
		Description: "List all users currently present at a given location, along with any motion sensors that are currently triggered. Use location IDs from the server configuration.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"locationID": map[string]any{"type": "string", "description": "The location's unique identifier (e.g. 'home1')."},
			},
			"required": []string{"locationID"},
		},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input struct {
		LocationID string `json:"locationID"`
	}) (*mcp.CallToolResult, any, error) {
		location := runtime.ResolveLocationID(input.LocationID)
		result, err := presenceTracker.CheckLocation(ctx, location)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to check location presence: %w", err)
		}
		text := formatPresenceResultText(location, result)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: text}},
		}, map[string]any{"presenceResult": result}, nil
	})
}

func formatUserPresenceText(user *domain.User, p *presence.UserPresence) string {
	lastSeen := "unknown"
	if p.LastSeen != nil && !p.LastSeen.IsZero() {
		lastSeen = p.LastSeen.Format("2006-01-02 15:04:05")
	}
	if p.Location != nil {
		return fmt.Sprintf("User '%s' is at '%s' (%s).\nLast seen: %s at '%s'.",
			user.Name, p.Location.Name, p.Location.Address,
			lastSeen, p.Location.Name)
	}
	if p.LastSeenLocation != nil {
		return fmt.Sprintf("User '%s' is not currently present at any known location.\nLast seen: %s at '%s' (%s).",
			user.Name, lastSeen,
			p.LastSeenLocation.Name, p.LastSeenLocation.Address)
	}
	return fmt.Sprintf("User '%s' has never been seen at any known location.", user.Name)
}

func formatPresenceResultText(location *domain.Location, result *presence.PresenceResult) string {
	text := fmt.Sprintf("Location '%s' (%s):\n", location.Name, location.Address)
	if len(result.Users) == 0 {
		text += "  No users present.\n"
	} else {
		text += fmt.Sprintf("  %d user(s) present:\n", len(result.Users))
		for _, user := range result.Users {
			text += fmt.Sprintf("    - %s (%s)\n", user.Name, user.ID)
		}
	}
	if len(result.MotionSensors) == 0 {
		text += "  No motion sensors triggered.\n"
	} else {
		text += fmt.Sprintf("  %d motion sensor(s) triggered:\n", len(result.MotionSensors))
		for _, sensor := range result.MotionSensors {
			text += fmt.Sprintf("    - %s (%s)\n", sensor.Name, sensor.ID)
		}
	}
	return text
}
