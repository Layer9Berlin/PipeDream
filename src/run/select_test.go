package run

import (
	"bytes"
	"github.com/Layer9Berlin/pipedream/src/middleware"
	pipeline2 "github.com/Layer9Berlin/pipedream/src/pipeline"
	"github.com/stretchr/testify/require"
	"io"
	"io/ioutil"
	"os"
	"sync"
	"testing"
)

func TestSelect_letUserSelectPipelineFile_noFiles(t *testing.T) {
	executionContext := &middleware.ExecutionContext{
		PipelineFiles: nil,
	}
	file, err := letUserSelectPipelineFile(executionContext, 10, os.Stdin, os.Stdout)
	require.NotNil(t, err)
	require.Nil(t, file)
}

func TestSelect_letUserSelectPipelineFile_singleFile(t *testing.T) {
	testFile := pipeline2.File{Path: "test"}
	executionContext := middleware.NewExecutionContext()
	executionContext.PipelineFiles = []pipeline2.File{
		testFile,
	}
	executionContext.SelectableFiles = []string{
		"test",
	}
	file, err := letUserSelectPipelineFile(
		executionContext,
		10,
		os.Stdin,
		os.Stdout,
	)
	require.Nil(t, err)
	require.NotNil(t, file)
	require.Equal(t, "test", file.Path)
}

func TestSelect_letUserSelectPipelineFile_userSelection(t *testing.T) {
	testFile1 := pipeline2.File{Path: "test1.pipe", FileName: "Test1"}
	testFile2 := pipeline2.File{Path: "test2.pipe", FileName: "Test2"}
	executionContext := middleware.NewExecutionContext()
	executionContext.PipelineFiles = []pipeline2.File{
		testFile1,
		testFile2,
	}
	executionContext.SelectableFiles = []string{
		"test1.pipe",
		"test2.pipe",
	}
	reader, writer := io.Pipe()
	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(1)
	go func() {
		result, _ := ioutil.ReadAll(reader)
		require.Contains(t, string(result), "Test1")
		waitGroup.Done()
	}()
	file, err := letUserSelectPipelineFile(
		executionContext,
		10,
		ioutil.NopCloser(bytes.NewBuffer([]byte{10, 0})),
		writer)
	go func() {
		_ = writer.Close()
	}()
	waitGroup.Wait()
	require.Nil(t, err)
	require.NotNil(t, file)
	require.Equal(t, "test1.pipe", file.Path)
}

func TestSelect_letUserSelectPipelineFile_fileSelectionError(t *testing.T) {
	testFile1 := pipeline2.File{FileName: "test1.pipe"}
	testFile2 := pipeline2.File{FileName: "test2.pipe"}
	executionContext := middleware.NewExecutionContext()
	executionContext.PipelineFiles = []pipeline2.File{
		testFile1,
		testFile2,
	}
	executionContext.SelectableFiles = []string{
		"test1.pipe",
		"test2.pipe",
	}
	reader, writer := io.Pipe()
	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(1)
	go func() {
		_, _ = ioutil.ReadAll(reader)
		waitGroup.Done()
	}()
	_, err := letUserSelectPipelineFile(
		executionContext,
		-1,
		ioutil.NopCloser(bytes.NewBuffer([]byte{10, 0})),
		writer)
	go func() {
		_ = writer.Close()
	}()
	waitGroup.Wait()
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "prompt failed")
}

func TestSelect_letUserSelectPipelineFileAndPipeline_noFiles(t *testing.T) {
	executionContext := middleware.NewExecutionContext()
	executionContext.PipelineFiles = []pipeline2.File{}
	pipeline, file, err := letUserSelectPipelineFileAndPipeline(
		executionContext,
		10,
		os.Stdin,
		os.Stdout)
	require.Nil(t, err)
	require.Equal(t, "no-pipelines::handle", pipeline)
	require.Equal(t, "", file)
}

