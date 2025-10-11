package qrev

import "fmt"

type MarkCmd struct {
	Status string `arg:"" enum:"skip,fail" help:"TODO"`
}

func (cmd *MarkCmd) Run(options *Options) error {
	fmt.Println("TODO") // TODO:
	return nil
}
