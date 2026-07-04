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

package smarthome

import (
	"context"

	"github.com/tdrn-org/totem/internal/domain"
)

// Light represents a single smart light.
type Light struct {
	ID         string
	Name       string
	On         bool
	Brightness uint8
	Location   *domain.Location
}

// TemperatureSensor represents a sensor that measures temperature.
type TemperatureSensor struct {
	ID          string
	Name        string
	Type        string
	Temperature float64
	Location    *domain.Location
}

// BatteryDevice represents a battery-powered device with a charge level.
type BatteryDevice struct {
	ID           string
	Name         string
	Type         string
	BatteryLevel int // Battery percentage 0-100
	Location     *domain.Location
}

// MotionSensor represents a motion detector.
type MotionSensor struct {
	ID       string
	Name     string
	Motion   bool
	Location *domain.Location
}

// SmarthomeService defines the operations for smart-home lighting and sensors.
type SmarthomeService interface {
	Location() *domain.Location
	ListLights(ctx context.Context) ([]*Light, error)
	GetLight(ctx context.Context, id string) (*Light, error)
	SetLight(ctx context.Context, id string, on *bool, brightness *uint8) (*Light, error)
	ListTemperatureSensors(ctx context.Context) ([]*TemperatureSensor, error)
	GetTemperatureSensor(ctx context.Context, id string) (*TemperatureSensor, error)
	ListBatteryDevices(ctx context.Context) ([]*BatteryDevice, error)
	GetBatteryDevice(ctx context.Context, id string) (*BatteryDevice, error)
	ListMotionSensors(ctx context.Context) ([]*MotionSensor, error)
	GetMotion(ctx context.Context, id string) (*MotionSensor, error)
}
