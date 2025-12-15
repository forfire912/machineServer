package coverage

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Coverage represents code coverage data
type Coverage struct {
	Files []FileCoverage `json:"files"`
	Summary Summary       `json:"summary"`
}

// FileCoverage represents coverage for a single file
type FileCoverage struct {
	Path       string          `json:"path"`
	Lines      []LineCoverage  `json:"lines"`
	Functions  []FunctionCoverage `json:"functions,omitempty"`
	Branches   []BranchCoverage   `json:"branches,omitempty"`
}

// LineCoverage represents coverage for a single line
type LineCoverage struct {
	LineNumber int  `json:"line_number"`
	Covered    bool `json:"covered"`
	HitCount   int  `json:"hit_count"`
}

// FunctionCoverage represents coverage for a function
type FunctionCoverage struct {
	Name       string `json:"name"`
	StartLine  int    `json:"start_line"`
	EndLine    int    `json:"end_line"`
	Covered    bool   `json:"covered"`
	HitCount   int    `json:"hit_count"`
}

// BranchCoverage represents coverage for a branch
type BranchCoverage struct {
	LineNumber int  `json:"line_number"`
	BranchID   int  `json:"branch_id"`
	Taken      bool `json:"taken"`
	HitCount   int  `json:"hit_count"`
}

// Summary represents coverage summary statistics
type Summary struct {
	TotalLines    int     `json:"total_lines"`
	CoveredLines  int     `json:"covered_lines"`
	LineCoverage  float64 `json:"line_coverage"`
	TotalFuncs    int     `json:"total_functions,omitempty"`
	CoveredFuncs  int     `json:"covered_functions,omitempty"`
	FuncCoverage  float64 `json:"function_coverage,omitempty"`
	TotalBranches int     `json:"total_branches,omitempty"`
	CoveredBranches int   `json:"covered_branches,omitempty"`
	BranchCoverage float64 `json:"branch_coverage,omitempty"`
}

// Collector collects code coverage data
type Collector struct {
	coverage *Coverage
}

// NewCollector creates a new coverage collector
func NewCollector() *Collector {
	return &Collector{
		coverage: &Coverage{
			Files: make([]FileCoverage, 0),
		},
	}
}

// AddFile adds coverage data for a file
func (c *Collector) AddFile(filePath string, lines []LineCoverage) {
	c.coverage.Files = append(c.coverage.Files, FileCoverage{
		Path:  filePath,
		Lines: lines,
	})
}

// Calculate calculates coverage summary
func (c *Collector) Calculate() {
	totalLines := 0
	coveredLines := 0

	for _, file := range c.coverage.Files {
		totalLines += len(file.Lines)
		for _, line := range file.Lines {
			if line.Covered {
				coveredLines++
			}
		}
	}

	c.coverage.Summary = Summary{
		TotalLines:   totalLines,
		CoveredLines: coveredLines,
		LineCoverage: float64(coveredLines) / float64(totalLines) * 100,
	}
}

// SaveJSON saves coverage data to JSON file
func (c *Collector) SaveJSON(outputPath string) error {
	data, err := json.MarshalIndent(c.coverage, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal coverage data: %w", err)
	}

	return ioutil.WriteFile(outputPath, data, 0644)
}

// SaveLCOV saves coverage data in LCOV format
func (c *Collector) SaveLCOV(outputPath string) error {
	var lcov string

	for _, file := range c.coverage.Files {
		lcov += fmt.Sprintf("TN:\n")
		lcov += fmt.Sprintf("SF:%s\n", file.Path)

		for _, line := range file.Lines {
			lcov += fmt.Sprintf("DA:%d,%d\n", line.LineNumber, line.HitCount)
		}

		lcov += fmt.Sprintf("end_of_record\n")
	}

	return ioutil.WriteFile(outputPath, []byte(lcov), 0644)
}

// GenerateHTML generates an HTML coverage report
func (c *Collector) GenerateHTML(outputDir string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate index.html
	indexPath := filepath.Join(outputDir, "index.html")
	html := c.generateIndexHTML()

	if err := ioutil.WriteFile(indexPath, []byte(html), 0644); err != nil {
		return fmt.Errorf("failed to write index.html: %w", err)
	}

	// Generate file-specific HTML reports
	for _, file := range c.coverage.Files {
		fileHTML := c.generateFileHTML(file)
		fileName := filepath.Base(file.Path) + ".html"
		filePath := filepath.Join(outputDir, fileName)

		if err := ioutil.WriteFile(filePath, []byte(fileHTML), 0644); err != nil {
			return fmt.Errorf("failed to write file HTML: %w", err)
		}
	}

	return nil
}

func (c *Collector) generateIndexHTML() string {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>Coverage Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .summary { background: #f0f0f0; padding: 15px; margin-bottom: 20px; }
        table { border-collapse: collapse; width: 100%; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #4CAF50; color: white; }
        .high { background-color: #90EE90; }
        .medium { background-color: #FFFFE0; }
        .low { background-color: #FFB6C1; }
    </style>
</head>
<body>
    <h1>Code Coverage Report</h1>
    <div class="summary">
        <h2>Summary</h2>
        <p>Line Coverage: %.2f%% (%d/%d)</p>
    </div>
    <table>
        <tr>
            <th>File</th>
            <th>Lines</th>
            <th>Coverage</th>
        </tr>
`
	html = fmt.Sprintf(html, c.coverage.Summary.LineCoverage, 
		c.coverage.Summary.CoveredLines, c.coverage.Summary.TotalLines)

	for _, file := range c.coverage.Files {
		covered := 0
		for _, line := range file.Lines {
			if line.Covered {
				covered++
			}
		}
		coverage := float64(covered) / float64(len(file.Lines)) * 100
		
		class := "low"
		if coverage >= 80 {
			class = "high"
		} else if coverage >= 60 {
			class = "medium"
		}

		html += fmt.Sprintf(`        <tr class="%s">
            <td><a href="%s.html">%s</a></td>
            <td>%d</td>
            <td>%.2f%%</td>
        </tr>
`, class, filepath.Base(file.Path), file.Path, len(file.Lines), coverage)
	}

	html += `    </table>
</body>
</html>`

	return html
}

func (c *Collector) generateFileHTML(file FileCoverage) string {
	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>Coverage: %s</title>
    <style>
        body { font-family: monospace; margin: 20px; }
        .line { padding: 2px 5px; }
        .covered { background-color: #90EE90; }
        .not-covered { background-color: #FFB6C1; }
        .line-number { color: #999; margin-right: 10px; }
    </style>
</head>
<body>
    <h1>%s</h1>
    <div>
`, file.Path, file.Path)

	for _, line := range file.Lines {
		class := "not-covered"
		if line.Covered {
			class = "covered"
		}
		html += fmt.Sprintf(`        <div class="line %s">
            <span class="line-number">%d</span>
            <span>Hits: %d</span>
        </div>
`, class, line.LineNumber, line.HitCount)
	}

	html += `    </div>
</body>
</html>`

	return html
}
