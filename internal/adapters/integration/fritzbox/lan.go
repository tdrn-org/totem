package fritzbox

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	"github.com/tdrn-org/go-tr064"
	"github.com/tdrn-org/go-tr064/services/tr64desc/hosts"
	"github.com/tdrn-org/totem/internal/domain/lan"
)

func (s *fritzboxService) ListDevices(ctx context.Context) ([]*lan.LanDevice, error) {
	serviceClient := s.hostsClient()
	hostNumberOfEntriesResponse := &hosts.GetHostNumberOfEntriesResponse{}
	err := serviceClient.GetHostNumberOfEntries(hostNumberOfEntriesResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to get host count from FRITZ!Box (cause: %w)", err)
	}
	entryCount := hostNumberOfEntriesResponse.NewHostNumberOfEntries
	devices := make([]*lan.LanDevice, 0, int(entryCount))
	for entryIndex := uint16(0); entryIndex < entryCount; entryIndex++ {
		genericHostEntryRequest := &hosts.GetGenericHostEntryRequest{NewIndex: entryIndex}
		genericHostEntryResponse := &hosts.GetGenericHostEntryResponse{}
		err = serviceClient.GetGenericHostEntry(genericHostEntryRequest, genericHostEntryResponse)
		if err != nil {
			return nil, fmt.Errorf("failed to get host entry from FRITZ!Box (cause: %w)", err)
		}
		macString := genericHostEntryResponse.NewMACAddress
		mac, err := net.ParseMAC(macString)
		if err != nil {
			s.logger.Warn("unable to parse host hardware address; ignoring host", slog.String("mac", macString))
			continue
		}
		device := &lan.LanDevice{
			MAC:      mac,
			Name:     genericHostEntryResponse.NewHostName,
			Active:   genericHostEntryResponse.NewActive,
			IP:       genericHostEntryResponse.NewIPAddress,
			Location: s.host.ResolveLocationID(s.config.LocationID()),
		}
		devices = append(devices, device)
	}
	return devices, nil
}

func (s *fritzboxService) GetDevice(ctx context.Context, mac net.HardwareAddr) (*lan.LanDevice, error) {
	return nil, nil
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
