Purpose and audience statement:  The purpose of this document is to teach
through examples on how to use Podman golang bindings.
The examples should be heavily commented and do simple things without
being too confusing.
The audience is developers who may or may not be familiar with golang.

Podman Go bindings

Introduction

In the recent release of Podman 2.0, we removed the experimental tag from
its recently introduced RESTful service.  While it might be interesting
to interact with a RESTFul server using curl, using a set of Go based
bindings is probably a more direct route to some production ready application.

The podman Go bindings are a set of functions to allow developers to execute
Podman operations from within their Go based application. The Go bindings
connect to a Podman service which can run locally or on a remote machine.
You can perform many operations including pulling and listing images,
starting, stopping and inspecting containers. Currently, the Podman repository
has bindings available for operations on images, containers, pods, networks
and manifests among others. The bindings are available on the upstream
Podman repository in the v2 branch. You can fetch the bindings for your
application using Go modules:

```bash
$ cd $HOME
$ mkdir example && cd example
$ go mod init example.com
$ go get github.com/containers/libpod/v2@v2.0.2
```

How do I use them

In this tutorial, you will learn through a basic example how to:
Create connection to the service
1. Pull an image
2. List images
3. Create a container
4. Start the container
5. Inspect the container
6. Stop the container

Background Setup
Open a Terminal window and start the Podman system service:
```bash
$ podman system service -t0
```

Open another Terminal window and check if the Podman socket exists:
```bash
$ ls -al /run/user/1000/podman
podman.sock
```

Example Zero -- Create a connection to the system service
After you set up your basic main method, you need to create a connection
that connects to the system service.  The critical piece of information
for setting up a new connection is the endpoint. The endpoint comes in
the form of an URI (method:/path/to/socket). For example, to connect to
the local rootful socket the URI would be
`unix:///run/podman/podman.sock` and for a rootless user it would be
`unix://$(XDG_RUNTIME_DIR)/podman/podman.sock`,
typically: `unix:///run/user/1000/podman/podman.sock`

The following example snippet shows how to set up a connection for a rootless user.
```golang
package main

import (
        "context"
        "fmt"
        "os"

        "github.com/containers/libpod/v2/libpod/define"
        "github.com/containers/libpod/v2/pkg/bindings/containers"
        "github.com/containers/libpod/v2/pkg/bindings/images"
        "github.com/containers/libpod/v2/pkg/domain/entities"
        "github.com/containers/libpod/v2/pkg/specgen"
        "github.com/containers/libpod/v2/pkg/bindings"
)

func main() {
        fmt.Println("Welcome to Go bindings tutorial")

        // Get Podman socket location
        sock_dir := os.Getenv("XDG_RUNTIME_DIR")
        socket := "unix://" + sock_dir + "/podman/podman.sock"

        // Connect to Podman socket
        conn, err := bindings.NewConnection(context.Background(), socket)
        if err != nil {
                fmt.Println(err)
                return
        }
}
```

The `conn` variable received from the NewConnection function is of type
context.Context().  In subsequent uses of the bindings, you will use
this context to direct the bindings to your connection. This can be
seen in the examples below.


Example One -- Pull an image Listing images
Next, we will pull an image using the images.Pull() binding.
This binding takes three arguments:
1. The context variable created earlier: conn
2. The image name: rawImage
3. Options for image pull: entities.ImagePullOptions{}

Append the following lines to your main() function.
```golang
        // Pull image
        rawImage := "registry.fedoraproject.org/fedora:latest"
        fmt.Println("Pulling image...")
        _, err = images.Pull(conn, rawImage, entities.ImagePullOptions{})
        if err != nil {
                fmt.Println(err)
                return
        }
```

Next, we will run our code 
Example Two -- List imagesPull an image and run the container
connection to the socket and pull an image using the images.Pull() binding.
To create the container spec, we use specgen.NewSpecGenerator() followed by
calling containers.CreateWithSpec() to actually create a new container.


Example Two -- List Images


Example Three -- Create and Start Container from Image
```golang
        // Container create
        s := specgen.NewSpecGenerator(rawImage, false)
        s.Terminal = true
        r, err := containers.CreateWithSpec(conn, s)
        if err != nil {
                fmt.Println(err)
                return
        }
```


Example Four -- Start Container
```golang
        // Container start
        fmt.Println("Starting Fedora container...")
        err = containers.Start(conn, r.ID, nil)
        if err != nil {
                fmt.Println(err)
                return
        }
```


Example Five -- Inspect Container
```golang
        // Container inspect
        ctrData, err := containers.Inspect(conn, r.ID, nil)
        if err != nil {
                fmt.Println(err)
                return
        }
        fmt.Printf("Container uses image %s\n", ctrData.ImageName)
        fmt.Printf("Container running status is %s\n", ctrData.State.Status)
```


Example Six -- Wait for Container to Run
```golang
        // Wait for container to run
        running := define.ContainerStateRunning
        _, err = containers.Wait(conn, r.ID, &running)
        if err != nil {
                fmt.Println(err)
                return
        }
```


Example Seven -- Stop Container
```golang
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
```


Complete Sample:
```golang
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
        fmt.Println("Welcome to Go bindings tutorial")

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

        // TODO: Insert Image List code here

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

        // Container inspect
        ctrData, err := containers.Inspect(conn, r.ID, nil)
        if err != nil {
                fmt.Println(err)
                return
        }
        fmt.Printf("Container uses image %s\n", ctrData.ImageName)
        fmt.Printf("Container running status is %s\n", ctrData.State.Status)

        running := define.ContainerStateRunning
        _, err = containers.Wait(conn, r.ID, &running)
        if err != nil {
                fmt.Println(err)
                return
        }

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
```

