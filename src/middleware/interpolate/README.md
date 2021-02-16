# `interpolate` - Argument Replacer

Pipedream has a dedicated argument interpolation mechanism in addition to the standard shell interpolation of environment variables (such as $EXAMPLE_ENV_VAR). It pulls values from the map of arguments provided to a pipe and uses the form `@{argument-name}` for value replacement and `@?{argument-name}` to check for the presence of an argument.

In addition, any occurrences of `@!!` will be replaced with the input to the pipe.

The `interpolate` middleware determines how the argument interpolation takes place. Nested arguments, i.e. maps cannot be interpolated, but strings, integers and arrays of strings are supported. String arrays will be turned into a single string, separated by new line (`\n`) characters.

> If any substitutions have been made by the interpolation middleware, a full run (through the entire middleware stack) will be executed for the respective pipe with the new arguments. This is to ensure both that each middleware has a chance to parse the new arguments and that the full input is available for interpolation, if required.

## Arguments

The `interpolate` middleware takes four arguments:

### Enable

The `enable` argument is a boolean value switch that turns the interpolation on (default) or off.

```yaml
private:
    some-pipe:
        interpolate:
            enable: false
        arg1: bar
        # this value will remain as is, without interpolation
        arg2: "foo: @{arg1}"

    some-other-pipe:
        # the interpolation is enabled by default
        arg1: bar
        # this value will change to `foo: bar`
        arg2: "foo: @{arg1}"
```

> Note that because interpolation is deep by default, inline arguments for nested child pipes will be interpolated at the time of invocation of the parent pipe. This may or may not be what you want. To delay the interpolation until the invocation of a child, it is often appropriate to use a separate definition for the child. Unlike the inline arguments, definition arguments are interpolated at the time of child invocation.

#### Inline Invocation
```yaml
private:
    some-pipe:
        arg1: "fizz"
        pipe:
            - child-1:
                arg1: bar
                # this value will become "foo: fizz", as it is interpolated at parent invocation time
                # using the parent's arguments
                arg2: "foo: @{arg1}"
                arg3: bar
                # in this case, interpolation at parent invocation time cause a warning:
                # "unable to find value for argument: `arg3`"
                # subsequent interpolation at child invocation time will change the value to "foo: bar"
                arg4: "foo: @{arg3}"
            - child-2:
                arg1: bar

    child-2:
        # this value will become "foo: fizz", as it is interpolated at invocation time
        # of the child, using the child's arguments
        arg2: "foo: @{arg1}"
```



### IgnoreWarnings

If an expression of the form `@{some-argument-key}` is encountered, but no suitable argument is found, a warning will be issued. This can be suppressed by setting the `ignoreWarnings` argument to `true`.

```yaml
private:
    some-pipe:
        # a warning will be issued here:
        # "unable to find value for argument: `arg3`"
        arg1: "@{not-present}"
        # this will not cause a warning
        # (presence of the argument is being checked and thus not expected)
        arg2: "@?{also-not-present}"
    some-pipe:
        interpolate:
            ignoreWarnings: true
        # no warning will be issued, as suppression is enabled
        arg1: "@{not-present}"
```

### Quote

The `Quote` argument determines whether substituted values should be quoted.

It has four valid values:

- `single` (default):
    Wrap each replacement in single quotes `'`, escaping any existing single quotes using `\'`.
- `double`:
    Wrap each replacement in double quotes `"`, escaping any existing double quotes using `\"`.
- `backticks`:
    Wrap each replacement in backticks `\``, escaping any existing backticks using `\\\``.
- `none`:
    Do not quote the replacements at all.


```yaml
private:
    single-quoting-pipe:
        interpolate:
            quote: single
        arg1: "\"value\" 'with' `quotes`"
        # will be interpolated to "'\"value\" \\'with\\' `quotes`'"
        arg2: "@{arg1}"

    double-quoting-pipe:
        interpolate:
            quote: double
        arg1: "\"value\" 'with' `quotes`"
        # will be interpolated to "\"\\\"value\\\" 'with' `quotes`\""
        arg2: "@{arg1}"

    backtick-quoting-pipe:
        interpolate:
            quote: backtick
        arg1: "\"value\" 'with' `quotes`"
        # will be interpolated to "`\"value\" 'with' \\`quotes\\``'"
        arg2: "@{arg1}"

    not-quoting-pipe:
        interpolate:
            quote: none
        arg1: "\"value\" 'with' `quotes`"
        # will be interpolated to "\"value\" 'with' `quotes`"
        arg2: "@{arg1}"
```


### EscapeQuotes

The `escapeQuote` argument can be used to escape quote characters in addition to the escaping performed during quoting (if any).

It has five valid values:

- `single` (default):
    Replace each single quote character `'` with `\'`.
- `double`:
    Replace each double quote character `"` with `\"`.
- `backticks`:
    Replace each backtick character `\`` with `\\\``.
- `all`:
    Perform all of the above.
- `none` (default):
    Do not replace any quotes.

```yaml
private:
    single-quote-escaping-pipe:
        interpolate:
            escapeQuotes: single
        arg1: "\"value\" 'with' `quotes`"
        # will be interpolated to "'\"value\" \\\\'with\\\\' `quotes`'"
        # (single quotes escaped once, then again when the entire string is single quoted)
        arg2: "@{arg1}"

    single-quote-escaping-pipe-without-quoting:
        interpolate:
            escapeQuotes: single
            quote: none
        arg1: "\"value\" 'with' `quotes`"
        # will be interpolated to "\"value\" \\'with\\' `quotes`"
        arg2: "@{arg1}"

    double-quote-escaping-pipe:
        interpolate:
            escapeQuotes: double
        arg1: "\"value\" 'with' `quotes`"
         # will be interpolated to "'\\\"value\\\" \\'with\\' `quotes`'"
       arg2: "@{arg1}"

    double-quote-escaping-pipe-without-quoting:
        interpolate:
            escapeQuotes: double
            quote: none
        arg1: "\"value\" 'with' `quotes`"
        # will be interpolated to "\\\"value\\\" 'with' `quotes`"
        arg2: "@{arg1}"

    backtick-escaping-pipe:
         interpolate:
             escapeQuotes: backtick
         arg1: "\"value\" 'with' `quotes`"
        # will be interpolated to "'\"value\" \\'with\\' \\`quotes\\`'"
         arg2: "@{arg1}"

    backtick-escaping-pipe-without-quoting:
         interpolate:
             escapeQuotes: backtick
             quote: none
        arg1: "\"value\" 'with' `quotes`"
        # will be interpolated to "\"value\" 'with' \\`quotes\\`"
        arg2: "@{arg1}"

```

## Nested interpolation

Nested argument interpolation is supported up to a nesting level of 5 (this behavior may change in a future version of pipedream). It is also legal to create an interpolation expression using interpolation itself, as in this example:

```yaml
private:
    some-pipe:
        interpolate:
 	        quote: "none"
        arg1: "value"
        # simple substitution
        arg2: "@{arg1}"
        # creating an interpolation expression using interpolation itself is legal
        arg4: "@{"
        arg5: "arg2"
        arg6: "}"
        # this will evaluate to "value"
        # after one step of interpolation, we have
        # arg2: "value"
        # arg3: @{arg2}
        # which is then interpolated to
        # arg2: "value"
        # arg3: "value"
        arg3: "@{arg4}@{arg5}@{arg6}"
```