# `input` - Input Override


The `input` middleware overrides the pipe's input, replacing it with a fixed value.

## Arguments

It takes a single string argument:

### Text

The `text` argument is a string value that specifies the pipe's input. The default input (typically the output of a previous pipe) will be discarded.

```yaml
private:
    some-pipe:
        input: "This is the new input"
        shell:
            run: some-shell-command
```

