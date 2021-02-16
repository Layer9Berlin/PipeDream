# `each` - Input Duplicator

## Arguments

The `each` middleware runs several pipes in parallel, copying the parent's input into each child and merging the children's stdout and stderr output respectively.

Although the child pipes are executed in parallel, their (stdout/stderr) output is merged into the parent sequentially, i.e. the complete output of the first child, followed by the complete output of the second child, etc.

Each child can be referenced either by a string (referring to a definition located elsewhere) or a map containing additional arguments (in which case the definition is optional). Any additional arguments provided in the invocation take precedence over the default arguments in the pipe's definition.

> Note that if several child pipes are running in parallel, the console output may only reflect the first one. This limitation will probably be removed in a future version of pipedream.

```yaml
private:
    some-pipe:
        each:
            # a child can be referenced by a string
            - child-pipe-1
            # or a map with additional arguments
            - child-pipe-2:
                invocation: arguments
            # in which case defining the pipe in another place is optional
            - child-pipe-3:
                invocation: arguments

    # define the behavior of the child pipes
    child-pipe-1:
        # default arguments may be overwritten by invocation arguments
        default: arguments

    child-pipe-2:
        # default arguments may be overwritten by invocation arguments
        default: arguments
```
