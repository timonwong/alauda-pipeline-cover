package coverreport

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mattn/go-zglob"
	"golang.org/x/tools/cover"
)

const (
	OrderAsc  = "asc"
	OrderDesc = "desc"
)

const (
	SortByFilename      = "filename"
	SortByPackage       = "package"
	SortByBlock         = "block"
	SortByStmt          = "stmt"
	SortByMissingBlocks = "missing-blocks"
	SortByMissingStmts  = "missing-stmts"
	SortByBlockCoverage = "block-coverage"
	SortByStmtCoverage  = "stmt-coverage"
)

// Configuration structure
type Configuration struct {
	Root       string
	Exclusions []string
	SortBy     string
	Order      string
}

// Summary is coverage summary for a file or module
type Summary struct {
	Name                                       string
	Blocks, Stmts, MissingBlocks, MissingStmts int64
	BlockCoverage, StmtCoverage                float64
}

// Report of the coverage results
type Report struct {
	Total Summary   // Global coverage
	Files []Summary // Coverage by file
}

// GenerateReport generates a coverage report given the coverage profile file, and the following configurations:
// exclusions: packages to be excluded (if a package is excluded, all its subpackages are excluded as well)
// sortBy: the order in which the files will be sorted in the report (see sortResults)
// order: the direction of the the sorting
func GenerateReport(coverprofile string, conf *Configuration, packages bool) (*Report, error) {
	profiles, err := cover.ParseProfiles(coverprofile)
	if err != nil {
		return nil, fmt.Errorf("invalid coverprofile: %w", err)
	}
	total := &accumulator{name: "Total"}
	files := make(map[string]*accumulator)
	for _, profile := range profiles {
		fileName := normalizeName(profile.FileName, conf.Root, packages)
		if isExcluded(fileName, conf.Exclusions) {
			continue
		}
		fileCover, ok := files[fileName]
		if !ok {
			// Create new accumulator
			fileCover = &accumulator{name: fileName}
			files[fileName] = fileCover
		}
		total.addAll(profile.Blocks)
		fileCover.addAll(profile.Blocks)
	}
	return makeReport(total, files, conf.SortBy, conf.Order)
}

// Removes root dir part if configured to do so
func normalizeName(filename, root string, packages bool) string {
	if packages {
		filename = filepath.Dir(filename)
	}

	if root == "" {
		return filename
	}
	if packages {
		return "." + strings.TrimPrefix(filename, root)
	}
	return strings.TrimPrefix(filename, root)
}

func isExcluded(fileName string, exclusions []string) bool {
	for _, exclusion := range exclusions {
		if ok, _ := zglob.Match(exclusion, fileName); ok {
			return true
		}
	}
	return false
}

// Creates a Report struct from the coverage summarization results
func makeReport(total *accumulator, files map[string]*accumulator, sortBy, order string) (*Report, error) {
	fileReports := make([]Summary, 0, len(files))
	for _, fileCover := range files {
		fileReports = append(fileReports, fileCover.results())
	}
	if err := sortResults(fileReports, sortBy, order); err != nil {
		return nil, err
	}
	return &Report{
		Total: total.results(),
		Files: fileReports,
	}, nil
}

// Accumulates the coverage of a file and returns a summary
type accumulator struct {
	name                                       string
	blocks, stmts, coveredBlocks, coveredStmts int64
}

// Accumulates a profile block
func (a *accumulator) add(block cover.ProfileBlock) {
	a.blocks++
	a.stmts += int64(block.NumStmt)
	if block.Count > 0 {
		a.coveredBlocks++
		a.coveredStmts += int64(block.NumStmt)
	}
}

func (a *accumulator) addAll(blocks []cover.ProfileBlock) {
	for _, block := range blocks {
		a.add(block)
	}
}

// Creates a summary with the accumulated values
func (a *accumulator) results() Summary {
	return Summary{
		Name:          a.name,
		Blocks:        a.blocks,
		Stmts:         a.stmts,
		MissingBlocks: a.blocks - a.coveredBlocks,
		MissingStmts:  a.stmts - a.coveredStmts,
		BlockCoverage: float64(a.coveredBlocks) / float64(a.blocks) * 100,
		StmtCoverage:  float64(a.coveredStmts) / float64(a.stmts) * 100,
	}
}

// Sorts the individual coverage reports by a given column
// (block --block coverage--, stmt --stmt coverage--, missing-blocks or missing-stmts)
// and a sorting direction (asc or desc)
func sortResults(reports []Summary, sortBy, order string) error {
	var reverse bool
	var cmp func(i, j int) bool

	switch order {
	case OrderAsc:
		reverse = false
	case OrderDesc:
		reverse = true
	default:
		return fmt.Errorf("order must be either asc or desc, got %q", order)
	}
	switch sortBy {
	case SortByFilename, SortByPackage:
		cmp = func(i, j int) bool {
			return reports[i].Name < reports[j].Name
		}
	case SortByBlock:
		cmp = func(i, j int) bool {
			return reports[i].BlockCoverage < reports[j].BlockCoverage
		}
	case SortByStmt:
		cmp = func(i, j int) bool {
			return reports[i].StmtCoverage < reports[j].StmtCoverage
		}
	case SortByMissingBlocks:
		cmp = func(i, j int) bool {
			return reports[i].MissingBlocks < reports[j].MissingBlocks
		}
	case SortByMissingStmts:
		cmp = func(i, j int) bool {
			return reports[i].MissingStmts < reports[j].MissingStmts
		}
	case SortByBlockCoverage:
		cmp = func(i, j int) bool {
			return reports[i].BlockCoverage < reports[j].BlockCoverage
		}
	case SortByStmtCoverage:
		cmp = func(i, j int) bool {
			return reports[i].StmtCoverage < reports[j].StmtCoverage
		}
	default:
		return fmt.Errorf("invalid sort column %q", sortBy)
	}
	sort.SliceStable(reports, func(i, j int) bool {
		if reverse {
			return !cmp(i, j)
		}
		return cmp(i, j)
	})
	return nil
}
