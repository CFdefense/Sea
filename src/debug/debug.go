package debug

import "log"

type Debug struct {
	Name  string
	Debug bool
}

func InitializeDebugger(name string, debug bool) *Debug {
	return &Debug{
		name,
		debug,
	}
}

func (d Debug) DebugLog(msg string, fatal bool) {
	if d.Debug {
		if fatal {
			log.Fatalf("[%s] %s\n", d.Name, msg)
		}
		log.Printf("[%s] %s\n", d.Name, msg)
	}
}
