# `dir` - Directory Navigator

## Arguments

### Dir Argument

The `dir` middleware takes a single string argument indicating the directory in which any commands should be executed on the local machine.

You can specify either a relative or an absolute. Relative paths are evaluated with respect to the directory in which the pipedream command is executed.

> Note that other middleware (like `ssh` and `docker`) might cause commands to take effect on remote machines. The `dir` middleware only controls the working directory for the local command. Use the `dir` argument of the `shell` middleware to control the working directory on the remote machine.

#### Relative Path
```yaml
private:
    some-pipe:
        # any shell commands will be executed in the specified subdirectory of the directory in which pipedream is executed
        dir: some/relative/path

    some-other-pipe:
        # this has the same effect
        dir: ./some/relative/path
```

#### Absolute Path
```yaml
private:
    some-pipe:
        # any shell commands will be executed in the specified directory
        dir: /some/absolute/path
```
