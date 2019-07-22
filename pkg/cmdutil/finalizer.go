package cmdutil

type (
	// A Finalizer is a function that should run before a program terminates.
	Finalizer func() error

	// Finalizers are a set of finalizers.
	Finalizers []Finalizer
)

// Run runs all the finalizers in fs.
func (fs Finalizers) Run(callbacks ...func(error)) {
	for i := len(fs) - 1; i >= 0; i-- {
		if err := fs[i](); err != nil {
			for _, cb := range callbacks {
				cb(err)
			}
		}
	}
}
