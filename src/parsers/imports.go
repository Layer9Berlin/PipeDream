package parsers

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"pipedream/src/models"
	"sort"
)

func (parser *Parser) recursiveImportStep(unprocessedFilePaths []string, processedFilePaths []string) ([]string, error) {
	sort.Strings(unprocessedFilePaths)

	additionalFilePaths := make([]string, 0, 10)
	for _, filePath := range unprocessedFilePaths {
		fileData, err := parser.readFile(filePath)
		if err != nil {
			return nil, err
		}
		importSkeleton := models.PipelineFileImportSkeleton{}
		err = yaml.Unmarshal(fileData, &importSkeleton)
		if err != nil {
			return nil, fmt.Errorf("error parsing %q: %w", filePath, err)
		}
		for _, additionalFilePath := range importSkeleton.Import {
			i := sort.SearchStrings(unprocessedFilePaths, additionalFilePath)
			if !(i < len(unprocessedFilePaths) && unprocessedFilePaths[i] == additionalFilePath) {
				additionalFilePaths = append(additionalFilePaths, additionalFilePath)
			}
		}
	}

	processedFilePaths = append(processedFilePaths, unprocessedFilePaths...)
	if len(additionalFilePaths) > 0 {
		return parser.recursiveImportStep(additionalFilePaths, processedFilePaths)
	}
	return processedFilePaths, nil
}