func TestSelect_letUserSelectPipelineFileAndPipeline_noSelectableFiles(t *testing.T) {
	testFile1 := pipeline2.File{FileName: "test1.pipe"}
	executionContext := middleware.NewExecutionContext()
	executionContext.PipelineFiles = []pipeline2.File{
		testFile1,
	}
	_, _, err := letUserSelectPipelineFileAndPipeline(
		executionContext,
		10,
		os.Stdin,
		os.Stdout)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "no pipeline file found")
}

func TestSelect_letUserSelectPipelineFileAndPipeline_pipelineSelectionError(t *testing.T) {
	testFile1 := pipeline2.File{FileName: "test1.pipe", Path: "test1.pipe"}
	executionContext := middleware.NewExecutionContext()
	executionContext.PipelineFiles = []pipeline2.File{
		testFile1,
	}
	executionContext.SelectableFiles = []string{
		"test1.pipe",
	}
	_, _, err := letUserSelectPipelineFileAndPipeline(
		executionContext,
		-1,
		os.Stdin,
		os.Stdout)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "prompt failed")
}

func TestSelect_letUserSelectPipelineFileAndPipeline_singleFile(t *testing.T) {
	testFile1 := pipeline2.File{
		FileName: "test1.pipe",
		Path:     "test1.pipe",
		Public: map[string]map[string]interface{}{
			"test": nil,
		},
	}
	executionContext := middleware.NewExecutionContext()
	executionContext.Definitions = map[string][]pipeline2.Definition{
		"test": {
			{
				DefinitionArguments: map[string]interface{}{
					"selectable": true,
				},
			},
		},
	}
	executionContext.PipelineFiles = []pipeline2.File{
		testFile1,
	}
	executionContext.SelectableFiles = []string{
		"test1.pipe",
	}
	pipeline, file, err := letUserSelectPipelineFileAndPipeline(
		executionContext,
		10,
		os.Stdin,
		os.Stdout,
	)
	require.Nil(t, err)
	require.Equal(t, "test", pipeline)
	require.Equal(t, "test1.pipe", file)
}

func TestSelect_letUserSelectPipelineFileAndPipeline_singlePipeline(t *testing.T) {
	testFile1 := pipeline2.File{
		FileName: "test1.pipe",
		Path:     "test1.pipe",
		Public: map[string]map[string]interface{}{
			"test_public": nil,
		},
		Private: map[string]map[string]interface{}{
			"test_private": nil,
		},
	}
	executionContext := middleware.NewExecutionContext()
	executionContext.PipelineFiles = []pipeline2.File{
		testFile1,
	}
	executionContext.SelectableFiles = []string{
		"test1.pipe",
	}
	pipeline, file, err := letUserSelectPipelineFileAndPipeline(
		executionContext,
		10,
		os.Stdin,
		os.Stdout,
	)
	require.Nil(t, err)
	require.Equal(t, "test_public", pipeline)
	require.Equal(t, "test1.pipe", file)
}

func TestSelect_letUserSelectPipelineFileAndPipeline_userSelectsPipeline(t *testing.T) {
	testFile1 := pipeline2.File{
		FileName: "test1.pipe",
		Path:     "test1.pipe",
		Public: map[string]map[string]interface{}{
			"test_public_1": nil,
			"test_public_2": nil,
		},
		Private: map[string]map[string]interface{}{
			"test_private": nil,
		},
	}
	executionContext := middleware.NewExecutionContext()
	executionContext.PipelineFiles = []pipeline2.File{
		testFile1,
	}
	executionContext.SelectableFiles = []string{
		"test1.pipe",
	}
	reader, writer := io.Pipe()
	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(1)
	go func() {
		_, _ = ioutil.ReadAll(reader)
		waitGroup.Done()
	}()
	pipeline, file, err := letUserSelectPipelineFileAndPipeline(
		executionContext,
		10,
		ioutil.NopCloser(bytes.NewBuffer([]byte{10, 0})),
		writer)
	go func() {
		_ = writer.Close()
	}()
	waitGroup.Wait()
	require.Nil(t, err)
	require.Equal(t, "test_public_1", pipeline)
	require.Equal(t, "test1.pipe", file)
}

