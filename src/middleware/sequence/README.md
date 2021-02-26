# `sequence` - Synchronous Executor

The `sequence` middleware executes pipes synchronously, starting each pipe only once the previous pipe has completed.

## Arguments

### Options

The `sequence` middleware takes an array argument, whose items are called *children*. The calling pipe is called the *parent*.

```yaml
private:
    some-pipe:
        sequence:
            - some-pipe::step-1
            - some-pipe::step-2:
                arg: value
            - some-pipe::step-3
```

will first execute `some-pipe::step-1`. After completion, `some-pipe::step-2` will be executed with arguments ```arg: value```. Finally, when `some-pipe::step-2` is complete, `some-pipe::step-3` is executed.
