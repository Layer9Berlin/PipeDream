# `inherit` - Arguments Propagator


The `inherit` middleware manages the propagation of arguments from parent to child pipes.

## Arguments

It takes an array of string arguments, listing the arguments that should be inherited from the parent pipe.

> Note that inheritance currently considers only a single level of nesting. A grandchild pipe will not inherit arguments from the grandparent, unless these are explicitly inherited by the parent as well. This behavior may change in a future version of pipedream.

```yaml
private:
    some-pipe:
        argument-1: value-1
        argument-2: value-2
        argument-3: value-3
        pipe:
            - child-pipe:
                # in this case, the child pipe will inherit `argument-1`
                inherit:
                    - argument-1
                    - argument-3
                # the child's own arguments take precedence over inherited arguments
                # so `argument-3` with have value `other-value`, not `value-3`
                argument-3: other-value
```

