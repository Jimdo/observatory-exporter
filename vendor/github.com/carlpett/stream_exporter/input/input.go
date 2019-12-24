package input

import (
	"errors"
	"fmt"
)

type StreamInput interface {
	StartStream(ch chan<- string)
}

var inputTypes = make(map[string]func() (StreamInput, error))

func registerInput(inputType string, factory func() (StreamInput, error)) {
	inputTypes[inputType] = factory
}

func GetAvailableInputs() []string {
	inputs := make([]string, 0, len(inputTypes))
	for key := range inputTypes {
		inputs = append(inputs, key)
	}
	return inputs
}

func NewInput(inputType string) (StreamInput, error) {
	factory, registered := inputTypes[inputType]
	if !registered {
		return nil, errors.New(fmt.Sprintf("Unknown input type %v", inputType))
	}
	return factory()
}
