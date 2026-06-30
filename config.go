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

package totem

import (
	"log/slog"
	"net/url"
	"reflect"
	"time"

	"github.com/rs/cors"
	"github.com/tdrn-org/go-conf/service/loglevel"
	"github.com/tdrn-org/go-httpserver"
	"github.com/tdrn-org/go-httpserver/certificate"
	"github.com/tdrn-org/go-log"
	"github.com/tdrn-org/totem/config"
	"github.com/tdrn-org/totem/internal/adapters/integration"
	"github.com/tdrn-org/totem/internal/adapters/integration/fritzbox"
	"github.com/tdrn-org/totem/internal/adapters/integration/huebridge"
	"github.com/tdrn-org/totem/internal/adapters/integration/jellyfin"
	"github.com/tdrn-org/totem/internal/domain"
)

func applyLoggingConfig(cfg *config.LoggingConfig) {
	logConfig := &log.Config{
		Level:          cfg.Level.Value(),
		AddSource:      false,
		Target:         log.Target(cfg.Target),
		Color:          log.Color(cfg.Color),
		FileName:       cfg.FileName,
		FileSizeLimit:  cfg.FileSizeLimit,
		SyslogNetwork:  cfg.SyslogNetwork,
		SyslogAddress:  cfg.SyslogAddress,
		SyslogEncoding: cfg.SyslogEncoding,
		SyslogFacility: cfg.SyslogFacility,
		SyslogAppName:  reflect.TypeFor[Server]().PkgPath(),
	}
	logger, _ := logConfig.GetLogger(loglevel.LevelVar())
	slog.SetDefault(logger)
}

func httpServerOptions(cfg *config.ServerConfig) []httpserver.OptionSetter {
	httpServerOptions := make([]httpserver.OptionSetter, 0)
	// TLS
	if cfg.Protocol == config.ServerProtocolHttps {
		certificateProvider := &certificate.FileCertificateProvider{
			CertFile: cfg.CertFile,
			KeyFile:  cfg.KeyFile,
		}
		httpServerOptions = append(httpServerOptions, httpserver.WithCertificateProvider(certificateProvider))
	}
	// Proxy configuration
	if len(cfg.TrustedProxies) > 0 {
		httpServerOptions = append(httpServerOptions, httpserver.WithTrustedProxyPolicy(httpserver.AllowNetworks("trusted proxies", cfg.TrustedProxies.Prefixes())))
	}
	if len(cfg.TrustedHeaders) > 0 {
		httpServerOptions = append(httpServerOptions, httpserver.WithTrustedHeaders(cfg.TrustedHeaders...))
	}
	// CORS
	if len(cfg.AllowedOrigins) > 0 {
		corsOptions := &cors.Options{
			AllowedOrigins: cfg.AllowedOrigins,
		}
		httpServerOptions = append(httpServerOptions, httpserver.WithCorsOptions(corsOptions))
	}
	// Access log
	var accessLogConfig *log.Config
	switch cfg.AccessLog {
	case "stdout":
		accessLogConfig = &log.Config{
			Target: log.TargetStdout,
		}
	case "stderr":
		accessLogConfig = &log.Config{
			Target: log.TargetStderr,
		}
	case "":
		// disable Access log
	default:
		accessLogConfig = &log.Config{
			Target:        log.TargetFileText,
			FileName:      cfg.AccessLog,
			FileSizeLimit: cfg.AccessLogSizeLimit,
		}
	}
	if accessLogConfig != nil {
		accessLogLogger := slog.New(log.NewRawHandler(accessLogConfig.GetWriter()))
		httpServerOptions = append(httpServerOptions, httpserver.WithAccessLog(accessLogLogger))
	}
	return httpServerOptions
}

func (s *Server) registerService(service integration.Service) {
	serviceID := service.ID()
	s.logger.Info("registering service", slog.String("serviceID", serviceID))
	_, exists := s.services[serviceID]
	if exists {
		s.logger.Info("re-registering service", slog.String("serviceID", serviceID))
	}
	s.services[serviceID] = service
}

var defaultLocation *domain.Location = &domain.Location{
	Name:      "Default",
	Latitude:  0.0,
	Longitude: 0.0,
}

func (s *Server) resolveLocationID(locationID string) (*domain.Location, bool) {
	if locationID == "" {
		return defaultLocation, false
	}
	for _, locationConfig := range s.cfg.Locations {
		if locationConfig.ID == locationID && locationConfig.Name != "" {
			location := &domain.Location{
				ID:        locationConfig.ID,
				Name:      locationConfig.Name,
				Address:   locationConfig.Address,
				Latitude:  locationConfig.Latitude,
				Longitude: locationConfig.Longitude,
			}
			return location, true
		}
	}
	return defaultLocation, false
}

func (s *Server) bindIntegrationServices(cfg *config.Config) {
	for _, fritzboxConfig := range cfg.Services.FritzboxServices {
		if fritzboxConfig.Enabled {
			s.bindFritzboxService(&fritzboxConfig)
		}
	}
	for _, huebridgeConfig := range cfg.Services.HuebridgeServices {
		if huebridgeConfig.Enabled {
			s.bindHuebridgeService(&huebridgeConfig)
		}
	}
	for _, jellyfinConfig := range cfg.Services.JellyfinServices {
		if jellyfinConfig.Enabled {
			s.bindJellyfinService(&jellyfinConfig)
		}
	}
}

func (s *Server) bindFritzboxService(cfg *config.FritzboxConfig) {
	fritzbox.Register(s.runtime(), &fritzboxConfig{cfg: cfg})
}

type fritzboxConfig struct {
	cfg *config.FritzboxConfig
}

func (c *fritzboxConfig) LocationID() string {
	return c.cfg.LocationID
}

func (c *fritzboxConfig) DeviceURL() *url.URL {
	return c.cfg.DeviceURL.URL
}

func (s *Server) bindHuebridgeService(cfg *config.HuebridgeConfig) {
	huebridge.Register(s.runtime(), &huebridgeConfig{cfg: cfg})
}

type huebridgeConfig struct {
	cfg *config.HuebridgeConfig
}

func (c *huebridgeConfig) LocationID() string {
	return c.cfg.LocationID
}

func (c *huebridgeConfig) Address() string {
	return c.cfg.Address
}

func (c *huebridgeConfig) BridgeID() string {
	return c.cfg.BridgeID
}

func (c *huebridgeConfig) Login() string {
	return c.cfg.Login
}

const defaultHuebridgeTimeout time.Duration = 30 * time.Second

func (c *huebridgeConfig) Timeout() time.Duration {
	timeout := time.Duration(c.cfg.Timeout)
	if timeout <= 0 {
		return defaultHuebridgeTimeout
	}
	return timeout
}

func (s *Server) bindJellyfinService(cfg *config.JellyfinConfig) {
	jellyfin.Register(s.runtime(), &jellyfinConfig{cfg: cfg})
}

type jellyfinConfig struct {
	cfg *config.JellyfinConfig
}

func (c *jellyfinConfig) APIURL() *url.URL {
	return c.cfg.APIURL.URL
}

func (c *jellyfinConfig) APIKey() string {
	return c.cfg.APIKey
}

func (c *jellyfinConfig) UserID() string {
	return c.cfg.UserID
}
