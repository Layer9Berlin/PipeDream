package with

import (
	"bytes"
	"github.com/Layer9Berlin/pipedream/src/logging/log_fields"
	"github.com/Layer9Berlin/pipedream/src/middleware"
	"github.com/Layer9Berlin/pipedream/src/models"
	"io/ioutil"
	"regexp"
	"sync"
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
		waitGroup := &sync.WaitGroup{}
		go func() {
			completeInput, err := ioutil.ReadAll(stdinCopy)
			inputMutex := &sync.Mutex{}
			run.Log.PossibleError(err)
			matches := regex.FindAllSubmatch(completeInput, -1)
			waitGroup.Add(len(matches))
			for _, match := range matches {
				match := match
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
							log_fields.Message("copying child stderr into parent stderr"),
							log_fields.Middleware(withMiddleware),
						)
						matchRun.Stderr.StartCopyingInto(stderrAppender)

						go func() {
							defer waitGroup.Done()
							matchRun.Wait()
							defer inputMutex.Unlock()
							inputMutex.Lock()
							completeStdout := matchRun.Stdout.Bytes()
							completeInput = bytes.Replace(completeInput, match[0], completeStdout, -1)
						}()
					}))
			}
			go func() {
				waitGroup.Wait()
				_, err = stdoutIntercept.Write(completeInput)
				run.Log.PossibleError(err)
				run.Log.PossibleError(stdoutIntercept.Close())
				run.Log.PossibleError(stderrAppender.Close())
			}()
		}()
	}
}
