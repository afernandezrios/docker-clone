# docker-clone

https://codingchallenges.fyi/challenges/challenge-docker/

## How to run

`ccrun` is a cli tool to run commands inside a container. The expected input for the cli tool is:
`sudo go run main.go imageName command`

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
