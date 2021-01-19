package with

import (
	"bytes"
	"io/ioutil"
	"pipedream/src/logging/log_fields"
	"pipedream/src/middleware"
	"pipedream/src/models"
	"regexp"
)

// Pattern Extractor
type WithMiddleware struct {
}

func (withMiddleware WithMiddleware) String() string {
	return "with"
}

func NewWithMiddleware() WithMiddleware {
	return WithMiddleware{}
}

type withMiddlewareArguments struct {
	Pattern string
}

func (withMiddleware WithMiddleware) Apply(
	run *models.PipelineRun,
	next func(*models.PipelineRun),
	executionContext *middleware.ExecutionContext,
) {
	argument := withMiddlewareArguments{
		Pattern: "",
	}
	middleware.ParseArguments(&argument, "with", run)
	next(run)

	if argument.Pattern != "" {
		run.Log.DebugWithFields(
			log_fields.Symbol("／"),
			log_fields.Middleware(withMiddleware),
			log_fields.Message("pattern"),
			log_fields.Info(argument.Pattern),
		)

		regex, err := regexp.Compile(argument.Pattern)
		if err != nil {
			run.Log.Error(err, log_fields.Middleware(withMiddleware))
			return
		}

		run.Log.TraceWithFields(
			log_fields.Symbol("⎇"),
			log_fields.Message("copying stdin"),
			log_fields.Middleware(withMiddleware),
		)
		stdinCopy := run.Stdin.Copy()
		run.Log.TraceWithFields(
			log_fields.Symbol("⎇"),
			log_fields.Message("intercepting stdout"),
			log_fields.Middleware(withMiddleware),
		)
		stdoutIntercept := run.Stdout.Intercept()
		run.Log.TraceWithFields(
			log_fields.Symbol("⎇"),
			log_fields.Message("creating stderr writer"),
			log_fields.Middleware(withMiddleware),
		)
		stderrAppender := run.Stderr.WriteCloser()
		go func() {
			completeInput, err := ioutil.ReadAll(stdinCopy)
			run.Log.PossibleError(err)
			matches := regex.FindAllSubmatch(completeInput, -1)
			for _, match := range matches {
				fullOutput := new(bytes.Buffer)
				executionContext.FullRun(
					middleware.WithParentRun(run),
					middleware.WithIdentifier(run.Identifier),
					middleware.WithArguments(run.ArgumentsCopy()),
					middleware.WithSetupFunc(func(matchRun *models.PipelineRun) {
						run.Log.TraceWithFields(
							log_fields.Symbol("⎇"),
							log_fields.Message("replacing stdin with regex match"),
							log_fields.Middleware(withMiddleware),
						)
						matchRun.Stdin.Replace(bytes.NewBuffer(match[0]))
					}),
					middleware.WithTearDownFunc(func(matchRun *models.PipelineRun) {
						run.Log.TraceWithFields(
							log_fields.Symbol("⎇"),
							log_fields.Message("copying child stdout into parent stdout"),
							log_fields.Middleware(withMiddleware),
						)
						matchRun.Stdout.StartCopyingInto(fullOutput)
						run.Log.TraceWithFields(
							log_fields.Symbol("⎇"),
							log_fields.Message("copying child stderr into parent stderr"),
							log_fields.Middleware(withMiddleware),
						)
						matchRun.Stderr.StartCopyingInto(stderrAppender)

						matchRun.Close()
						matchRun.Wait()
						completeStdout := matchRun.Stdout.Bytes()
						completeInput = bytes.Replace(completeInput, match[0], completeStdout, -1)
					}))
			}
			_, err = stdoutIntercept.Write(completeInput)
			run.Log.PossibleError(err)
			run.Log.PossibleError(stdoutIntercept.Close())
			run.Log.PossibleError(stderrAppender.Close())
		}()
	} else {
		next(run)
	}
}
