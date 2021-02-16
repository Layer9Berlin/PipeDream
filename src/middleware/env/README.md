# `env` - Environment Variable Manager

The `env` middleware manages the handling of environment variables.

## Arguments

It takes two arguments:

### Save
The `save` argument is an optional string that causes the result of the pipe's execution to be stored in an environment variable of that name (excluding the `$`). The output is not swallowed, but still passed on as usual. This behavior might change in a future version of pipedream.

```yaml
private:
    some-pipe:
        shell:
            run: "command"
        env:
            # the result of "command" will be saved in the environment as $ENV_VAR
            save: ENV_VAR
```

### Interpolate

The `interpolate` argument is a string that determines how environment variables will be interpolated when the pipe is invoked. Possible values are:
    - `deep`:
        Values located within a nested map of arguments will be interpolated, irrespective of the level of nesting.
    - `shallow`:
        Only string values with a single level of nesting will be interpolated at invocation time.
    - `none` (default):
        No interpolation takes place at invocation time.

> Note that environment variables that are not interpolated by the `env` middleware may still be evaluated at execution time, e.g. by the `shell` middleware. For this reason, you will only need to use this feature if interpolation should take place before execution - for example, to make conditional execution (using the `when` middleware) dependent on the value of an environment variable.

```yaml
private:
    some-pipe:
        env:
            # added for clarity, `none` is the default value
            # note that using `deep` could cause problems here,
            # if the value of ENV_VAR is not yet set at the time of parent invocation
            interpolate: none
        pipe:
            - child-pipe:
                env:
                    interpolate: shallow
                # this value will be evaluated at child invocation time
                # note that if the value is set by another pipe,
                # the child's invocation must be delayed
                # until after the env var has been set
                # e.g. using the `waitFor` middleware
                when: "'$ENV_VAR' == 'YES!!'"
```
