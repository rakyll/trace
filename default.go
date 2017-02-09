package trace

// type dc struct {
// 	family string
// }

// // DefaultClient returns a backend that traces and provides a web UI on /debug/requests and /debug/events.
// // These paths are registered on the default mux, start a server in order to be able to see them.
// func DefaultClient(family string) Client {
// 	return &dc{family: family}
// }

// func (d *dc) NewSpan(parent context.Context, casual []byte, name string) context.Context {
// 	// TODO(jbd): Handle parent/child and casual relationships.
// 	tr := trace.New(d.family, name)
// 	return context.WithValue(parent, defaultTraceKey, tr)
// }

// func (d *dc) Span(ctx context.Context) context.Context {
// }

// func (d *dc) Info(ctx context.Context) []byte {
// 	return []byte("") // not provided by the default client
// }

// func (d *dc) Finish(ctx context.Context, labels map[string]interface{}) error {
// 	v := ctx.Value(defaultTraceKey)
// 	if v == nil {
// 		return nil
// 	}
// 	tr := v.(trace.Trace)
// 	tr.Finish()
// 	return nil
// }

// type stringer struct {
// 	format string
// 	args   []interface{}
// }

// func (s *stringer) String() string {
// 	return fmt.Sprintf(s.format, s.args...)
// }

// var defaultTraceKey = contextKey("defaultTrace")
