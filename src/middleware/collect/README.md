# `collect` - Data Aggregator

## Arguments

The `collect` middleware aggregates the results of several pipes into a map, optionally saving them to a file.

### `values` Argument

The `values` argument is an array of pipeline references, whose items are called *children*. The calling pipe is called the *parent*.

Each child will be executed and its result is stored in a map whose key is the child's identifier. The resulting map will either be saved to a file or comprises the parent's output.

```yaml
private:
    collect:
        values:
            - some-pipe:
                shell:
                    run: "echo \"test1\""
            - some-other-pipe:
                shell:
                    run: "echo \"test2\""
```

results in the following parent output:

```yaml
some-pipe: test1
some-other-pipe: test2
```

### `nested` Argument

The `nested` argument is an optional boolean that will cause the output of the children to be included as a sub-map - as opposed to a string.

```yaml
private:
    collect:
        values:
            - key:
                collect:
                    values:
                        - sub-key:
                            shell:
                                run: "echo \"test1\""
```

results in the following parent output:

```yaml
key:
    sub-key: test1
```

### `file` Argument

The `file` argument is an optional file path relative to the current working directory to which the parent's output will be redirected, if provided.
