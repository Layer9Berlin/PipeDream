# `catch` - Error Handler

## Arguments

```yaml
- some-pipeline:
    shell:
        run: "some-command-with-possible-stderr-output"
    catch:
      error-handler-pipeline:
        output:
            text: "don't worry about it"
```
