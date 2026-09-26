package find

import (
	"strconv"
	"strings"
	"time"
)

// Predicates encapsulates the search criteria for find.
type Predicates struct {
	NamePattern      string // -name
	INamePattern     string // -iname
	PathPattern      string // -path
	EntityType       string // -type: f, d, l, s, p
	MinSize          int64  // -size +N
	MaxSize          int64  // -size -N
	ExactSize        int64  // -size N
	HasMinSize       bool
	HasMaxSize       bool
	HasExactSize     bool
	ModifiedAfter    time.Time // -mtime -Nd
	ModifiedBefore   time.Time // -mtime +Nd
	HasModAfter      bool
	HasModBefore     bool
	EmptyOnly        bool // -empty
	MaxDepth         int  // -maxdepth
	MinDepth         int  // -mindepth
	IncludeHidden    bool // --hidden
	Plain            bool // --plain
	JSON             bool // --json
	RootPath         string
}

// DefaultPredicates returns empty predicates matching everything.
func DefaultPredicates() Predicates {
	return Predicates{
		MaxDepth:      -1,
		MinDepth:      0,
		IncludeHidden: false,
		Plain:         false,
		JSON:          false,
		RootPath:      ".",
	}
}

// ParseFlags parses command arguments and predicates for find.
func ParseFlags(args []string) Predicates {
	p := DefaultPredicates()
	var paths []string

	i := 0
	for i < len(args) {
		arg := args[i]

		// Global flags
		if arg == "--plain" {
			p.Plain = true
			i++
			continue
		}
		if arg == "--json" {
			p.JSON = true
			i++
			continue
		}
		if strings.HasPrefix(arg, "--theme=") || strings.HasPrefix(arg, "--color=") || strings.HasPrefix(arg, "--config=") {
			i++
			continue
		}
		if arg == "--theme" || arg == "--color" || arg == "--config" {
			i += 2
			continue
		}

		if arg == "--hidden" || arg == "-hidden" {
			p.IncludeHidden = true
			i++
		} else if arg == "-empty" || arg == "--empty" {
			p.EmptyOnly = true
			i++
		} else if (arg == "-name" || arg == "--name") && i+1 < len(args) {
			p.NamePattern = args[i+1]
			i += 2
		} else if (arg == "-iname" || arg == "--iname") && i+1 < len(args) {
			p.INamePattern = strings.ToLower(args[i+1])
			i += 2
		} else if (arg == "-path" || arg == "--path") && i+1 < len(args) {
			p.PathPattern = args[i+1]
			i += 2
		} else if (arg == "-type" || arg == "--type") && i+1 < len(args) {
			p.EntityType = args[i+1]
			i += 2
		} else if (arg == "-maxdepth" || arg == "--maxdepth") && i+1 < len(args) {
			if val, err := strconv.Atoi(args[i+1]); err == nil {
				p.MaxDepth = val
			}
			i += 2
		} else if (arg == "-mindepth" || arg == "--mindepth") && i+1 < len(args) {
			if val, err := strconv.Atoi(args[i+1]); err == nil {
				p.MinDepth = val
			}
			i += 2
		} else if (arg == "-size" || arg == "--size") && i+1 < len(args) {
			parseSizePredicate(args[i+1], &p)
			i += 2
		} else if (arg == "-mtime" || arg == "--mtime") && i+1 < len(args) {
			parseMtimePredicate(args[i+1], &p)
			i += 2
		} else if !strings.HasPrefix(arg, "-") && len(paths) == 0 {
			paths = append(paths, arg)
			i++
		} else {
			i++
		}
	}

	if len(paths) > 0 {
		p.RootPath = paths[0]
	}
	return p
}

func parseSizePredicate(val string, p *Predicates) {
	if val == "" {
		return
	}
	prefix := val[0]
	numStr := val
	if prefix == '+' || prefix == '-' {
		numStr = val[1:]
	}

	mult := int64(1)
	lower := strings.ToLower(numStr)
	switch {
	case strings.HasSuffix(lower, "k"):
		mult = 1024
		numStr = numStr[:len(numStr)-1]
	case strings.HasSuffix(lower, "m"):
		mult = 1024 * 1024
		numStr = numStr[:len(numStr)-1]
	case strings.HasSuffix(lower, "g"):
		mult = 1024 * 1024 * 1024
		numStr = numStr[:len(numStr)-1]
	case strings.HasSuffix(lower, "b"):
		numStr = numStr[:len(numStr)-1]
	}

	bytes, err := strconv.ParseInt(numStr, 10, 64)
	if err != nil {
		return
	}
	total := bytes * mult

	if prefix == '+' {
		p.MinSize = total
		p.HasMinSize = true
	} else if prefix == '-' {
		p.MaxSize = total
		p.HasMaxSize = true
	} else {
		p.ExactSize = total
		p.HasExactSize = true
	}
}

func parseMtimePredicate(val string, p *Predicates) {
	if val == "" {
		return
	}
	prefix := val[0]
	numStr := val
	if prefix == '+' || prefix == '-' {
		numStr = val[1:]
	}

	durUnit := time.Hour * 24
	if strings.HasSuffix(numStr, "d") {
		numStr = numStr[:len(numStr)-1]
	} else if strings.HasSuffix(numStr, "h") {
		durUnit = time.Hour
		numStr = numStr[:len(numStr)-1]
	} else if strings.HasSuffix(numStr, "m") {
		durUnit = time.Minute
		numStr = numStr[:len(numStr)-1]
	}

	n, err := strconv.Atoi(numStr)
	if err != nil {
		return
	}
	dur := time.Duration(n) * durUnit
	now := time.Now()

	if prefix == '-' {
		// modified within the last N time: mod_time > now - dur
		p.ModifiedAfter = now.Add(-dur)
		p.HasModAfter = true
	} else if prefix == '+' {
		// modified more than N time ago: mod_time < now - dur
		p.ModifiedBefore = now.Add(-dur)
		p.HasModBefore = true
	}
}
