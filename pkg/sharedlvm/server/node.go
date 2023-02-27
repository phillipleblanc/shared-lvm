package server

import (
	"context"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/phillipleblanc/sharedlvm/pkg/sharedlvm"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog"
)

type node struct {
	nodeId string
}

func NewNode(nodeId string) csi.NodeServer {
	return &node{
		nodeId: nodeId,
	}
}

func (ns *node) NodePublishVolume(
	ctx context.Context,
	req *csi.NodePublishVolumeRequest,
) (*csi.NodePublishVolumeResponse, error) {
	klog.Infof("NodePublishVolume: %v", req)
	readOnly := req.GetReadonly()
	targetPath := req.GetTargetPath()
	volumeId := req.GetVolumeId()
	volumeName, volumeGroup := sharedlvm.GetVolumeNameAndGroup(volumeId)
	fsType := req.GetVolumeCapability().GetMount().GetFsType()

	mountOptions := req.GetVolumeCapability().GetMount().GetMountFlags()
	if readOnly {
		mountOptions = append(mountOptions, "ro")
	}

	err := sharedlvm.ActivateVolumeGroupLock(volumeGroup)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	err = sharedlvm.ActivateVolume(volumeName, volumeGroup)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	err = sharedlvm.MountFilesystem(volumeName, volumeGroup, targetPath, fsType, mountOptions)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &csi.NodePublishVolumeResponse{}, nil
}

func (ns *node) NodeUnpublishVolume(
	ctx context.Context,
	req *csi.NodeUnpublishVolumeRequest,
) (*csi.NodeUnpublishVolumeResponse, error) {
	klog.Infof("NodeUnpublishVolume: %v", req)
	targetPath := req.GetTargetPath()
	volumeId := req.GetVolumeId()
	volumeName, volumeGroup := sharedlvm.GetVolumeNameAndGroup(volumeId)

	err := sharedlvm.UnmountFilesystem(targetPath)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	err = sharedlvm.DeactivateVolume(volumeName, volumeGroup)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &csi.NodeUnpublishVolumeResponse{}, nil
}

func (ns *node) NodeGetInfo(
	ctx context.Context,
	req *csi.NodeGetInfoRequest,
) (*csi.NodeGetInfoResponse, error) {
	resp := &csi.NodeGetInfoResponse{
		NodeId: ns.nodeId,
	}

	klog.Infof("NodeGetInfo: %v", resp)

	return resp, nil
}

func (ns *node) NodeGetCapabilities(
	ctx context.Context,
	req *csi.NodeGetCapabilitiesRequest,
) (*csi.NodeGetCapabilitiesResponse, error) {
	resp := &csi.NodeGetCapabilitiesResponse{
		Capabilities: []*csi.NodeServiceCapability{},
	}

	klog.Infof("NodeGetCapabilities: %v", resp)

	return resp, nil
}

func (ns *node) NodeStageVolume(
	ctx context.Context,
	req *csi.NodeStageVolumeRequest,
) (*csi.NodeStageVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (ns *node) NodeUnstageVolume(
	ctx context.Context,
	req *csi.NodeUnstageVolumeRequest,
) (*csi.NodeUnstageVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (ns *node) NodeExpandVolume(
	ctx context.Context,
	req *csi.NodeExpandVolumeRequest,
) (*csi.NodeExpandVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (ns *node) NodeGetVolumeStats(
	ctx context.Context,
	req *csi.NodeGetVolumeStatsRequest,
) (*csi.NodeGetVolumeStatsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}
