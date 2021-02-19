package pipeline

type DataConnection struct {
	Source *Run
	Target *Run
	Label  *string
}

func NewDataConnection(source *Run, target *Run, label string) *DataConnection {
	return &DataConnection{
		Source: source,
		Target: target,
		Label:  &label,
	}
}
