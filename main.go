package main

import (
	"context"
	"fmt"

	"github.com/containers/libpod/v2/pkg/bindings"
	"github.com/containers/libpod/v2/pkg/bindings/containers"
	"github.com/containers/libpod/v2/pkg/specgen"
)

type binding struct {
	sock string
	conn context.Context
}

func (b *binding) NewConnection() error {
	connText, err := bindings.NewConnection(context.Background(), b.sock)
	if err != nil {
		return err
	}
	b.conn = connText
	return nil
}

func newBinding() *binding {
	b := binding{
		sock: "unix:///run/user/1000/podman/podman.sock",
	}
	return &b
}

func main() {
	rawImage := "quay.io/libpod/alpine_nginx:latest"
	fmt.Println("Welcome to Go bindings tutorial")
	b := newBinding()
	err := b.NewConnection()
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}

	// Container create
	s := specgen.NewSpecGenerator(rawImage, false)
	s.Terminal = true
	r, err := containers.CreateWithSpec(b.conn, s)
	// Container start
	err = containers.Start(b.conn, r.ID, nil)

	// Container inspect
	ctrData, err := containers.Inspect(b.conn, r.ID, nil)
	fmt.Printf("Container source image is %s\n", ctrData.ImageName)
	fmt.Printf("Container running status is %s\n", ctrData.State.Status)

	// Container pause
	//crun failure
	//err = containers.Pause(b.conn, r.ID)

	// Container stop
	err = containers.Stop(b.conn, r.ID, nil)
	ctrData, err = containers.Inspect(b.conn, r.ID, nil)
	fmt.Printf("Container running status is now %s\n", ctrData.State.Status)
}
