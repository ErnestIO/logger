package adapters

// Adapter : interface for Logger adapters
type Adapter interface {
	Manage([]string, MessageProcessor) error
	Stop()
	Name() string
}

// MessageProcessor : Manage will receive this interface in order to
// process the input messages
type MessageProcessor func(string) string
