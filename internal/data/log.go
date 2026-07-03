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

package data

import (
	"context"
	"net/netip"
	"time"

	"github.com/tdrn-org/totem/internal/data/model"
	"github.com/tdrn-org/totem/internal/domain"
)

type UserPresenceLog struct {
	User      *domain.User
	Location  *domain.Location
	Active    bool
	IP        netip.Addr
	Name      string
	Timestamp time.Time
}

func (s *Store) LogUserPresence(ctx context.Context, user *domain.User, location *domain.Location, active bool, ip netip.Addr, name string) (*UserPresenceLog, error) {
	storeLog, err := model.LogPresence(ctx, s.driver, user.ID, location.ID, active, ip, name)
	if err != nil {
		return nil, err
	}
	return s.model2UserPresenceLog(user, storeLog), nil
}

func (s *Store) ListUserPresenceLogs(ctx context.Context, user *domain.User, lastOnly bool) ([]*UserPresenceLog, error) {
	storeLogs, err := model.SelectPresenceLogByUserID(ctx, s.driver, user.ID)
	if err != nil {
		return nil, err
	}
	locationMap := make(map[string]string, len(storeLogs))
	logs := make([]*UserPresenceLog, 0, len(storeLogs))
	for _, storeLog := range storeLogs {
		if lastOnly {
			_, exists := locationMap[storeLog.LocationID]
			if exists {
				continue
			}
		}
		locationMap[storeLog.LocationID] = storeLog.LocationID
		logs = append(logs, s.model2UserPresenceLog(user, storeLog))
	}
	return logs, nil
}

func (s *Store) ListLocationPresenceLogs(ctx context.Context, location *domain.Location) ([]*UserPresenceLog, error) {
	storeLogs, err := model.SelectPresenceLogByLocationID(ctx, s.driver, location.ID)
	if err != nil {
		return nil, err
	}
	userMap := make(map[string]string, len(storeLogs))
	logs := make([]*UserPresenceLog, 0, len(storeLogs))
	for _, storeLog := range storeLogs {
		_, exists := userMap[storeLog.UserID]
		if exists {
			continue
		}
		userMap[storeLog.UserID] = storeLog.UserID
		user, err := s.GetUser(ctx, storeLog.UserID)
		if err != nil {
			return nil, err
		}
		logs = append(logs, s.model2UserPresenceLog(user, storeLog))
	}
	return logs, nil
}

func (s *Store) model2UserPresenceLog(user *domain.User, storeLog *model.PresenceLog) *UserPresenceLog {
	ip := netip.Addr{}
	if storeLog.IP != "" {
		ip, _ = netip.ParseAddr(storeLog.IP)
	}
	log := &UserPresenceLog{
		User:     user,
		Location: s.runtime.ResolveLocationID(storeLog.LocationID),
		Active:   storeLog.Active,
		IP:       ip,
		Name:     storeLog.Name,
	}
	return log
}
