package atlas

import (
	"fmt"
	"io"
	"strings"
)

var (
	buildVersion   = "dev"
	buildSourceSHA = "unknown"
)

type rootCommand struct {
	name string
	run  func([]string, io.Writer) error
}

func rootCommandRegistry() []rootCommand {
	return []rootCommand{
		{name: "instance", run: runInstance},
		{name: "intake", run: runIntake},
		{name: "blueprint", run: runBlueprint},
		{name: "mission", run: runMission},
		{name: "blueprint-request", run: runBlueprintRequest},
		{name: "workgraph", run: runWorkgraph},
		{name: "mutation-classes", run: runMutationClasses},
		{name: "factory-task", run: runFactoryTask},
		{name: "factory", run: runFactory},
		{name: "context-pack", run: runContextPack},
		{name: "foundry", run: runFoundry},
		{name: "run-link", run: runRunLink},
		{name: "terminal-index", run: runTerminalIndex},
	}
}

func rootCommandNames() []string {
	commands := rootCommandRegistry()
	names := make([]string, 0, len(commands))
	for _, command := range commands {
		names = append(names, command.name)
	}
	return names
}

func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		usage(stderr)
		return 2
	}
	if args[0] == "--version" {
		if len(args) != 1 {
			fmt.Fprintln(stderr, "error: --version does not accept arguments")
			return 2
		}
		fmt.Fprintf(stdout, "ao-atlas version=%s source_sha=%s\n", buildVersion, buildSourceSHA)
		return 0
	}
	var err error
	matched := false
	for _, command := range rootCommandRegistry() {
		if args[0] == command.name {
			matched = true
			err = command.run(args[1:], stdout)
			break
		}
	}
	if !matched {
		err = fmt.Errorf("unknown command %q", args[0])
	}
	if err != nil {
		fmt.Fprintln(stderr, "error:", err)
		return 1
	}
	return 0
}

func usage(w io.Writer) {
	fmt.Fprintf(w, "atlas <%s> ...\n", strings.Join(rootCommandNames(), "|"))
}
