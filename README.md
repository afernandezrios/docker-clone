# ccrun - A Minimal Container Runtime in Go
https://codingchallenges.fyi/challenges/challenge-docker/

`ccrun` is an educational container runtime written in Go from scratch. The main goal of this project is to demystify container internals by building the core components by hand—without relying on heavy daemons or container management toolchains like Docker, `containerd`, or `runc`.

By writing each part step-by-step, this project demonstrates how containers are not magical virtual machines, but rather ordinary Linux processes constrained by Linux namespaces, cgroups, and an isolated root filesystem constructed from OCI image layers.

## What It Does

- **Pulls OCI / Docker Hub Images**: Communicates directly with registry endpoints, handles token authentication challenges, resolves multi-architecture manifest lists, and downloads layer blobs.
- **Builds the Root Filesystem (`rootfs`)**: Unpacks sequential gzip tarballs and processes OCI whiteout markers (`.wh.<file>` and `.wh..wh..opq`) to handle file deletions and layer diffs correctly.
- **Isolates Processes with Namespaces**: Spawns isolated environments using UTS, PID, Mount, Cgroup, and User namespaces (`CLONE_NEW*`).
- **Enforces Resource Limits**: Configures Linux cgroups (CPU, memory) and attaches the target container process.
- **Initializes the Environment**: Re-executes itself via `/proc/self/exe init` to mount virtual filesystems (`/proc`, `/sys`, `/dev`), apply image environment variables and working directories, and run the user process as PID 1.

## Getting Started

### Prerequisites

- **Linux Operating System**: Containers are a native Linux kernel feature. A modern kernel (5.x or newer) is recommended.
- **Go**: Version 1.22 or higher.
- **Root Privileges**: Setting up cgroups, mounting virtual filesystems, and configuring certain namespaces requires `sudo` / root capabilities.

## How to run

`ccrun` is a cli tool to run commands inside a container. The expected input for the cli tool is:

`sudo go run main.go <imageName> <command>`

For example, to run a shell inside the container:

```bash
cd src
sudo go run main.go ccrun library/alpine sh
```

**Not all the images are available at this moment. There are a lot of things not implemented in this POC. Some tested
images: library/alpine, library/ubuntu, alpine/git**

### Build and run program

```sh
cd src
go build
sudo ./docker-clone ccrun library/alpine sh
```

### Run test

```sh
cd src
sudo -E go test -v -run TestIntegration_RunContainer .
```
