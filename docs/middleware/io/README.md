# `io` - Input/Output Middleware

## Arguments

```yaml
- some-pipeline:
    io: merge
    pipe:
      - some-invocation:
          catch:
            - pattern: "test"
              handler: error-handler
      - some-other-invocation:
          out:
            merge:
              - in
              - out
              - parent
```
