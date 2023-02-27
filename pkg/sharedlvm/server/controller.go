package server

import (
	"context"
	"fmt"
	"strings"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/phillipleblanc/sharedlvm/pkg/sharedlvm"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog"
)

type controller struct{}

func NewController() csi.ControllerServer {
	return &controller{}
}

func (cs *controller) CreateVolume(
	ctx context.Context,
	req *csi.CreateVolumeRequest,
) (*csi.CreateVolumeResponse, error) {
	klog.Infof("CreateVolume: %v", req)

	name := req.Name
	capacityBytes := req.CapacityRange.GetRequiredBytes()
	volumeGroup, ok := req.Parameters["volumeGroup"]
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "volumeGroup parameter is required")
	}

	if err := sharedlvm.ValidateName(name); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid volume name: %s", err.Error())
	}

	if err := sharedlvm.ValidateName(volumeGroup); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid volume group: %s", err.Error())
	}

	err := sharedlvm.ActivateVolumeGroupLock(volumeGroup)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("Failed to activate volume group %s: %s", volumeGroup, err.Error()))
	}

	err = sharedlvm.CreateVolumeIfNotExists(name, volumeGroup, capacityBytes)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("Failed to create volume %s: %s", name, err.Error()))
	}

	err = sharedlvm.DeactivateVolume(name, volumeGroup)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("Failed to deactivate volume %s: %s", name, err.Error()))
	}

	return &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			VolumeId:      sharedlvm.GetVolumeId(name, volumeGroup),
			CapacityBytes: capacityBytes,
			VolumeContext: req.Parameters,
		},
	}, nil
}

func (cs *controller) DeleteVolume(
	ctx context.Context,
	req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	klog.Infof("DeleteVolume: %v", req)
	// Intentionally not deleting volumes in this first iteration
	return &csi.DeleteVolumeResponse{}, nil
}

func (cs *controller) ValidateVolumeCapabilities(
	ctx context.Context,
	req *csi.ValidateVolumeCapabilitiesRequest,
) (*csi.ValidateVolumeCapabilitiesResponse, error) {
	klog.Infof("ValidateVolumeCapabilities: %v", req)
	volumeID := strings.ToLower(req.GetVolumeId())
	if len(volumeID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID not provided")
	}
	volCaps := req.GetVolumeCapabilities()
	if len(volCaps) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume capabilities not provided")
	}

	confirmedVolCaps := []*csi.VolumeCapability{}

	for _, volCap := range volCaps {
		var supportedAccessType, supportedAccessMode bool
		switch volCap.GetAccessType().(type) {
		case *csi.VolumeCapability_Mount:
			supportedAccessType = true
		default:
		}

		switch volCap.GetAccessMode().GetMode() {
		case csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER, csi.VolumeCapability_AccessMode_SINGLE_NODE_SINGLE_WRITER:
			supportedAccessMode = true
		default:
		}

		if supportedAccessMode && supportedAccessType {
			confirmedVolCaps = append(confirmedVolCaps, volCap)
		}
	}

	confirmed := &csi.ValidateVolumeCapabilitiesResponse_Confirmed{VolumeCapabilities: confirmedVolCaps}
	return &csi.ValidateVolumeCapabilitiesResponse{
		Confirmed: confirmed,
	}, nil
}

func newControllerCapabilities() []*csi.ControllerServiceCapability {
	fromType := func(
		cap csi.ControllerServiceCapability_RPC_Type,
	) *csi.ControllerServiceCapability {
		return &csi.ControllerServiceCapability{
			Type: &csi.ControllerServiceCapability_Rpc{
				Rpc: &csi.ControllerServiceCapability_RPC{
					Type: cap,
				},
			},
		}
	}

	var capabilities []*csi.ControllerServiceCapability
	for _, cap := range []csi.ControllerServiceCapability_RPC_Type{
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
	} {
		capabilities = append(capabilities, fromType(cap))
	}
	return capabilities
}

func (cs *controller) ControllerGetCapabilities(
	ctx context.Context,
	req *csi.ControllerGetCapabilitiesRequest,
) (*csi.ControllerGetCapabilitiesResponse, error) {
	klog.Info("ControllerGetCapabilities")

	resp := &csi.ControllerGetCapabilitiesResponse{
		Capabilities: newControllerCapabilities(),
	}

	return resp, nil
}

func (cs *controller) ControllerGetVolume(
	ctx context.Context,
	req *csi.ControllerGetVolumeRequest,
) (*csi.ControllerGetVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (cs *controller) GetCapacity(
	ctx context.Context,
	req *csi.GetCapacityRequest,
) (*csi.GetCapacityResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (cs *controller) ControllerExpandVolume(
	ctx context.Context,
	req *csi.ControllerExpandVolumeRequest,
) (*csi.ControllerExpandVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (cs *controller) CreateSnapshot(
	ctx context.Context,
	req *csi.CreateSnapshotRequest,
) (*csi.CreateSnapshotResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (cs *controller) DeleteSnapshot(
	ctx context.Context,
	req *csi.DeleteSnapshotRequest,
) (*csi.DeleteSnapshotResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (cs *controller) ListSnapshots(
	ctx context.Context,
	req *csi.ListSnapshotsRequest,
) (*csi.ListSnapshotsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (cs *controller) ControllerUnpublishVolume(
	ctx context.Context,
	req *csi.ControllerUnpublishVolumeRequest,
) (*csi.ControllerUnpublishVolumeResponse, error) {

	return nil, status.Error(codes.Unimplemented, "")
}

func (cs *controller) ControllerPublishVolume(
	ctx context.Context,
	req *csi.ControllerPublishVolumeRequest,
) (*csi.ControllerPublishVolumeResponse, error) {

	return nil, status.Error(codes.Unimplemented, "")
}

func (cs *controller) ListVolumes(
	ctx context.Context,
	req *csi.ListVolumesRequest,
) (*csi.ListVolumesResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}
