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

package huebridge

import (
	"github.com/tdrn-org/go-hue/api"
)

func brightnessValue(response *api.Brightness) uint8 {
	if response == nil {
		return 0
	}
	value := uint8(*response * 100)
	if value > 100 {
		value = 100
	}
	return value
}
