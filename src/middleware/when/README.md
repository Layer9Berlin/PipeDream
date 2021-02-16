# `when` - Conditional Executor

The `when` middleware makes the execution of a pipe dependent on the evaluation of a boolean expression.

## Arguments

It takes a single string argument that should evaluate to a boolean expression. Please refer to [the GoValuate documentation](https://github.com/Knetic/govaluate#what-operators-and-types-does-this-support) for a list of supported operators and types.

> Since pipedream uses GoValuate under the hood, it makes no distinction between single and double quotes in boolean expressions ([see this issue](https://github.com/Knetic/govaluate/issues/94)).

```yaml
private:
    some-pipe:
        # this pipe will not be executed
        arg: fail
        when: "@{arg} == 'success'"

        # this pipe will be executed
        arg: successful
        # regexes can be used for matching
        when: "@{arg} =~ 'success'"
```

> Note that the behavior of `when` in child pipes can be a little confusing when argument interpolation is used. Interpolation of the child's invocation arguments takes place at parent invocation time, so values passed to the child might be ignored.

```yaml
private:
    parent-pipe:
        arg: fail
        pipe:
            - child-pipe:
                arg: success
                # this pipe will not be executed, because the argument is interpolated at parent invocation time
                # using the parent's argument value "fail"
                when: "@{arg} == 'success'"
            - another-child-pipe:
                arg: success

    another-child-pipe:
        # this pipe will be executed, because the argument is interpolated at child invocation time
        # using the child's argument value "success"
        when: "@{arg} == 'success'"
```

