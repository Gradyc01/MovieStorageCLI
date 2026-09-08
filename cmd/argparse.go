package cmd

import "strings"

// multiWordFlags lists flags whose values are free-form text that
// commonly contains spaces (a movie title, a franchise name, a list
// of directors) — flags where forcing the user to remember to wrap
// the value in quotes is more friction than it's worth. Flags NOT in
// this list (like --updateScore, a number) don't need this treatment.
var multiWordFlags = map[string]bool{
	"--updateTitle":     true,
	"--updateFranchise": true,
	"--updateDirectors": true,
}

// joinMultiWordFlagValues rewrites a token slice so that, for any flag
// listed in multiWordFlags, every following token that doesn't itself
// look like a flag (i.e. doesn't start with "-") gets glued back
// together into a single space-separated value. This runs on the
// ALREADY-tokenized argument list — it works the same way whether
// those tokens came from splitArgs() in the REPL or straight from
// os.Args in one-shot mode, since by the time either reaches here
// they're just a []string.
//
// Quoted input still works exactly as before: a quoted phrase like
// "Dune Saga" is already a single token by the time it gets here, so
// this function just finds one token to "join" (with nothing to
// join it to) and leaves it unchanged.
func joinMultiWordFlagValues(args []string) []string {
	result := make([]string, 0, len(args))

	i := 0
	for i < len(args) {
		arg := args[i]

		if !multiWordFlags[arg] {
			result = append(result, arg)
			i++
			continue
		}

		// Collect every following token up to (but not including) the
		// next one that looks like a flag, or the end of the args.
		j := i + 1
		var valueParts []string
		for j < len(args) && !strings.HasPrefix(args[j], "-") {
			valueParts = append(valueParts, args[j])
			j++
		}

		result = append(result, arg)
		if len(valueParts) > 0 {
			result = append(result, strings.Join(valueParts, " "))
		}
		// If valueParts is empty, we deliberately do NOT invent an
		// empty string value here — that would silently succeed and
		// clear the field. Leaving the flag with nothing after it
		// lets cobra produce its normal "flag needs an argument"
		// error instead, which is the more honest failure mode.

		i = j
	}

	return result
}
