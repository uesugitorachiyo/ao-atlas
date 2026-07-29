package atlas

import (
	"flag"
	"fmt"
	"io"
)

func runTerminalIndex(args []string, stdout io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("terminal-index requires build or verify")
	}
	switch args[0] {
	case "build":
		fs := flag.NewFlagSet("terminal-index build", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		root := fs.String("root", "", "evidence root")
		manifest := fs.String("manifest", "", "lineage manifest")
		out := fs.String("out", "", "canonical index output")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if *root == "" || *manifest == "" || *out == "" {
			return fmt.Errorf("--root, --manifest, and --out are required")
		}
		index, err := BuildCanonicalTerminalIndex(*root, *manifest)
		if err != nil {
			return err
		}
		if err := WriteJSON(*out, index); err != nil {
			return err
		}
		fmt.Fprintf(stdout, "status=written\nreadiness_passed=%t\ncompletion_observed=%t\ndigest=%s\n", index.ReadinessPassed, index.CompletionObserved, index.Digest)
		return nil
	case "verify":
		fs := flag.NewFlagSet("terminal-index verify", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		root := fs.String("root", "", "evidence root")
		indexPath := fs.String("index", "", "canonical index path")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if *root == "" || *indexPath == "" {
			return fmt.Errorf("--root and --index are required")
		}
		data, err := readBoundedRegularFile(*indexPath, canonicalTerminalArtifactMaxBytes)
		if err != nil {
			return err
		}
		var index CanonicalTerminalIndex
		if err := decodeStrictJSON(data, &index); err != nil {
			return err
		}
		if err := VerifyCanonicalTerminalIndex(*root, index); err != nil {
			return err
		}
		fmt.Fprintf(stdout, "status=valid\nreadiness_passed=%t\ncompletion_observed=%t\ndigest=%s\n", index.ReadinessPassed, index.CompletionObserved, index.Digest)
		return nil
	default:
		return fmt.Errorf("unknown terminal-index subcommand %q", args[0])
	}
}
