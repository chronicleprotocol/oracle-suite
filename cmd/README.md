# Command line tools available for building

## Docker Images

All docker images have reasonable defaults and are set up to run the main application for each service.

Check `docker-compose.yml` in each app for more information.

```shell
# Run:
docker-compose -f ./cmd/<app>/docker-compose.yml run --rm <app> --help
# to see available options and additional commands.
```
