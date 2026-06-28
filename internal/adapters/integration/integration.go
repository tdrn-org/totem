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

// Package integration defines the interaction between different services
// (e.g. smarthome lights) and the host executing them.
// While the [Host] interface defines the host related functions, used
// be the services to interact with their runtime environment, the
// [Service] interface covers the service related functions.
// A service is bound to the host during system start via a
// Registration call.
package integration

import (
	"context"
	"fmt"
)

// Host interface defines the functions needed by the services to interact
// with the host running them.
type Host interface {
	// Register makes the given service visible to the host und confirms
	// it is ready to be used.
	Register(service Service)
}

type Service interface {
	// Name gets the name of the service (e.g. huebridge). Multiple services
	// with the same name may exist at the same time.
	Name() string
	// ID gets the unique id of the service. This ID is build up from the
	// service name and an unique instance suffix (see [NewServiceID]).
	ID() string
	// Shutdown shuts down the service gracefully processing any pending
	// requests before comencing operation.
	Shutdown(ctx context.Context) error
	// Close terminates the service immediately releasing any allocated
	// resources.
	Close() error
}

// NewServiceID creates a unique service id as returned by [Service.ID].
// A service id consists of the service name and an instance suffix
// unique for the service. Ensuring the latter is in the responsibility
// of the service implementation.
func NewServiceID(name, instance string) string {
	return fmt.Sprintf("%s[%s]", name, instance)
}
