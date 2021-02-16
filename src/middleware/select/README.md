# `select` - User Selection Prompter

The `select` middleware shows a prompt in the terminal, allowing the user to make a selection from a pre-defined list of pipes. The selected pipe will the be executed.

## Arguments

### Options

The `options` argument is a list of pipeline references (either strings or maps) that the user will be able to choose from. The string shown to the user is either the contents of the `description` argument (if present and a string) or the pretty-printed identifier of the pipe.

### Prompt

The `prompt` argument is an optional string shown to the user before selection. The default value is `Please select an option`.

### Initial

The `initial` argument is an optional integer indicating the zero-based index of the initially selected item (before the user has pressed any buttons).

```yaml
private:
    some-pipe:
        select:
            prompt: "Pick your poison!"
            initial: 1
            options:
                - some-pipe::option-1
                - some-pipe::option-2:
                    arg: value
                - some-pipe::option-3:
                    description "Choose me!"
```

will result in a prompt such as the following:

```console
Pick your poison!
  Some Pipe > Child Pipe 1
> Some Pipe > Child Pipe 2
  Choose me!
```
