# Middleware

Each middleware implements a function with the following signature

```
func (middleware CatchMiddleware) Run(
	invocation *models.PipelineInvocation,
	result *models.RunResult,
	next func(*models.PipelineInvocation, *models.RunResult),
	logger *logging.MiddlewareLogger,
	stack *middleware.Stack) {}
```



## Built-in middleware


### [`args` - Shell Arguments Appender](./args)



### [`timer` - Execution timing](./timer)

### [`pipe` - Child invoker](./timer)

### [`pipe` - Child invoker](./timer)


## Writing your own middleware

