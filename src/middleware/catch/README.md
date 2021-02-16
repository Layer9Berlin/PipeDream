# `catch` - Error Handling Middleware

## Arguments

### Error Handling Pipeline

The `catch` middleware handles stderr output by passing it as stdin to another pipe. The error handling pipe can be referenced either by a string (referencing a pipe defined elsewhere) or a map, providing additional arguments.

The handling pipe's stdout (if any) will be added to the calling pipe's stdout. Any stderr output from the error handling pipe will also be passed to the calling pipe, replacing the previous stderr output.

> Note that the error handling pipe will only be invoked if the calling pipe has non-trivial stderr output.

#### String Reference
```yaml
private:
    some-pipe:
        # use a string to refer to a pipe defined elsewhere
        catch: error-handling-pipe

    # this is the definition
    error-handling-pipe:
        # a map of arguments that define the handling pipe's behavior
        invocation: arguments
```

#### Map Reference
```yaml
private:
    some-pipe:
        catch:
            # pass a map to provide additional arguments to the error handling pipe
            # this can also be used to overwrite arguments in the definition
            # or to define the entire pipe inline
            error-handling-pipe:
                # these additional arguments will be merged into the definition's default arguments
                invocation: arguments

    # the definition is optional
    error-handling-pipe:
        # default arguments may be overwritten by invocation arguments
        default: arguments
```
