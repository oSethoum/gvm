package version

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Normalize strips a leading "go" prefix and whitespace.
func Normalize(v string) string {
	v = strings.TrimSpace(v)
	v = strings.TrimPrefix(v, "go")
	v = strings.TrimPrefix(v, "GO")
	return v
}

// GoPrefix returns "go1.27.4".
func GoPrefix(v string) string {
	v = Normalize(v)
	if v == "" {
		return ""
	}
	return "go" + v
}

var versionRe = regexp.MustCompile(`^(\d+)(?:\.(\d+))?(?:\.(\d+))?(?:-(.+))?$`)

// Parts holds parsed version components.
type Parts struct {
	Major int
	Minor int
	Patch int
	Pre   string // rc1, beta, etc.
	Raw   string
}

func Parse(v string) (Parts, error) {
	v = Normalize(v)
	m := versionRe.FindStringSubmatch(v)
	if m == nil {
		return Parts{}, fmt.Errorf("invalid version %q", v)
	}
	p := Parts{Raw: v}
	if m[1] != "" {
		p.Major, _ = strconv.Atoi(m[1])
	}
	if m[2] != "" {
		p.Minor, _ = strconv.Atoi(m[2])
	}
	if m[3] != "" {
		p.Patch, _ = strconv.Atoi(m[3])
	}
	p.Pre = m[4]
	return p, nil
}

func (p Parts) String() string {
	s := fmt.Sprintf("%d.%d.%d", p.Major, p.Minor, p.Patch)
	if p.Pre != "" {
		s += "-" + p.Pre
	}
	return s
}

func (p Parts) Less(o Parts) bool {
	if p.Major != o.Major {
		return p.Major < o.Major
	}
	if p.Minor != o.Minor {
		return p.Minor < o.Minor
	}
	if p.Patch != o.Patch {
		return p.Patch < o.Patch
	}
	// pre-release sorts before release
	if p.Pre == "" && o.Pre != "" {
		return false
	}
	if p.Pre != "" && o.Pre == "" {
		return true
	}
	return p.Pre < o.Pre
}

func IsPreRelease(v string) bool {
	p, err := Parse(v)
	if err != nil {
		return false
	}
	return p.Pre != ""
}

// Compare returns -1/0/1.
func Compare(a, b string) int {
	pa, err1 := Parse(a)
	pb, err2 := Parse(b)
	if err1 != nil || err2 != nil {
		return strings.Compare(Normalize(a), Normalize(b))
	}
	if pa.Less(pb) {
		return -1
	}
	if pb.Less(pa) {
		return 1
	}
	return 0
}

// MatchLatest returns the highest version in candidates whose major.minor matches prefix (if prefix has fewer parts).
func MatchLatest(prefix string, candidates []string) (string, bool) {
	prefix = Normalize(prefix)
	pp, err := Parse(prefix)
	if err != nil {
		return "", false
	}
	var best string
	var bestP Parts
	have := false
	for _, c := range candidates {
		cp, err := Parse(c)
		if err != nil {
			continue
		}
		// major match always; minor if specified in prefix
		if pp.Major != 0 && cp.Major != pp.Major {
			continue
		}
		if strings.Count(prefix, ".") >= 1 && cp.Minor != pp.Minor {
			continue
		}
		if strings.Count(prefix, ".") >= 2 && cp.Patch != pp.Patch {
			continue
		}
		if IsPreRelease(c) {
			// skip pre-release unless prefix explicitly pre-release
			if !IsPreRelease(prefix) {
				continue
			}
		}
		if !have || bestP.Less(cp) {
			best = Normalize(c)
			bestP = cp
			have = true
		}
	}
	return best, have
}

// SortDesc sorts version strings descending.
func SortDesc(versions []string) {
	sort.Slice(versions, func(i, j int) bool {
		return Compare(versions[i], versions[j]) > 0
	})
}
