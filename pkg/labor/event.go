package labor

type Event string

func (e Event) String() string {
	return string(e)
}
