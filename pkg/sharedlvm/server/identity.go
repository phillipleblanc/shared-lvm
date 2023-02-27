package server

import (
	"context"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"k8s.io/klog"
)

const (
	PluginName    = "sharedlvm.csi.leblanc.tech"
	PluginVersion = "v0.0.1+alpha.01"
)

type identity struct{}

func NewIdentity() csi.IdentityServer {
	return &identity{}
}

// GetPluginInfo returns the version and name of
// this service
//
// This implements csi.IdentityServer
func (id *identity) GetPluginInfo(
	ctx context.Context,
	req *csi.GetPluginInfoRequest,
) (*csi.GetPluginInfoResponse, error) {
	resp := &csi.GetPluginInfoResponse{
		Name:          PluginName,
		VendorVersion: PluginVersion,
	}

	klog.Infof("GetPluginInfo: %v", resp)

	return resp, nil
}

// Probe checks if the plugin is running or not
func (id *identity) Probe(
	ctx context.Context,
	req *csi.ProbeRequest,
) (*csi.ProbeResponse, error) {
	return &csi.ProbeResponse{
		Ready: wrapperspb.Bool(true),
	}, nil
}

// GetPluginCapabilities returns supported capabilities
// of this plugin
//
// Currently it reports whether this plugin can serve
// the Controller interface. Controller interface methods
// are called dependant on this
func (id *identity) GetPluginCapabilities(
	ctx context.Context,
	req *csi.GetPluginCapabilitiesRequest,
) (*csi.GetPluginCapabilitiesResponse, error) {
	resp := &csi.GetPluginCapabilitiesResponse{
		Capabilities: []*csi.PluginCapability{
			{
				Type: &csi.PluginCapability_Service_{
					Service: &csi.PluginCapability_Service{
						Type: csi.PluginCapability_Service_CONTROLLER_SERVICE,
					},
				},
			},
		},
	}

	klog.Infof("GetPluginCapabilities: %v", resp)

	return resp, nil
}
