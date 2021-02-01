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

### [`catch` - Error Handler](./catch)
### [`dir` - Directory Navigator](./dir)
### [`docker` - Docker Executor](./docker)
### [`each` - Input Duplicator](./each)
### [`timer` - Directory Timer Middleware](./timer)


## Writing your own middleware

