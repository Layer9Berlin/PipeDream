package models

type PipelineFileImportSkeleton struct {
	Import []string
}

type PipelineFile struct {
	Default  DefaultSettings
	FileName string
	Hooks    HookDefinitions
	Import   []string
	// in pipeline files, each command can have arbitrary parameters
	// it may also have "steps"
	// each step can be either a string referencing another pipeline
	// or a dictionary containing additional parameters
	Public  map[string]map[string]interface{}
	Private map[string]map[string]interface{}
}

type DefaultSettings struct {
	Command string
	Dir     string
}
