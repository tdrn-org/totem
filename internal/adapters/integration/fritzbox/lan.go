package fritzbox

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/netip"

	"github.com/tdrn-org/go-tr064"
	"github.com/tdrn-org/go-tr064/services/tr64desc/hosts"
	"github.com/tdrn-org/totem/internal/domain/lan"
)

func (s *fritzboxService) ListDevices(ctx context.Context) ([]*lan.LanDevice, error) {
	serviceClient := s.hostsClient()
	ghnoeOut := &hosts.GetHostNumberOfEntriesResponse{}
	err := serviceClient.GetHostNumberOfEntries(ghnoeOut)
	if err != nil {
		return nil, fmt.Errorf("failed to get host count from FRITZ!Box (cause: %w)", err)
	}
	entryCount := int(ghnoeOut.NewHostNumberOfEntries)
	devices := make([]*lan.LanDevice, 0, entryCount)
	for entryIndex := 0; entryIndex < entryCount; entryIndex++ {
		ggheIn := &hosts.GetGenericHostEntryRequest{NewIndex: uint16(entryIndex)}
		ggheOut := &hosts.GetGenericHostEntryResponse{}
		err = serviceClient.GetGenericHostEntry(ggheIn, ggheOut)
		if err != nil {
			return nil, fmt.Errorf("failed to get host entry from FRITZ!Box (cause: %w)", err)
		}
		macString := ggheOut.NewMACAddress
		if macString == "" {
			continue
		}
		mac, err := net.ParseMAC(macString)
		if err != nil {
			s.logger.Warn("unable to parse host hardware address; ignoring host", slog.String("mac", macString))
			continue
		}
		ipString := ggheOut.NewIPAddress
		ip, _ := netip.ParseAddr(ipString)
		device := &lan.LanDevice{
			MAC:      mac,
			Name:     ggheOut.NewHostName,
			Active:   ggheOut.NewActive,
			IP:       ip,
			Location: s.host.ResolveLocationID(s.config.LocationID()),
		}
		devices = append(devices, device)
	}
	return devices, nil
}

func (s *fritzboxService) hostsClient() *hosts.ServiceClient {
	return &hosts.ServiceClient{
		TR064Client: s.tr064Client,
		Service: &tr064.StaticServiceDescriptor{
			ServiceSpec:       tr064.ServiceSpec("tr64desc"),
			ServiceType:       "urn:dslforum-org:service:Hosts:1",
			ServiceId:         "urn:LanDeviceHosts-com:serviceId:Hosts1",
			ServiceControlUrl: "/upnp/control/hosts",
		},
	}
}
