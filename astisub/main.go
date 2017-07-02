package main

import (
	"flag"

	"github.com/asticode/go-astilog"
	"github.com/asticode/go-astisub"
	"github.com/asticode/go-astitools/flag"
)

// Flags
var (
	fragmentDuration = flag.Duration("f", 0, "the fragment duration")
	inputPath        = astiflag.Strings{}
	outputPath       = flag.String("o", "", "the output path")
	syncDuration     = flag.Duration("s", 0, "the sync duration")
)

func main() {
	// Init
	var s = astiflag.Subcommand()
	flag.Var(&inputPath, "i", "the input paths")
	flag.Parse()
	astilog.SetLogger(astilog.New(astilog.FlagConfig()))

	// Validate input path
	if len(inputPath) == 0 {
		astilog.Fatal("Use -i to provide at least one input path")
	}

	// Validate output path
	if len(*outputPath) <= 0 {
		astilog.Fatal("Use -o to provide an output path")
	}

	// Open first input path
	var sub *astisub.Subtitles
	var err error
	if sub, err = astisub.Open(inputPath[0]); err != nil {
		astilog.Fatalf("%s while opening %s", err, inputPath[0])
	}

	// Switch on subcommand
	switch s {
	case "convert":
		// Write
		if err = sub.Write(*outputPath); err != nil {
			astilog.Fatalf("%s while writing to %s", err, *outputPath)
		}
	case "fragment":
		// Validate fragment duration
		if *fragmentDuration <= 0 {
			astilog.Fatal("Use -f to provide a fragment duration")
		}

		// Fragment
		sub.Fragment(*fragmentDuration)

		// Write
		if err = sub.Write(*outputPath); err != nil {
			astilog.Fatalf("%s while writing to %s", err, *outputPath)
		}
	case "merge":
		// Validate second input path
		if len(inputPath) == 1 {
			astilog.Fatal("Use -i to provide at least two input paths")
		}

		// Open second input path
		var sub2 *astisub.Subtitles
		if sub2, err = astisub.Open(inputPath[1]); err != nil {
			astilog.Fatalf("%s while opening %s", err, inputPath[1])
		}

		// Merge
		sub.Merge(sub2)

		// Write
		if err = sub.Write(*outputPath); err != nil {
			astilog.Fatalf("%s while writing to %s", err, *outputPath)
		}
	case "sync":
		// Validate sync duration
		if *syncDuration == 0 {
			astilog.Fatal("Use -s to provide a sync duration")
		}

		// Fragment
		sub.Add(*syncDuration)

		// Write
		if err = sub.Write(*outputPath); err != nil {
			astilog.Fatalf("%s while writing to %s", err, *outputPath)
		}
	case "unfragment":
		// Unfragment
		sub.Unfragment()

		// Write
		if err = sub.Write(*outputPath); err != nil {
			astilog.Fatalf("%s while writing to %s", err, *outputPath)
		}
	default:
		astilog.Fatalf("Invalid subcommand %s", s)
	}
}
