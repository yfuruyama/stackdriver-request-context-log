package stackdriver

type contextKey struct{}

var (
	contextLoggerKey = &contextKey{}
)
