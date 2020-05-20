package worker

import "os"

type Info struct {
	Name         string
	Queue        string
	ProcessQueue string
}

func NewInfo() Info {
	name, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	queue := "runs:work"
	processQueue := "runs:worker:" + name
	return Info{name, queue, processQueue}
}
