# `ssh` - SSH Executor

## Arguments

The `ssh` middleware enables the execution of pipedream commands on a remote server via SSH.

It takes a single string argument indicating the name of the SSH host on which the command should be run.

> Note that for convenience, the `ssh` argument is inherited automatically. Child pipes without an `ssh` argument will look within their direct ancestors and apply the most recent definition, if any.

## Inline Shell Command
```yaml
private:
    some-pipe:
        # the name of the remote host on which any shell commands should be run
        # it will be resolved using your SSH configuration
        ssh: hostname
        # this will be executed as `ssh hostname "command"`
        shell:
            run: "command"
```

## Automatic Inheritance
```yaml
private:
    some-pipe:
        # the name of the remote host on which any shell commands should be run
        # it will be resolved using your SSH configuration
        ssh: hostname
        # invoke a child pipe
        pipe:
            child-pipe

    child-pipe:
        # this will be executed as `ssh hostname "command"`
        # the argument inheritance does not need to be specified explicitly
        shell:
            run: "command"
```
