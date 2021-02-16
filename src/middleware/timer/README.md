# `timer` - Execution Timer

The `timer` middleware measures the time taken to execute a pipe and prints it to the console for performance debugging purposes.

> The execution time of the pipe will be logged at `debug` level. You might need to use the `-v debug` argument to ensure the log entries are actually visible.

## Arguments

Enable time recording by setting the `record` argument to `true`.
```yaml
private:
    some-pipeline:
        timer:
            record: true
```
