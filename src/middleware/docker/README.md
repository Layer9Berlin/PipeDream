# `docker` - Docker Executor

## Arguments

The `docker` middleware enables the execution of pipedream commands within a Docker container, using `docker-compose exec`.

It takes a single string argument indicating the name of the Docker service in which the command should be run.

> Note that for convenience, the `docker` argument is inherited automatically. Child pipes without a `docker` argument will look within their direct ancestors and apply the most recent definition, if any.

# Inline Shell Command
```yaml
private:
    some-pipe:
        # the service in which any shell commands should be run
        docker: service-name
        # this will be executed as `docker-compose exec service-name "command"`
        shell:
            run: "command"
```

# Automatic Inheritance
```yaml
private:
    some-pipe:
        # define the service in which any shell commands should be run
        docker: service-name
        # invoke a child pipe
        pipe:
            child-pipe

    child-pipe:
        # this will be executed as `docker-compose exec service-name "command"`
        # the argument inheritance does not need to be specified explicitly
        shell:
            run: "command"
```
