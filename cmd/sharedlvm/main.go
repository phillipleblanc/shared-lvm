package main

import (
	"flag"
	"net"
	"os"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/phillipleblanc/sharedlvm/pkg/config"
	lvmserver "github.com/phillipleblanc/sharedlvm/pkg/sharedlvm/server"
	"google.golang.org/grpc"
	"k8s.io/klog"
)

func main() {
	cfg := config.Config{}
	flag.StringVar(&cfg.Endpoint, "endpoint", "/csi/csi.sock", "CSI endpoint")
	flag.StringVar(&cfg.ServerType, "servertype", "controller", "Server type (controller or node)")
	flag.StringVar(&cfg.NodeId, "nodeid", "", "Node ID (required for node server)")

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

	csi.RegisterIdentityServer(server, lvmserver.NewIdentity())

	if cfg.ServerType == "controller" {
		csi.RegisterControllerServer(server, lvmserver.NewController())
	}

	if cfg.ServerType == "node" && cfg.NodeId == "" {
		klog.Fatalf("nodeid is required for node server")
	}

	if cfg.ServerType == "node" {
		csi.RegisterNodeServer(server, lvmserver.NewNode(cfg.NodeId))
	}

	err = server.Serve(listener)
	if err != nil {
		klog.Fatalf("failed to start server: %v", err)
	}
}
