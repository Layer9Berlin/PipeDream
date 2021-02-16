package parsing

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"syscall"
)

// BuiltInPipelineFilePaths lists the file paths of all built-in pipelines
func (parser *Parser) BuiltInPipelineFilePaths(projectPath string) ([]string, error) {
	resolvedPipesPath, err := parser.evalSymlinks(path.Join(projectPath, "pipedream_pipes"))
	if err != nil {
		if pathErr, ok := err.(*os.PathError); ok {
			if pathErr.Err == syscall.ENOENT {
				resolvedPipesPath = "./include/pipedream_pipes"
			} else {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	glob := path.Join(resolvedPipesPath, "**/*.pipe")
	matches, err := parser.findByGlob(glob)
	if err != nil {
		return []string{}, fmt.Errorf("failed to glob pipeline files: %v", err)
	}
	if len(matches) == 0 {
		return []string{}, fmt.Errorf("no built-in pipeline files found, please double-check your installation")
	}
	return matches, nil
}

// UserPipelineFilePaths lists the file paths of all user-defined pipelines
func (parser *Parser) UserPipelineFilePaths(explicitFileName string) ([]string, error) {
	pipelineFilePaths := make([]string, 0, 10)
	if explicitFileName != "" {
		pipelineName := explicitFileName
		extension := filepath.Ext(pipelineName)
		if extension == "" {
			pipelineName = pipelineName + ".pipe"
		}
		pipelineFilePaths = append(pipelineFilePaths, pipelineName)
	} else {
		matches, err := parser.findByGlob("*.pipe")
		if err != nil {
			return []string{}, fmt.Errorf("failed to glob pipeline files: %v", err)
		}
		pipelineFilePaths = matches
	}
	return pipelineFilePaths, nil
}
