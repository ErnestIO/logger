package adapters

// Adapter : interface for Logger adapters
type Adapter interface {
	Manage([]string, MessageProcessor) error
	Stop()
	Name() string
}