func TestSelect_letUserSelectPipelineFileAndPipeline_defaultPreselection(t *testing.T) {
	testFile1 := pipeline2.File{
		Default: pipeline2.DefaultSettings{
			Command: "test_public_2",
		},
		FileName: "test1.pipe",
		Path:     "test1.pipe",
		Public: map[string]map[string]interface{}{
			"test_public_1": nil,
			"test_public_2": nil,
		},
		Private: map[string]map[string]interface{}{
			"test_private": nil,
		},
	}
	executionContext := middleware.NewExecutionContext()
	executionContext.PipelineFiles = []pipeline2.File{
		testFile1,
	}
	executionContext.SelectableFiles = []string{
		"test1.pipe",
	}
	reader, writer := io.Pipe()
	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(1)
	go func() {
		_, _ = ioutil.ReadAll(reader)
		waitGroup.Done()
	}()
	pipeline, file, err := letUserSelectPipelineFileAndPipeline(
		executionContext,
		10,
		ioutil.NopCloser(bytes.NewBuffer([]byte{10, 0})),
		writer)
	go func() {
		_ = writer.Close()
	}()
	waitGroup.Wait()
	require.Nil(t, err)
	require.Equal(t, "test_public_2", pipeline)
	require.Equal(t, "test1.pipe", file)
}

func TestSelect_letUserSelectPipelineFileAndPipeline_pipelineFlag(t *testing.T) {
	PipelineFlag = "test_public_2"
	testFile1 := pipeline2.File{
		FileName: "test1.pipe",
		Path:     "test1.pipe",
		Public: map[string]map[string]interface{}{
			"test_public_1": nil,
			"test_public_2": nil,
		},
		Private: map[string]map[string]interface{}{
			"test_private": nil,
		},
	}
	executionContext := middleware.NewExecutionContext()
	executionContext.PipelineFiles = []pipeline2.File{
		testFile1,
	}
	executionContext.SelectableFiles = []string{
		"test1.pipe",
	}
	reader, writer := io.Pipe()
	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(1)
	go func() {
		_, _ = ioutil.ReadAll(reader)
		waitGroup.Done()
	}()
	pipeline, file, err := letUserSelectPipelineFileAndPipeline(
		executionContext,
		10,
		ioutil.NopCloser(bytes.NewBuffer([]byte{10, 0})),
		writer)
	go func() {
		_ = writer.Close()
	}()
	waitGroup.Wait()
	require.Nil(t, err)
	require.Equal(t, "test_public_2", pipeline)
	require.Equal(t, "test1.pipe", file)
}

func TestSelect_letUserSelectPipeline_fileFlag(t *testing.T) {
	FileFlag = "test.file"
	testFile1 := pipeline2.File{
		FileName: "test.file",
		Path:     "test.file",
		Public: map[string]map[string]interface{}{
			"test_public_1": nil,
			"test_public_2": nil,
		},
		Private: map[string]map[string]interface{}{
			"test_private": nil,
		},
	}
	executionContext := middleware.NewExecutionContext()
	executionContext.PipelineFiles = []pipeline2.File{
		testFile1,
	}
	reader, writer := io.Pipe()
	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(1)
	go func() {
		_, _ = ioutil.ReadAll(reader)
		waitGroup.Done()
	}()
	file, err := letUserSelectPipelineFile(
		executionContext,
		10,
		ioutil.NopCloser(bytes.NewBuffer([]byte{10, 0})),
		writer)
	go func() {
		_ = writer.Close()
	}()
	waitGroup.Wait()
	require.Nil(t, err)
	require.Equal(t, testFile1, *file)
}
