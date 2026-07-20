package app

// Progress desacopla app.Run* da apresentação (CLI ANSI ou TUI).
// Migração incremental: quando nil, runner usa ui.Session.
type Progress interface {
	Step(label string, fn func() error) error
	StepQuiet(fn func() error) error
	Detail(msg string)
	Info(msg string)
	Warn(msg string)
	Success(msg string)
}

// NopProgress discards presentation output (desktop bindings, tests).
func NopProgress() Progress {
	return nopProgress{}
}

type nopProgress struct{}

func (nopProgress) Step(_ string, fn func() error) error { return fn() }
func (nopProgress) StepQuiet(fn func() error) error     { return fn() }
func (nopProgress) Detail(string)                       {}
func (nopProgress) Info(string)                         {}
func (nopProgress) Warn(string)                         {}
func (nopProgress) Success(string)                      {}
