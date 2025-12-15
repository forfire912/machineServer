package coverage

import (
	"encoding/json"
	"fmt"
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
	totalFuncs := 0
	coveredFuncs := 0
	totalBranches := 0
	coveredBranches := 0

	for _, file := range c.coverage.Files {
		// Lines
		totalLines += len(file.Lines)
		for _, line := range file.Lines {
			if line.Covered {
				coveredLines++
			}
		}

		// Functions
		totalFuncs += len(file.Functions)
		for _, fn := range file.Functions {
			if fn.Covered {
				coveredFuncs++
			}
		}

		// Branches
		totalBranches += len(file.Branches)
		for _, br := range file.Branches {
			if br.Taken {
				coveredBranches++
			}
		}
	}

	lineCoverage := 0.0
	if totalLines > 0 {
		lineCoverage = float64(coveredLines) / float64(totalLines) * 100
	}

	funcCoverage := 0.0
	if totalFuncs > 0 {
		funcCoverage = float64(coveredFuncs) / float64(totalFuncs) * 100
	}

	branchCoverage := 0.0
	if totalBranches > 0 {
		branchCoverage = float64(coveredBranches) / float64(totalBranches) * 100
	}

	c.coverage.Summary = Summary{
		TotalLines:      totalLines,
		CoveredLines:    coveredLines,
		LineCoverage:    lineCoverage,
		TotalFuncs:      totalFuncs,
		CoveredFuncs:    coveredFuncs,
		FuncCoverage:    funcCoverage,
		TotalBranches:   totalBranches,
		CoveredBranches: coveredBranches,
		BranchCoverage:  branchCoverage,
	}
}

// SaveJSON saves coverage data to JSON file
func (c *Collector) SaveJSON(outputPath string) error {
	data, err := json.MarshalIndent(c.coverage, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal coverage data: %w", err)
	}

	return os.WriteFile(outputPath, data, 0644)
}

// SaveLCOV saves coverage data in LCOV format
func (c *Collector) SaveLCOV(outputPath string) error {
	var lcov string

	for _, file := range c.coverage.Files {
		lcov += fmt.Sprintf("TN:\n")
		lcov += fmt.Sprintf("SF:%s\n", file.Path)

		// Functions
		for _, fn := range file.Functions {
			lcov += fmt.Sprintf("FN:%d,%s\n", fn.StartLine, fn.Name)
		}
		for _, fn := range file.Functions {
			lcov += fmt.Sprintf("FNDA:%d,%s\n", fn.HitCount, fn.Name)
		}
		if len(file.Functions) > 0 {
			coveredFuncs := 0
			for _, fn := range file.Functions {
				if fn.Covered {
					coveredFuncs++
				}
			}
			lcov += fmt.Sprintf("FNF:%d\n", len(file.Functions))
			lcov += fmt.Sprintf("FNH:%d\n", coveredFuncs)
		}

		// Branches
		for _, br := range file.Branches {
			taken := "-"
			if br.Taken {
				taken = fmt.Sprintf("%d", br.HitCount)
			}
			lcov += fmt.Sprintf("BRDA:%d,%d,%d,%s\n", br.LineNumber, 0, br.BranchID, taken)
		}
		if len(file.Branches) > 0 {
			coveredBranches := 0
			for _, br := range file.Branches {
				if br.Taken {
					coveredBranches++
				}
			}
			lcov += fmt.Sprintf("BRF:%d\n", len(file.Branches))
			lcov += fmt.Sprintf("BRH:%d\n", coveredBranches)
		}

		// Lines
		for _, line := range file.Lines {
			lcov += fmt.Sprintf("DA:%d,%d\n", line.LineNumber, line.HitCount)
		}
		if len(file.Lines) > 0 {
			coveredLines := 0
			for _, line := range file.Lines {
				if line.Covered {
					coveredLines++
				}
			}
			lcov += fmt.Sprintf("LF:%d\n", len(file.Lines))
			lcov += fmt.Sprintf("LH:%d\n", coveredLines)
		}

		lcov += fmt.Sprintf("end_of_record\n")
	}

	return os.WriteFile(outputPath, []byte(lcov), 0644)
}

