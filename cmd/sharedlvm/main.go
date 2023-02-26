package main

import (
	"flag"
	"net"
	"os"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/phillipleblanc/shared-lvm/pkg/config"
	"github.com/phillipleblanc/shared-lvm/pkg/sharedlvm"
	"google.golang.org/grpc"
	"k8s.io/klog"
)

func main() {
	cfg := config.Config{}
	flag.StringVar(&cfg.Endpoint, "endpoint", "/csi/csi.sock", "CSI endpoint")
	flag.StringVar(&cfg.NodeId, "nodeid", "spice-node-1", "Node ID")

	flag.Parse()

	if err := os.Remove(cfg.Endpoint); err != nil && !os.IsNotExist(err) {
		klog.Fatalf("Failed to remove %s, error: %s", cfg.Endpoint, err.Error())
	}

	listener, err := net.Listen("unix", cfg.Endpoint)
	if err != nil {
		klog.Fatalf("failed to listen: %v", err)
	}
	defer listener.Close()

	server := grpc.NewServer()

	csi.RegisterIdentityServer(server, sharedlvm.NewIdentity())
	csi.RegisterControllerServer(server, sharedlvm.NewController())
	csi.RegisterNodeServer(server, sharedlvm.NewNode(cfg.NodeId))

	err = server.Serve(listener)
	if err != nil {
		klog.Fatalf("failed to start server: %v", err)
	}
}
