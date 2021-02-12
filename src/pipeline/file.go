package pipeline

// File is a representation of a yaml pipeline file as a Go struct
//
// Note that this struct does not have exactly the same structure as the yaml file
type File struct {
	Default  DefaultSettings
	Path     string
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

// FileImportSkeleton is a very basic representation of a yaml pipeline file concerned only with import declarations
type FileImportSkeleton struct {
	Import []string
}

// DefaultSettings are file-level options (to be refined in future)
type DefaultSettings struct {
	Command string
	Dir     string
}
