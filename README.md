# docker-clone
https://codingchallenges.fyi/challenges/challenge-docker/

## How to run

`ccrun` is a cli tool to run commands. For example, to run "ls ." command:

```bash
cd src
sudo go run main.go ccrun /bin/busybox sh
```

### User namespace

To run the container rootless we need to create a new User namespace. This way we avoid that the root user within the container has root privileges on the host.

To create this new namespace we have to include the Linux system call flag `CLONE_NEWUSER`.

Besides, we need to map the User ID (UID) and Group ID (GID) of the host machine to the UID and GID of the root user within the container. More info: [man User namespace](https://man7.org/linux/man-pages/man7/user_namespaces.7.html)