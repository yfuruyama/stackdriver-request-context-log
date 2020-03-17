package stackdriverlog

type contextKey struct{}

var (
	contextLoggerKey = &contextKey{}
)