// GenerateHTML generates an HTML coverage report
func (c *Collector) GenerateHTML(outputDir string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate index.html
	indexPath := filepath.Join(outputDir, "index.html")
	html := c.generateIndexHTML()

	if err := os.WriteFile(indexPath, []byte(html), 0644); err != nil {
		return fmt.Errorf("failed to write index.html: %w", err)
	}

	// Generate file-specific HTML reports
	for _, file := range c.coverage.Files {
		fileHTML := c.generateFileHTML(file)
		fileName := filepath.Base(file.Path) + ".html"
		filePath := filepath.Join(outputDir, fileName)

		if err := os.WriteFile(filePath, []byte(fileHTML), 0644); err != nil {
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
        <p>Function Coverage: %.2f%% (%d/%d)</p>
        <p>Branch Coverage: %.2f%% (%d/%d)</p>
    </div>
    <table>
        <tr>
            <th>File</th>
            <th>Lines</th>
            <th>Functions</th>
            <th>Branches</th>
            <th>Coverage</th>
        </tr>
`
	html = fmt.Sprintf(html, 
		c.coverage.Summary.LineCoverage, c.coverage.Summary.CoveredLines, c.coverage.Summary.TotalLines,
		c.coverage.Summary.FuncCoverage, c.coverage.Summary.CoveredFuncs, c.coverage.Summary.TotalFuncs,
		c.coverage.Summary.BranchCoverage, c.coverage.Summary.CoveredBranches, c.coverage.Summary.TotalBranches)

	for _, file := range c.coverage.Files {
		coveredLines := 0
		for _, line := range file.Lines {
			if line.Covered {
				coveredLines++
			}
		}
		lineCov := 0.0
		if len(file.Lines) > 0 {
			lineCov = float64(coveredLines) / float64(len(file.Lines)) * 100
		}

		coveredFuncs := 0
		for _, fn := range file.Functions {
			if fn.Covered {
				coveredFuncs++
			}
		}
		funcCov := 0.0
		if len(file.Functions) > 0 {
			funcCov = float64(coveredFuncs) / float64(len(file.Functions)) * 100
		}

		coveredBranches := 0
		for _, br := range file.Branches {
			if br.Taken {
				coveredBranches++
			}
		}
		branchCov := 0.0
		if len(file.Branches) > 0 {
			branchCov = float64(coveredBranches) / float64(len(file.Branches)) * 100
		}
		
		class := "low"
		if lineCov >= 80 {
			class = "high"
		} else if lineCov >= 60 {
			class = "medium"
		}

		html += fmt.Sprintf(`        <tr class="%s">
            <td><a href="%s.html">%s</a></td>
            <td>%.2f%% (%d/%d)</td>
            <td>%.2f%% (%d/%d)</td>
            <td>%.2f%% (%d/%d)</td>
            <td>%.2f%%</td>
        </tr>
`, class, filepath.Base(file.Path), file.Path, 
			lineCov, coveredLines, len(file.Lines),
			funcCov, coveredFuncs, len(file.Functions),
			branchCov, coveredBranches, len(file.Branches),
			lineCov)
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
        .function { background-color: #E6E6FA; margin: 5px 0; padding: 5px; }
        .branch { background-color: #FFFACD; margin: 2px 0; padding: 2px 20px; font-size: 0.9em; }
    </style>
</head>
<body>
    <h1>%s</h1>
    
    <div class="functions">
        <h2>Functions</h2>
        <ul>
`, file.Path, file.Path)

	for _, fn := range file.Functions {
		status := "Not Covered"
		if fn.Covered {
			status = fmt.Sprintf("Covered (Hits: %d)", fn.HitCount)
		}
		html += fmt.Sprintf(`            <li class="function">%s (Lines %d-%d): %s</li>
`, fn.Name, fn.StartLine, fn.EndLine, status)
	}

	html += `        </ul>
    </div>

    <div class="code">
        <h2>Source</h2>
`

	// Create a map for quick lookup
	lineMap := make(map[int]LineCoverage)
	for _, l := range file.Lines {
		lineMap[l.LineNumber] = l
	}
	
	branchMap := make(map[int][]BranchCoverage)
	for _, b := range file.Branches {
		branchMap[b.LineNumber] = append(branchMap[b.LineNumber], b)
	}

	// Find max line number
	maxLine := 0
	for _, l := range file.Lines {
		if l.LineNumber > maxLine {
			maxLine = l.LineNumber
		}
	}

	for i := 1; i <= maxLine; i++ {
		line, exists := lineMap[i]
		class := ""
		hits := ""
		
		if exists {
			if line.Covered {
				class = "covered"
				hits = fmt.Sprintf("Hits: %d", line.HitCount)
			} else {
				class = "not-covered"
				hits = "Hits: 0"
			}
		}

		html += fmt.Sprintf(`        <div class="line %s">
            <span class="line-number">%d</span>
            <span>%s</span>
        </div>
`, class, i, hits)

		// Add branches if any
		if branches, ok := branchMap[i]; ok {
			for _, br := range branches {
				taken := "Not Taken"
				if br.Taken {
					taken = fmt.Sprintf("Taken (Hits: %d)", br.HitCount)
				}
				html += fmt.Sprintf(`        <div class="branch">Branch %d: %s</div>
`, br.BranchID, taken)
			}
		}
	}

	html += `    </div>
</body>
</html>`

	return html
}
