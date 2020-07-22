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

This creates a new `go.mod` file in the current directory that looks like so:
```
module example.com

go 1.14

require github.com/containers/libpod/v2 v2.0.2 // indirect
```

How do I use them

In this tutorial, you will learn through basic examples how to:

0. Background setup
1. Connect to the Podman system service
2. Pull an image
3. List images
4. Create and start a container from an image
5. Inspect the container
6. Stop the container
7. Complete Sample
8. Debugging tips

0. Background Setup

The recommended way to start podman system service in prod mode is via systemd
socket-activation:
```bash
$ systemctl --user start podman.socket
```

But for purposes of this demo, we will start the service using the podman
command itself. Open a terminal window and start the Podman system service:
```bash
$ podman system service -t0
```

Open another terminal window and check if the Podman socket exists:
```bash
$ ls -al /run/user/1000/podman
podman.sock
```

1. Create a connection to the system service
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


2. Pull an image

Next, we will pull an image using the images.Pull() binding.
This binding takes three arguments:
    - The context variable created earlier
    - The image name
    - Options for image pull

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

Run it:

```bash
$ go run main.go
Welcome to Go bindings tutorial
Pulling image...
$
```

The system service side should echo messages like so:
```bash
Trying to pull registry.fedoraproject.org/fedora:latest...
Getting image source signatures
Copying blob dd9f43919ba0 done
Copying config 00ff39a8bf done
Writing manifest to image destination
Storing signatures
```


3. List images


4. Create and Start Container from Image
To create the container spec, we use specgen.NewSpecGenerator() followed by
calling containers.CreateWithSpec() to actually create a new container.
specgen.NewSpecGenerator takes 2 arguments:
    - name of the image
    - whether it's a rootfs

containers.CreateWithSpec takes 2 arguments
    - the context created earlier
    - the spec created by NewSpecGenerator

Next, the container is actually started using the containers.Start() binding.
containers.Start() takes three args:
    - the context
    - the name or ID of the container created
    - an optional parameter for detach keys

After the container is started, it's a good idea to ensure the container is in
a running state before you proceed with further operations. The
containers.Wait() takes care of that.
containers.Wait() takes three args:
    - the context
    - the name or ID of the container created
    - container state (running/paused/stopped)
```golang
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

        running := define.ContainerStateRunning
        _, err = containers.Wait(conn, r.ID, &running)
        if err != nil {
                fmt.Println(err)
                return
        }
```

Run it:
```bash
$ go run main.go
Welcome to Go bindings tutorial
Pulling image...
Starting Fedora container...
$
```

Check if the container is running:
```bash
$ podman ps
CONTAINER ID  IMAGE                                     COMMAND    CREATED                 STATUS                     PORTS   NAMES
665831d31e90  registry.fedoraproject.org/fedora:latest  /bin/bash  Less than a second ago  Up Less than a second ago          dazzling_mclean
```


5. Inspect Container
Containers can be inspected using the containers.Inspect() binding.
containers.Inspect() takes 3 args:
    - context
    - image name or ID
    - optional boolean to check for container size

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

Run it:
```bash
$ go run main.go
Welcome to Go bindings tutorial
Pulling image...
Starting Fedora container...
Container uses image registry.fedoraproject.org/fedora:latest
Container running status is running
$
```

6. Stop Container
A container can be stopped by the containers.Stop() binding.
containers.Stop() takes 3 args:
    - context
    - image name or ID
    - optional timeout
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

Run it:
```bash
$ go run main.go
Welcome to Go bindings tutorial
Pulling image...
Starting Fedora container...
Container uses image registry.fedoraproject.org/fedora:latest
Container running status is running
Stopping the container...
Container running status is now exited
$
```


7. Complete Sample:
The sample can be cloned from https://github.com/lsm5/bindings-sample . This
repo includes the go module information required to build the code.

You can also find it below for reference.

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


        running := define.ContainerStateRunning
        _, err = containers.Wait(conn, r.ID, &running)
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

8. Debugging tips
To debug in a dev setup, you can start the Podman system service in debug mode
like so:
```bash
$ podman --log-level=debug system service -t0
```

