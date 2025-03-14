package labor

type Job struct {
	Name     string
	Sequence []Process
	Data     Pipeline
}

type Pipeline struct {
	Input  any
	Output any
}
