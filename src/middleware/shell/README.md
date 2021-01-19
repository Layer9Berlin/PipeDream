# `args` - Shell Arguments Appender Middleware

The `args` middleware appends the values provided under the `args` key to a shell command about to be executed. This is mostly for convenience and clarity - arguments can also be provided to the command string directly, although this often results in long execution strings that are harder to test and annotate with meaningful comments.

> Note the difference between **shell arguments** (passed to the shell command) and **pipe arguments** (everything provided in the pipe invocation yaml, including the shell arguments under the `args` key).

## Arguments

The `args` middleware expects an array of values, each in one of the following formats (mixing formats is fine).

### Formats

#### Plain string

```yaml
- some_pipeline:
    args:
      - string_argument
```

#### Long form options

For convenience, options of the common form `--flag` and `--key=string_value` can be provided as

```yaml
- some_pipeline:
    args:
      - flag: true
      - unused_flag: false
      - key: string_value
```

This form is applied whenever the key does not start with a `-` and is not a single letter.

#### Short form options

If the key is a single letter, it will instead be appended to the shell command as `-a` (for boolean `true` value) or `-cstring_value` (for string value) without a space between key and value.

```yaml
- some_pipeline:
    args:
      - a: true
      - b: false
      - c: string_value
```

#### Explicit form

Keys that start with a `-` will be appended as is, with the value concatenated.

```yaml
- some_pipeline:
    args:
      - -a: string_value
```

results in `-astring_value` being appended.

This form can be used to enforce some custom formats:

```yaml
- some_pipeline:
    args:
      - "-a ": string_value
      - "-b=": string_value
      - "--c": string_value
      - "--d ": string_value
```

results in `-a string_value -b=string_value --string_value --d string_value` being appended to the shell command.