This will echo additional messages, a snippet of which can be seen below:
```
INFO[0000] podman filtering at log level debug          
DEBU[0000] Called service.PersistentPreRunE(podman --log-level=debug system service -t0) 
DEBU[0000] Ignoring libpod.conf EventsLogger setting "/home/lsm5/.config/containers/containers.conf". Use "journald" if you want to change this setting and remove libpod.conf files. 
DEBU[0000] Reading configuration file "/usr/share/containers/containers.conf" 
DEBU[0000] Merged system config "/usr/share/containers/containers.conf": &{{[] [] containers-default-0.14.4 [] private enabled [CAP_AUDIT_WRITE CAP_CHOWN CAP_DAC_OVERRIDE CAP_FOWNER CAP_FSETID CAP_KILL CAP_MKNOD CAP_NET_BIND_SERVICE CAP_NET_RAW CAP_SETFCAP CAP_SETGID CAP_SETPCAP CAP_SETUID CAP_SYS_CHROOT] [] []  [] [] [] true [PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin] false false false  private k8s-file -1 slirp4netns false 2048 private /usr/share/containers/seccomp.json 65536k private host 65536} {true systemd [PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin] [/usr/libexec/podman/conmon /usr/local/libexec/podman/conmon /usr/local/lib/podman/conmon /usr/bin/conmon /usr/sbin/conmon /usr/local/bin/conmon /usr/local/sbin/conmon /run/current-system/sw/bin/conmon] ctrl-p,ctrl-q true /run/user/1000/libpod/tmp/events/events.log file [/usr/share/containers/oci/hooks.d] docker:// /pause k8s.gcr.io/pause:3.2 /usr/libexec/podman/catatonit shm   false 2048 /usr/bin/crun map[crun:[/usr/bin/crun /usr/sbin/crun /usr/local/bin/crun /usr/local/sbin/crun /sbin/crun /bin/crun /run/current-system/sw/bin/crun] kata:[/usr/bin/kata-runtime /usr/sbin/kata-runtime /usr/local/bin/kata-runtime /usr/local/sbin/kata-runtime /sbin/kata-runtime /bin/kata-runtime /usr/bin/kata-qemu /usr/bin/kata-fc] runc:[/usr/bin/runc /usr/sbin/runc /usr/local/bin/runc /usr/local/sbin/runc /sbin/runc /bin/runc /usr/lib/cri-o-runc/sbin/runc /run/current-system/sw/bin/runc]] missing false   [] [crun runc] [crun] [kata kata-runtime kata-qemu kata-fc] {false false false false false false} /etc/containers/policy.json false 3 /home/lsm5/.local/share/containers/storage/libpod 10 /run/user/1000/libpod/tmp /home/lsm5/.local/share/containers/storage/volumes} {[/usr/libexec/cni /usr/lib/cni /usr/local/lib/cni /opt/cni/bin] podman /etc/cni/net.d/}} 
DEBU[0000] Using conmon: "/usr/bin/conmon"              
DEBU[0000] Initializing boltdb state at /home/lsm5/.local/share/containers/storage/libpod/bolt_state.db 
DEBU[0000] Overriding run root "/run/user/1000/containers" with "/run/user/1000" from database 
DEBU[0000] Using graph driver overlay                   
DEBU[0000] Using graph root /home/lsm5/.local/share/containers/storage 
DEBU[0000] Using run root /run/user/1000                
DEBU[0000] Using static dir /home/lsm5/.local/share/containers/storage/libpod 
DEBU[0000] Using tmp dir /run/user/1000/libpod/tmp      
DEBU[0000] Using volume path /home/lsm5/.local/share/containers/storage/volumes 
DEBU[0000] Set libpod namespace to ""                   
DEBU[0000] Not configuring container store              
DEBU[0000] Initializing event backend file              
DEBU[0000] using runtime "/usr/bin/runc"                
DEBU[0000] using runtime "/usr/bin/crun"                
WARN[0000] Error initializing configured OCI runtime kata: no valid executable found for OCI runtime kata: invalid argument 
DEBU[0000] using runtime "/usr/bin/crun"                
INFO[0000] Setting parallel job count to 25             
INFO[0000] podman filtering at log level debug          
DEBU[0000] Called service.PersistentPreRunE(podman --log-level=debug system service -t0) 
DEBU[0000] Ignoring libpod.conf EventsLogger setting "/home/lsm5/.config/containers/containers.conf". Use "journald" if you want to change this setting and remove libpod.conf files. 
DEBU[0000] Reading configuration file "/usr/share/containers/containers.conf" 
```

If the Podman system service has been started via systemd socket activation,
you can view the logs using journalctl. The logs after a sample run look like so:

```bash
$ journalctl --user --no-pager -u podman.socket
-- Reboot --
Jul 22 13:50:40 nagato.nanadai.me systemd[1048]: Listening on Podman API Socket.
$
```

