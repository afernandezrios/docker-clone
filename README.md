# docker-clone
https://codingchallenges.fyi/challenges/challenge-docker/

## How to run

`ccrun` is a cli tool to run commands inside a container. For example, to run a shell inside the container:

```bash
cd src
sudo go run main.go ccrun sh
```

### User namespace

To run the container rootless we need to create a new User namespace. This way we avoid that the root user within the container has root privileges on the host.

To create this new namespace we have to include the Linux system call flag `CLONE_NEWUSER`.

Besides, we need to map the User ID (UID) and Group ID (GID) of the host machine to the UID and GID of the root user within the container. More info: [man User namespace](https://man7.org/linux/man-pages/man7/user_namespaces.7.html)


### Control groups

Control groups are used to limit the resources that processes can use. In this case, to limit how much memory and cpu the conteiner is able to use, a new `cgroup` is created (`/sys/fs/cgroup/containercg`).

Once it is created, container resources can be limited, for example, file `/sys/fs/cgroup/containercg/memory.max` limits the max memory usage.

**IMPORTANT**: It is necessary to include the pid of the container to the new cgroup. To do that, the pid must be added to the `/sys/fs/cgroup/containercg/cgroup.procs` file.

See [this tutorial](https://labs.iximiuz.com/tutorials/controlling-process-resources-with-cgroups) to see how a Control Group is created manually.
