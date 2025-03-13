package labor

type Job struct {
	Name     string
	Data     any
	Pipeline Pipeline
}

type Pipeline struct {
	Sequence []Process
	Data     any
}
