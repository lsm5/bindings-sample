package main

import (
	"context"
	"fmt"

	"github.com/containers/libpod/v2/libpod/define"
	"github.com/containers/libpod/v2/pkg/bindings"
	"github.com/containers/libpod/v2/pkg/bindings/containers"
	"github.com/containers/libpod/v2/pkg/bindings/images"
	"github.com/containers/libpod/v2/pkg/domain/entities"
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
	rawImage := "registry.fedoraproject.org/fedora:latest"
	running := define.ContainerStateRunning
	fmt.Println("Welcome to Go bindings tutorial")
	b := newBinding()
	err := b.NewConnection()
	if err != nil {
		fmt.Println(err)
		return
	}

	// Pull image
	fmt.Println("Pulling image...")
	_, err = images.Pull(b.conn, rawImage, entities.ImagePullOptions{})
	if err != nil {
		fmt.Println(err)
	}

	// Container create
	s := specgen.NewSpecGenerator(rawImage, false)
	s.Terminal = true
	r, err := containers.CreateWithSpec(b.conn, s)
	if err != nil {
		fmt.Println(err)
	}
	// Container start
	fmt.Println("Starting Fedora container...")
	err = containers.Start(b.conn, r.ID, nil)
	if err != nil {
		fmt.Println(err)
	}

	// Container inspect
	ctrData, err := containers.Inspect(b.conn, r.ID, nil)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("Container uses image %s\n", ctrData.ImageName)
	fmt.Printf("Container running status is %s\n", ctrData.State.Status)

	_, err = containers.Wait(b.conn, r.ID, &running)
	if err != nil {
		fmt.Println(err)
	}

	// Container stop
	fmt.Println("Stopping the container...")
	err = containers.Stop(b.conn, r.ID, nil)
	if err != nil {
		fmt.Println(err)
	}
	ctrData, err = containers.Inspect(b.conn, r.ID, nil)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("Container running status is now %s\n", ctrData.State.Status)

}
