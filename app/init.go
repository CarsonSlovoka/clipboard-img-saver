package app

const (
	Name    string = "Clipboard Image saver"
	ExeName string = "cis"
	Version string = "v0.1.0"
)

func About() string {
	return Name + "(" + Version + ")"
}
