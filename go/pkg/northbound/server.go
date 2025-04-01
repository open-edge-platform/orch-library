// SPDX-FileCopyrightText: (C) 2023 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

// Package northbound houses implementations of various application-oriented interfaces
// for the ONOS configuration subsystem.
package northbound

import (
	"crypto/tls"
	"fmt"
	"github.com/open-edge-platform/orch-library/go/dazl"
	"github.com/open-edge-platform/orch-library/go/pkg/certs"
	"github.com/open-edge-platform/orch-library/go/pkg/grpc/auth"
	"google.golang.org/grpc/credentials"
	"net"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"

	"google.golang.org/grpc"
)

var log = dazl.GetPackageLogger()

// Service provides service-specific registration for grpc services.
type Service interface {
	Register(s *grpc.Server)
}

// Server provides NB gNMI server for onos-config.
type Server struct {
	cfg      *ServerConfig
	services []Service
	server   *grpc.Server
}

// SecurityConfig security configuration
type SecurityConfig struct {
	AuthenticationEnabled bool
	AuthorizationEnabled  bool
}

// ServerConfig comprises a set of server configuration options.
type ServerConfig struct {
	CaPath      *string
	KeyPath     *string
	CertPath    *string
	Port        int16
	Insecure    bool
	SecurityCfg *SecurityConfig
}

// NewServer initializes gNMI server using the supplied configuration.
func NewServer(cfg *ServerConfig) *Server {
	return &Server{
		services: []Service{},
		cfg:      cfg,
	}
}

// NewServerConfig creates a server config created with the specified end-point security details.
// Deprecated: Use NewServerCfg instead
func NewServerConfig(caPath string, keyPath string, certPath string, port int16, insecure bool) *ServerConfig {
	return &ServerConfig{
		Port:        port,
		Insecure:    insecure,
		CaPath:      &caPath,
		KeyPath:     &keyPath,
		CertPath:    &certPath,
		SecurityCfg: &SecurityConfig{},
	}
}

// NewServerCfg creates a server config created with the specified end-point security details.
func NewServerCfg(caPath string, keyPath string, certPath string, port int16, insecure bool, secCfg SecurityConfig) *ServerConfig {
	return &ServerConfig{
		Port:        port,
		Insecure:    insecure,
		CaPath:      &caPath,
		KeyPath:     &keyPath,
		CertPath:    &certPath,
		SecurityCfg: &secCfg,
	}
}

// NewInsecureServerConfig creates an insecure server configuration for the specified port.
func NewInsecureServerConfig(port int16) *ServerConfig {
	return &ServerConfig{
		Port:     port,
		Insecure: true,
		SecurityCfg: &SecurityConfig{
			AuthenticationEnabled: false,
			AuthorizationEnabled:  false,
		},
	}
}

// AddService adds a Service to the server to be registered on Serve.
func (s *Server) AddService(r Service) {
	s.services = append(s.services, r)
}

// Serve starts the NB gNMI server.
func (s *Server) Serve(started func(string), grpcOpts ...grpc.ServerOption) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.cfg.Port))
	if err != nil {
		return err
	}
	tlsCfg := &tls.Config{MinVersion: tls.VersionTLS12}

	if (s.cfg.Insecure && s.cfg.CertPath == nil && s.cfg.KeyPath == nil) ||
		*s.cfg.CertPath == "" && *s.cfg.KeyPath == "" {
		// nothing
		log.Debug("Running in insecure mode")
	} else {
		log.Infof("Loading certs: %s %s", *s.cfg.CertPath, *s.cfg.KeyPath)
		clientCerts, err := tls.LoadX509KeyPair(*s.cfg.CertPath, *s.cfg.KeyPath)
		if err != nil {
			log.Info("Error loading default certs")
		}
		tlsCfg.Certificates = []tls.Certificate{clientCerts}
	}

	if s.cfg.Insecure {
		// RequestClientCert will ask client for a certificate but won't
		// require it to proceed. If certificate is provided, it will be
		// verified.
		tlsCfg.ClientAuth = tls.RequestClientCert
	} else {
		tlsCfg.ClientAuth = tls.RequireAndVerifyClientCert
	}

	if s.cfg.CaPath == nil ||
		*s.cfg.CaPath == "" {
		log.Debug("Running with no CA certificates")
	} else {
		tlsCfg.ClientCAs, err = certs.GetCertPool(*s.cfg.CaPath)
	}

	if err != nil {
		return err
	}

	opts := make([]grpc.ServerOption, 0, 5)
	if len(tlsCfg.Certificates) > 0 {
		opts = append(opts, grpc.Creds(credentials.NewTLS(tlsCfg)))
	}

	if s.cfg.SecurityCfg.AuthenticationEnabled {
		log.Info("Authentication Enabled")
		opts = append(opts, grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(
				grpc_auth.UnaryServerInterceptor(auth.AuthenticationInterceptor),
			)))
		opts = append(opts, grpc.StreamInterceptor(
			grpc_middleware.ChainStreamServer(
				grpc_auth.StreamServerInterceptor(auth.AuthenticationInterceptor))))

	}

	opts = append(opts, grpcOpts...)

	s.server = grpc.NewServer(opts...)
	for i := range s.services {
		s.services[i].Register(s.server)
	}
	started(lis.Addr().String())

	log.Infof("Starting RPC server on address: %s", lis.Addr().String())
	return s.server.Serve(lis)
}

// Stop stops the server.
func (s *Server) Stop() {
	s.server.Stop()
}

// GracefulStop stops the server gracefully.
func (s *Server) GracefulStop() {
	s.server.GracefulStop()
}

// StartInBackground starts serving in the background, returning an error if any issue is encountered
func (s *Server) StartInBackground() error {
	doneCh := make(chan error)
	go func() {
		err := s.Serve(func(started string) {
			log.Info("Started NBI on ", started)
			close(doneCh)
		})
		if err != nil {
			doneCh <- err
		}
	}()
	return <-doneCh
}
