# Documentation

## Architecture

PipeDream consists of four main components:
- ### The parser
    The parser turns pipes from yaml format into Go structs that are then passed to the runner.
- The runner
    The runner decides which pipes to execute and passes them to the middleware stack.
- Middleware stack
    An array of callable middleware items. Each item, when called, may make changes to the invocation object and the result to be returned. Most middleware simply passes these changed objects to the next item in the stack, although it is also possible to trigger one or several runs of the complete middleware stacks instead. See the [middleware docs](../middleware) for more details.
- Built-in pipes
    For convenience, pipes encapsulating common shell commands are bundled into the installation. It is generally preferable to use these pipes instead of naked shell commands.

### Pipeline file format

Pipelines are defined in files with the `pipe` extension, containing yaml content.

It is good practice to start your yaml content with `---`, indicating the start of the document.

#### Version

Your pipeline file should contain a version indicator:

```yaml
Version: 0.0.1
```

This will help PipeDream interpret your pipe in the case of future syntax changes. If no version is specified, `0.0.1` is assumed.

#### Default settings

```yaml
Default:
    Command: some_pipe
    Dir: working_directory
```




```yaml
public:
    - some-pipeline:
        parameter1: value
        parameter2:
          key1: value1
          key2: value2
        pipe:
          - some-invocation:
             - 
```


## Extending PipeDream

### Defining Middleware


