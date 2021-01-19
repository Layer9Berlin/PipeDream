# `timer` - Timer Middleware

## Arguments
Enable time recording by setting the `record` argument to `true`.
```yaml
- some_pipeline:
    timer:
      record: bool,
```
The execution time of the pipe will be logged at `debug` level. You might need to use the `-v debug` argument to ensure the log entries are actually visible.

### TODO:
- use dedicated logging channel to merge timing information into actual output