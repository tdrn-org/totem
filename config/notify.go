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

package config

import (
	"fmt"

	"github.com/gorhill/cronexpr"
)

type NotifyConfig struct {
	MattermostURL         string                 `toml:"mattermost_url"`
	MattermostToken       string                 `toml:"mattermost_token"`
	ArrivalNotification   string                 `toml:"arrival_notification"`
	DepartureNotification string                 `toml:"departure_notification"`
	PulseNotification     string                 `toml:"pulse_notification"`
	Profiles              []NotifyProfileConfig  `toml:"profile"`
	Schedules             []NotifyScheduleConfig `toml:"schedule"`
}

type NotifyProfileConfig struct {
	Agent    string `toml:"agent"`
	Presence bool   `toml:"presence"`
}

type NotifyScheduleConfig struct {
	Agent        string       `toml:"agent"`
	Schedule     ScheduleSpec `toml:"schedule"`
	Notification string       `toml:"notification"`
}

type ScheduleSpec struct {
	*cronexpr.Expression
}

func (spec *ScheduleSpec) Value() string {
	if spec.Expression == nil {
		return ""
	}
	return spec.Value()
}

func (spec *ScheduleSpec) MarshalTOML() ([]byte, error) {
	return []byte(`"` + spec.Value() + `"`), nil
}

func (spec *ScheduleSpec) UnmarshalTOML(value any) error {
	expressionString, ok := value.(string)
	if !ok {
		return fmt.Errorf("unexpected Cron expression type %v", value)
	}
	if expressionString == "" {
		return nil
	}
	parsedExpression, err := cronexpr.Parse(expressionString)
	if err != nil {
		return fmt.Errorf("invalid Cron expression: '%s' (cause: %w)", expressionString, err)
	}
	spec.Expression = parsedExpression
	return nil
}
