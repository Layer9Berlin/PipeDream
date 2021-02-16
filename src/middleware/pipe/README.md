# `pipe` - Invocation Chainer

The `pipe` middleware chains pipes together, so that the output of one pipe becomes the input of another. It is asynchronous in the sense that all pipes will be invoked in sequence, but the execution happens in parallel. In other words, pipes can process their input as it becomes available and do not need to wait for previous pipes to finish before they can start executing. If you are familiar with linux pipes, this is probably what you would expect.

## Arguments

The `pipe` middleware takes an array argument, whose items are called *children*. The calling pipe is called the *parent*.

Each child can be referenced either by a string (referring to a definition located elsewhere) or a map containing additional arguments (in which case the definition is optional). Any additional arguments provided in the invocation take precedence over the default arguments in the pipe's definition.

```yaml
private:
    parent-pipe:
        pipe:
            # a child can be referenced by a string
            # the input of `child-pipe-1` will be the output of `parent-pipe`
            - child-pipe-1
            # or a map with additional arguments
            # the input of `child-pipe-2` will be the output of `child-pipe-1`
            - child-pipe-2:
                invocation: arguments
            # in which case defining the pipe in another place is optional
            # the input of `child-pipe-3` will be the output of `child-pipe-2`
            - child-pipe-3:
                invocation: arguments
    # the (stdout) output of `parent-pipe` will be the output of `child-pipe-3`
    # the stderr output of `parent-pipe` will also be the stderr output of `child-pipe-3`

    # define the behavior of the child pipes
    child-pipe-1:
        # default arguments may be overwritten by invocation arguments
        default: arguments

    child-pipe-2:
        # default arguments may be overwritten by invocation arguments
        default: arguments
```
