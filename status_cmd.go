package qrev

import "fmt"

type StatusCmd struct {
}

func (cmd *StatusCmd) Run(options *Options) error {
	fmt.Println("status") // TODO:
	return nil
}
