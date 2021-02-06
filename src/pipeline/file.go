// Provides a representation of a yaml pipeline file as a Go struct
package pipeline

// Note that this struct does not have exactly the same structure as the yaml file
type File struct {
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

type PipelineFileImportSkeleton struct {
	Import []string
}

type DefaultSettings struct {
	Command string
	Dir     string
}
