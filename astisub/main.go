package main

import (
	"flag"
	"log"

	"github.com/asticode/go-astikit"
	"github.com/asticode/go-astisub"
)

// Flags
var (
	fragmentDuration = flag.Duration("f", 0, "the fragment duration")
	inputPath        = astikit.NewFlagStrings()
	teletextPage     = flag.Int("p", 0, "the teletext page")
	outputPath       = flag.String("o", "", "the output path")
	syncDuration     = flag.Duration("s", 0, "the sync duration")
	syncTime         = flag.Duration("t", -1, "the sync time")
)

func main() {
	// Init
	cmd := astikit.FlagCmd()
	flag.Var(&inputPath, "i", "the input paths")
	flag.Parse()

	// Validate input path
	if len(*inputPath.Slice) == 0 {
		log.Fatal("Use -i to provide at least one input path")
	}

	// Validate output path
	if len(*outputPath) <= 0 {
		log.Fatal("Use -o to provide an output path")
	}

	// Open first input path
	var sub *astisub.Subtitles
	var err error
	if sub, err = astisub.Open(astisub.Options{Filename: (*inputPath.Slice)[0], Teletext: astisub.TeletextOptions{Page: *teletextPage}}); err != nil {
		log.Fatalf("%s while opening %s", err, (*inputPath.Slice)[0])
	}

	// Switch on subcommand
	switch cmd {
	case "convert":
		// Write
		if err = sub.Write(*outputPath); err != nil {
			log.Fatalf("%s while writing to %s", err, *outputPath)
		}
	case "fragment":
		// Validate fragment duration
		if *fragmentDuration <= 0 {
			log.Fatal("Use -f to provide a fragment duration")
		}

		// Fragment
		sub.Fragment(*fragmentDuration)

		// Write
		if err = sub.Write(*outputPath); err != nil {
			log.Fatalf("%s while writing to %s", err, *outputPath)
		}
	case "merge":
		// Validate second input path
		if len(*inputPath.Slice) == 1 {
			log.Fatal("Use -i to provide at least two input paths")
		}

		// Open second input path
		var sub2 *astisub.Subtitles
		if sub2, err = astisub.Open(astisub.Options{Filename: (*inputPath.Slice)[1], Teletext: astisub.TeletextOptions{Page: *teletextPage}}); err != nil {
			log.Fatalf("%s while opening %s", err, (*inputPath.Slice)[1])
		}

		// Merge
		sub.Merge(sub2)

		// Write
		if err = sub.Write(*outputPath); err != nil {
			log.Fatalf("%s while writing to %s", err, *outputPath)
		}
	case "optimize":
		// Optimize
		sub.Optimize()

		// Write
		if err = sub.Write(*outputPath); err != nil {
			log.Fatalf("%s while writing to %s", err, *outputPath)
		}
	case "sync":
		// Validate sync duration
		if *syncDuration == 0 && *syncTime < 0 {
			log.Fatal("Use -s or -t to provide a sync duration or sync time")
		}

		// Add
		if *syncDuration > 0 {
			sub.Add(*syncDuration)
		} else if *syncTime > -1 && len(sub.Items) > 0 {
			sub.Add(*syncTime - sub.Items[0].StartAt)
		}

		// Write
		if err = sub.Write(*outputPath); err != nil {
			log.Fatalf("%s while writing to %s", err, *outputPath)
		}
	case "unfragment":
		// Unfragment
		sub.Unfragment()

		// Write
		if err = sub.Write(*outputPath); err != nil {
			log.Fatalf("%s while writing to %s", err, *outputPath)
		}
	default:
		log.Fatalf("Invalid subcommand %s", cmd)
	}
}
