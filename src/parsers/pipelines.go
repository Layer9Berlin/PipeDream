package parsers

import (
	"fmt"
	"path"
	"path/filepath"
)

func (parser *Parser) BuiltInPipelineFilePaths(projectPath string) ([]string, error) {
	glob := path.Join(projectPath, "pipes/**/*.pipe")
	matches, err := parser.findByGlob(glob)
	if err != nil {
		return []string{}, fmt.Errorf("failed to glob pipeline files: %v", err)
	}
	if len(matches) == 0 {
		return []string{}, fmt.Errorf("no built-in pipeline files found, please double-check your installation")
	}
	return matches, nil
}

func (parser *Parser) UserPipelineFilePaths(args []string) ([]string, error) {
	pipelineFilePaths := make([]string, 0, 10)
	if len(args) == 1 {
		pipelineName := args[0]
		extension := filepath.Ext(args[0])
		if extension == "" {
			pipelineName = pipelineName + ".pipe"
		}
		pipelineFilePaths = append(pipelineFilePaths, pipelineName)
	} else {
		matches, err := parser.findByGlob("*.pipe")
		if err != nil {
			return []string{}, fmt.Errorf("failed to glob pipeline files: %v", err)
		}
		if len(matches) == 0 {
			return []string{}, fmt.Errorf("no pipeline files found, please switch directories or provide path argument")
		}
		pipelineFilePaths = matches
	}
	return pipelineFilePaths, nil
}
