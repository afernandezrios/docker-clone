# Architecture & Technical Design

This document details the internal design of `ccrun`, the lifecycle of a container execution, its current scope, and known limitations not yet covered in the implementation.

## Architectural Lifecycle

`ccrun` uses a two-stage supervisor/init process pattern (similar to `runc` and other production runtimes) to avoid the pitfalls of mutating namespace states inside an already running multi-threaded Go runtime.

```
+-------------------------------------------------------------------------+
| HOST SUPERVISOR (Parent Process: ccrun <image> <cmd>)                   |
|                                                                         |
|  1. Pull Manifest & Layers (pkg/docker)                                 |
|     - Registry token auth negotiation                                   |
|     - Architecture matching (GOOS / GOARCH)                             |
|     - Download blobs to temporary rootfs directory                      |
|     - Extract tarballs and reconcile OCI layer whiteouts                |
|                                                                         |
|  2. Setup Control Groups (pkg/cgroup)                                   |
|     - Create cgroup node for CPU/memory resource boundaries             |
|                                                                         |
|  3. Spawn Child Process                                                 |
|     - Command: /proc/self/exe init --rootfs                             |
|     - SysProcAttr: CLONE_NEWUTS | CLONE_NEWPID | CLONE_NEWNS | ...      |
|     - UID / GID mappings configured (User Namespace root mapping)       |
|                                                                         |
|  4. Supervise Execution                                                 |
|     - Write child PID to cgroup tasks/cgroup.procs                      |
|     - Wait for child exit and forward exit code                         |
|     - Defer cleanup: delete cgroup path, wipe temp rootfs               |
+------------------------------------+------------------------------------+
                                     |
                        fork / exec with CLONE flags
                                     |
                                     v
+-------------------------------------------------------------------------+
| CONTAINER INIT (Child Process: ccrun init --rootfs <tempDir> <cmd>)     |
|                                                                         |
|  1. Load Image Config (pkg/config)                                      |
|     - Parse config.json for Env vars, WorkingDir, and entrypoint        |
|                                                                         |
|  2. Mount & Pivot (pkg/container, pkg/vfs)                              |
|     - Isolate mount propagation (make private/slave)                    |
|     - Mount pseudo filesystems (/proc, /sys, /dev, /dev/pts)            |
|     - pivot_root / chroot into new rootfs location                      |
|                                                                         |
|  3. Process Execution                                                   |
|     - Set working directory and inject environment variables            |
|     - Exec user command as PID 1                                        |
+-------------------------------------------------------------------------+
```

## Core Components

| Package / Module  | Responsibility                                                                                                                                                                                    |
| :---              | :---                                                                                                                                                                                              |
| **`cmd/run.go`**  | User-facing entrypoint. Coordinates image downloading, temporary filesystem allocation, cgroup setup, and spawns the isolated child process.                                                      |
| **`cmd/init.go`** | Hidden subcommand (`ccrun init`). Runs inside the isolated namespaces, applies image configs, mounts `/proc` and other virtual filesystems, and replaces itself with the requested user command.  |
| **`docker/`**     | Implements the OCI / Docker Registry HTTP API v2 client, token-based authentication challenge handling, blob retrieval, tarball extraction, and whiteout processing.                              |
| **`config/`**     | Parses OCI/Docker `config.json` specs to extract container environment variables, working directory, and command defaults.                                                                        |
| **`cgroup/`**     | Handles creation, PID attachment, and teardown of Linux control groups.                                                                                                                           |
| **`vfs/`**        | Manages mounting and unmounting of required virtual filesystems (`procfs`, `sysfs`, `devpts`) inside the container mount namespace.                                                               |

## Current Scope

- Pulling public images directly from standard Docker Hub registry endpoints.
- Parsing multi-architecture manifest lists and selecting matches for `linux/amd64` or `linux/arm64`.
- Unpacking standard `.tar.gz` layer archives into a flat root filesystem.
- Basic OCI whiteout resolution (`.wh.<file>` for file deletion, `.wh..wh..opq` for opaque directories).
- Process isolation via Linux UTS, PID, Mount, Cgroup, and User namespaces.
- Basic user mapping (mapping the current host UID/GID to container root `0:0`).
- Process lifecycle cleanup (ensuring child exit statuses bubble up and temporary directories/cgroups are unlinked).

## What Is Not Covered Yet (Future Roadmap)

While `ccrun` covers the core principles of running a container, several real-world runtime features are omitted for simplicity:

### 1. Copy-on-Write / OverlayFS Storage
- **Current state**: Layers are downloaded and uncompressed directly on top of each other into a flat directory on every container launch.
- **Not covered**: Using Linux `overlayfs` (lowerdir/upperdir/workdir) to share base image layers across containers and enable instant start times with minimal disk usage.

### 2. Local Layer Caching
- **Current state**: Images and layers are downloaded on every `run` call.
- **Not covered**: A local content-addressable storage cache (e.g. `/var/lib/ccrun/blobs/`) to prevent re-downloading existing layer digests.

### 3. Container Networking (CNI & Bridge Interfaces)
- **Current state**: Containers do not create a separate network namespace (`CLONE_NEWNET`) and share the host's network stack by default.
- **Not covered**: `veth` pair creation, host bridge configuration (`ccrun0`), IPAM (IP address allocation), and iptables NAT/port-forwarding rules.

### 4. Interactive TTY / PTY Allocation
- **Current state**: Process streams (`os.Stdin`, `os.Stdout`, `os.Stderr`) are forwarded directly.
- **Not covered**: Pseudoterminal allocation (`pty`/`openpty`), terminal raw mode handling, and window resize signal forwarding (`SIGWINCH`).

### 5. Private Registries & Authentication Credentials
- **Current state**: Anonymous token retrieval targeting Docker Hub.
- **Not covered**: Basic Auth for private registries, reading credentials from `~/.docker/config.json`, or third-party registries with non-standard auth challenge schemes.

### 6. Advanced Security Hardening
- **Current state**: Process runs with capabilities allowed by user namespace mapping.
- **Not covered**: Dropping Linux capabilities (e.g., `CAP_SYS_ADMIN`, `CAP_NET_ADMIN`), applying Seccomp syscall filter profiles, or configuring AppArmor / SELinux labels.