package core

func Version() string {
	return "0.1.0"
}

func Not[F ~func(T) bool, T any](pred F) func(T) bool {
	return func(t T) bool {
		return !pred(t)
	}
}
