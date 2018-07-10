package stdlib

import "github.com/niolabs/gonio-framework"

func SetTerminal(terminal *nio.Terminal, defaultValue nio.Terminal) {
	if *terminal == "" {
		*terminal = defaultValue
	}
}
