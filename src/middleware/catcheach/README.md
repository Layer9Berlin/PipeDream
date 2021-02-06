# `catch` - Error Handling Middleware

## Arguments

### Matching rules
Define error matching rules using the `pattern` argument. It expects a regex that matches each handled line in the StdErr output.

Any errors that occur will be matched against your rules, and the first match will be applied. Subsequent pipes will not see these errors.

```yaml
- some_pipeline:
    catch:
      pattern: "^Error: ",
```

### Handling rules
The `ignore` argument will cause all matched errors to be swallowed instead of being passed to subsequent pipelines. Use with caution.
```yaml
- some_pipeline:
    catch:
      pattern: "^Error: ",
      ignore: true
```

Some errors are best treated as output. The `divert-to-output` argument prints the error message in StdOut and does not pass it on.
```yaml
- some_pipeline:
    catch:
      pattern: "^Error: ",
      divert-to-output: true
```

Use the `explain-error` argument to wrap the error using a GoLang format string. The placeholder `%w` can be used to refer to the underlying error. Note that a handler will not swallow the error, unless combined with another argument, such as `ignore`.
```yaml
- some_pipeline:
    catch:
      pattern: "^Error: ",
      explain-error: "failed to parse configuration file due to syntax error: %w"
```
