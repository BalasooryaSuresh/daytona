// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package headscale

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"path/filepath"

	"tailscale.com/tsnet"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	log "github.com/sirupsen/logrus"
)

var tsNetServer = &tsnet.Server{
	Hostname: "server",
}

func (s *HeadscaleServer) Connect(username string) error {
	err := s.CreateUser(username)
	if err != nil {
		log.Fatal(err)
	}

	authKey, err := s.CreateAuthKey(username)
	if err != nil {
		log.Fatal(err)
	}

	tsNetServer.ControlURL = fmt.Sprintf("http://localhost:%d", s.headscalePort)
	tsNetServer.AuthKey = authKey

	configDir, err := config.GetConfigDir()
	if err != nil {
		return err
	}
	tsNetServer.Dir = filepath.Join(configDir, "tsnet")

	defer tsNetServer.Close()
	ln, err := tsNetServer.Listen("tcp", ":80")
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	go func() {
		<-s.disconnectChan
		tsNetServer.Close()
		ln.Close()
	}()

	return http.Serve(ln, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Ok\n")
	}))
}

func (s *HeadscaleServer) HTTPClient() *http.Client {
	return tsNetServer.HTTPClient()
}

func (s *HeadscaleServer) Dial(ctx context.Context, network, addr string) (net.Conn, error) {
	return tsNetServer.Dial(ctx, network, addr)
}
