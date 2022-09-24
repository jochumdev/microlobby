package utils

import "github.com/urfave/cli/v2"

func MergeFlag(flags []cli.Flag, flag cli.Flag) []cli.Flag {
	has := false
	for _, f := range flags {
		if f.Names()[0] == flag.Names()[0] {
			has = true
			break
		}
	}

	if has {
		return flags
	}

	return append(flags, flag)
}

func MergeFlags(flags []cli.Flag, newFlags ...cli.Flag) []cli.Flag {
	for _, f := range newFlags {
		flags = MergeFlag(flags, f)
	}

	return flags
}