```bash
$ journalctl --user --no-pager -u podman.service
Jul 22 13:50:53 nagato.nanadai.me systemd[1048]: Starting Podman API Service...
Jul 22 13:50:54 nagato.nanadai.me podman[1527]: time="2020-07-22T13:50:54-04:00" level=error msg="Error refreshing volume 38480630a8bdaa3e1a0ebd34c94038591b0d7ad994b37be5b4f2072bb6ef0879: error acquiring lock 0 for volume 38480630a8bdaa3e1a0ebd34c94038591b0d7ad994b37be5b4f2072bb6ef0879: file exists"
Jul 22 13:50:54 nagato.nanadai.me podman[1527]: time="2020-07-22T13:50:54-04:00" level=error msg="Error refreshing volume 47d410af4d762a0cc456a89e58f759937146fa3be32b5e95a698a1d4069f4024: error acquiring lock 0 for volume 47d410af4d762a0cc456a89e58f759937146fa3be32b5e95a698a1d4069f4024: file exists"
Jul 22 13:50:54 nagato.nanadai.me podman[1527]: time="2020-07-22T13:50:54-04:00" level=error msg="Error refreshing volume 86e73f082e344dad38c8792fb86b2017c4f133f2a8db87f239d1d28a78cf0868: error acquiring lock 0 for volume 86e73f082e344dad38c8792fb86b2017c4f133f2a8db87f239d1d28a78cf0868: file exists"
Jul 22 13:50:54 nagato.nanadai.me podman[1527]: time="2020-07-22T13:50:54-04:00" level=error msg="Error refreshing volume 9a16ea764be490a5563e384d9074ab0495e4d9119be380c664037d6cf1215631: error acquiring lock 0 for volume 9a16ea764be490a5563e384d9074ab0495e4d9119be380c664037d6cf1215631: file exists"
Jul 22 13:50:54 nagato.nanadai.me podman[1527]: time="2020-07-22T13:50:54-04:00" level=error msg="Error refreshing volume bfd6b2a97217f8655add13e0ad3f6b8e1c79bc1519b7a1e15361a107ccf57fc0: error acquiring lock 0 for volume bfd6b2a97217f8655add13e0ad3f6b8e1c79bc1519b7a1e15361a107ccf57fc0: file exists"
Jul 22 13:50:54 nagato.nanadai.me podman[1527]: time="2020-07-22T13:50:54-04:00" level=error msg="Error refreshing volume f9b9f630982452ebcbed24bd229b142fbeecd5d4c85791fca440b21d56fef563: error acquiring lock 0 for volume f9b9f630982452ebcbed24bd229b142fbeecd5d4c85791fca440b21d56fef563: file exists"
Jul 22 13:50:54 nagato.nanadai.me podman[1527]: Trying to pull registry.fedoraproject.org/fedora:latest...
Jul 22 13:50:55 nagato.nanadai.me podman[1527]: Getting image source signatures
Jul 22 13:50:55 nagato.nanadai.me podman[1527]: Copying blob sha256:dd9f43919ba05f05d4f783c31e83e5e776c4f5d29dd72b9ec5056b9576c10053
Jul 22 13:50:55 nagato.nanadai.me podman[1527]: Copying config sha256:00ff39a8bf19f810a7e641f7eb3ddc47635913a19c4996debd91fafb6b379069
Jul 22 13:50:55 nagato.nanadai.me podman[1527]: Writing manifest to image destination
Jul 22 13:50:55 nagato.nanadai.me podman[1527]: Storing signatures
Jul 22 13:50:55 nagato.nanadai.me systemd[1048]: podman.service: unit configures an IP firewall, but not running as root.
Jul 22 13:50:55 nagato.nanadai.me systemd[1048]: (This warning is only shown for the first unit using IP firewalling.)
Jul 22 13:51:15 nagato.nanadai.me systemd[1048]: podman.service: Succeeded.
Jul 22 13:51:15 nagato.nanadai.me systemd[1048]: Finished Podman API Service.
Jul 22 13:51:15 nagato.nanadai.me systemd[1048]: podman.service: Consumed 1.339s CPU time.
$
```


Any issues with the bindings can be reported [upstream](https://github.com/containers/podman/issues/new/choose)


Wrap Up
    - Podman v2 provides a set of Go bindings to allow developers to integrate Podman
    functionality conveniently in their Go application.

    - These Go bindings need Podman system service to be running in the
    background. This can be achieved using systemd socket activation. 

    - **ANYTHING ELSE??**
