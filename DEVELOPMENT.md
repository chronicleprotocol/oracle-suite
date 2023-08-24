# Guidelines and instructions for developing the project

## Local development

Simplest way to run the project locally is to use Docker Compose.
> The examples use [Compose V2](https://www.docker.com/blog/announcing-compose-v2-general-availability/)

1. Build local image for Adria/MuSig

```shell
docker compose \
  -f ./cmd/spire/docker-compose.yaml \
  -f ./cmd/ghost/docker-compose.yaml \
  -f ./cmd/spectre/docker-compose.yaml \
  -f ./cmd/gofer/docker-compose.yaml \
  build
```

2. Run a local set of containers

> Note, the `network_mode:` is set to `host` for the provided `docker-compose.yaml`.

```shell
docker compose up
# or
docker compose up -d \
&& docker compose logs -f # to keep the containers running after exiting the logs
```

3. Debugging with breakpoints

Stop one of the containers

```shell
docker command stop <service>
```

and run the corresponding app in a debugger using
the same environment variables as in the `docker-compose.yaml` file.
