# `output` - Output Override


The `output` middleware manipulates the pipe's output, either processing it using another pipe or replacing it with a fixed value.

## Arguments

It takes either of two mutually exclusive string arguments:

### Process

The `process` argument is a string value containing the name of a pipe to be applied to the output.

```yaml
private:
    some-pipe:
        output: output-processor
        shell:
            run: some-shell-command

    output-processor:
        some: arguments
```

### Text

The `text` argument is a string value that specifies the pipe's output. If this argument is provided, the pipe's output will be discarded.

```yaml
private:
    some-pipe:
        input: "This is the new input"
        shell:
            run: some-shell-command
```
