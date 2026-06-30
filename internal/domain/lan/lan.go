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

package lan

import (
	"context"
	"net"
	"net/netip"

	"github.com/tdrn-org/totem/internal/domain"
)

type LanDevice struct {
	MAC      net.HardwareAddr // Normalisiert im Adapter
	Active   bool
	Name     string
	IP       netip.Addr
	Location *domain.Location
}

type LanService interface {
	ListDevices(ctx context.Context) ([]*LanDevice, error)
	GetDevice(ctx context.Context, mac net.HardwareAddr) (*LanDevice, error)
}
