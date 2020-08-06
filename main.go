package main

import (
	"context"
	"fmt"
	"os"

	"github.com/containers/libpod/v2/libpod/define"
	"github.com/containers/libpod/v2/pkg/bindings"
	"github.com/containers/libpod/v2/pkg/bindings/containers"
	"github.com/containers/libpod/v2/pkg/bindings/images"
	"github.com/containers/libpod/v2/pkg/domain/entities"
	"github.com/containers/libpod/v2/pkg/specgen"
)

func main() {
	fmt.Println("Welcome to Podman Go bindings tutorial")

	// Get Podman socket location
	sock_dir := os.Getenv("XDG_RUNTIME_DIR")
	socket := "unix://" + sock_dir + "/podman/podman.sock"

	// Connect to Podman socket
	conn, err := bindings.NewConnection(context.Background(), socket)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Pull image
	rawImage := "registry.fedoraproject.org/fedora:latest"
	fmt.Println("Pulling image...")
	_, err = images.Pull(conn, rawImage, entities.ImagePullOptions{})
	if err != nil {
		fmt.Println(err)
		return
	}

	// List images (WIP)
	imageSummary, err := images.List(conn, nil, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	var names []string
	for _, i := range imageSummary {
		names = append(names, i.RepoTags...)
	}
	fmt.Println(names)

	// Container create
	s := specgen.NewSpecGenerator(rawImage, false)
	s.Terminal = true
	r, err := containers.CreateWithSpec(conn, s)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Container start
	fmt.Println("Starting Fedora container...")
	err = containers.Start(conn, r.ID, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Wait for container to run
	running := define.ContainerStateRunning
	_, err = containers.Wait(conn, r.ID, &running)
	if err != nil {
		fmt.Println(err)
		return
	}

	// List containers
	var latestContainers = 1
	containerLatestList, err := containers.List(conn, nil, nil, &latestContainers, nil, nil, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Latest container is %s\n", containerLatestList[0].Names[0])

	// Container inspect
	ctrData, err := containers.Inspect(conn, r.ID, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Container uses image %s\n", ctrData.ImageName)
	fmt.Printf("Container running status is %s\n", ctrData.State.Status)

	// Container stop
	fmt.Println("Stopping the container...")
	err = containers.Stop(conn, r.ID, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	ctrData, err = containers.Inspect(conn, r.ID, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Container running status is now %s\n", ctrData.State.Status)
	return

}
