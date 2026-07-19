package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

type WeekKey string

type WeekStats struct {
	Add     int
	Del     int
	Commits int
}

type AuthorWeeks map[string]map[WeekKey]*WeekStats

func aggregateWeeks(weeks AuthorWeeks) map[WeekKey]*WeekStats {
	combo := map[WeekKey]*WeekStats{}
	for _, authorWeeks := range weeks {
		for week, stats := range authorWeeks {
			if _, ok := combo[week]; !ok {
				combo[week] = &WeekStats{}
			}
			combo[week].Add += stats.Add
			combo[week].Del += stats.Del
			combo[week].Commits += stats.Commits
		}
	}
	return combo
}

// ------------------------------------
// CLI flags
// ------------------------------------

var (
	flagSince  = flag.String("since", "", "Only include commits on or after this date (YYYY-MM-DD)")
	flagUntil  = flag.String("until", "", "Only include commits on or before this date (YYYY-MM-DD)")
	flagAuthor = flag.String("author", "", "Filter authors by regex")
	flagLimit  = flag.Int("limit", 0, "Limit the number of contributors shown (by impact)")
	flagDir    = flag.String("dir", ".", "Git directory to run inside (optional positional argument overrides)")
	flagShort  = flag.Bool("short", false, "Show only the aggregate All Contributors card")
)

// ------------------------------------
// Date utilities
// ------------------------------------

func parseDate(s string) *time.Time {
	if s == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Invalid date:", s)
		os.Exit(1)
	}
	return &t
}

func dateInRange(t time.Time, since, until *time.Time) bool {
	if since != nil && t.Before(*since) {
		return false
	}
	if until != nil && t.After(*until) {
		return false
	}
	return true
}

// ------------------------------------
// Git parsing
// ------------------------------------

func parseGitLog(repoDir string, since, until *time.Time, authorFilter *regexp.Regexp) (AuthorWeeks, error) {
	if _, err := os.Stat(repoDir); err != nil {
		return nil, fmt.Errorf("invalid --dir %q: %w", repoDir, err)
	}
	cmd := exec.Command("git", "log", "--numstat", "--date=short", "--pretty=%aN|%ad")
	cmd.Dir = repoDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			msg = err.Error()
		}
		return nil, fmt.Errorf("git log failed in %s: %s: %w", repoDir, msg, err)
	}

	lines := bytes.Split(out, []byte("\n"))

	weeks := AuthorWeeks{}
	var author string
	var week WeekKey
	var includeCommit bool

	for _, lnBytes := range lines {
		ln := string(lnBytes)

		// line with author|date
		if strings.Contains(ln, "|") {
			parts := strings.Split(ln, "|")
			author = strings.TrimSpace(parts[0])
			date := strings.TrimSpace(parts[1])

			t, terr := time.Parse("2006-01-02", date)
			if terr != nil {
				includeCommit = false
				continue
			}

			includeCommit = dateInRange(t, since, until)
			if includeCommit && authorFilter != nil {
				includeCommit = authorFilter.MatchString(author)
			}

			if includeCommit {
				year, weekNum := t.ISOWeek()
				week = WeekKey(fmt.Sprintf("%d-W%02d", year, weekNum))
				if _, ok := weeks[author]; !ok {
					weeks[author] = map[WeekKey]*WeekStats{}
				}
				if _, ok := weeks[author][week]; !ok {
					weeks[author][week] = &WeekStats{}
				}
				weeks[author][week].Commits++
			}
			continue
		}

		// numstat line
		if !includeCommit {
			continue
		}

		fields := strings.Fields(ln)
		if len(fields) == 3 {
			add := parseInt(fields[0])
			del := parseInt(fields[1])
			weeks[author][week].Add += add
			weeks[author][week].Del += del
		}
	}

	return weeks, nil
}

func parseInt(s string) int {
	if s == "-" {
		return 0
	}
	var n int
	if _, err := fmt.Sscanf(s, "%d", &n); err != nil {
		return 0
	}
	return n
}

// ------------------------------------
// Terminal width helper
// ------------------------------------

func terminalWidth() int {
	cmd := exec.Command("tput", "cols")
	out, err := cmd.Output()
	if err != nil {
		return 80
	}
	var n int
	fmt.Sscanf(string(out), "%d", &n)
	if n < 40 {
		return 80
	}
	return n
}

// ------------------------------------
// Styles
// ------------------------------------

var (
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00D7D7"))
	border     = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1)

	addColor = lipgloss.NewStyle().Foreground(lipgloss.Color("#5FFF87"))
	delColor = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5F87"))
	comColor = lipgloss.NewStyle().Foreground(lipgloss.Color("#5FAFFF"))
)

// ------------------------------------
// Stacked bar rendering
// ------------------------------------

func barSegment(val, max, width int, style lipgloss.Style) string {
	if max == 0 {
		return ""
	}
	n := int(float64(val) / float64(max) * float64(width))
	if n < 1 && val > 0 {
		n = 1
	}
	return style.Render(strings.Repeat("█", n))
}

func renderStackedWeeklyChart(stats map[WeekKey]*WeekStats) (string, []int) {
	// ordered weeks
	var weeks []string
	for w := range stats {
		weeks = append(weeks, string(w))
	}
	sort.Strings(weeks)

	// compute max impact
	max := 0
	var impacts []int
	for _, w := range weeks {
		ws := stats[WeekKey(w)]
		total := ws.Add + ws.Del + ws.Commits
		impacts = append(impacts, total)
		if total > max {
			max = total
		}
	}

	width := terminalWidth() - 22
	if width < 30 {
		width = 30
	}

	var out strings.Builder

	for _, w := range weeks {
		ws := stats[WeekKey(w)]
		addSeg := barSegment(ws.Add, max, width, addColor)
		delSeg := barSegment(ws.Del, max, width, delColor)
		comSeg := barSegment(ws.Commits, max, width, comColor)

		total := ws.Add + ws.Del + ws.Commits

		out.WriteString(fmt.Sprintf(
			"%s | %s%s%s (%d)\n",
			w,
			addSeg,
			delSeg,
			comSeg,
			total,
		))
	}

	return out.String(), impacts
}

// ------------------------------------
// Sparkline
// ------------------------------------

var sparkChars = []rune("▁▂▃▄▅▆▇█")

func sparkline(values []int) string {
	if len(values) == 0 {
		return ""
	}
	max := 0
	for _, v := range values {
		if v > max {
			max = v
		}
	}
	if max == 0 {
		return strings.Repeat("▁", len(values))
	}

	var out strings.Builder
	for _, v := range values {
		idx := int(float64(v) / float64(max) * float64(len(sparkChars)-1))
		out.WriteRune(sparkChars[idx])
	}
	return out.String()
}

// ------------------------------------
// Author card
// ------------------------------------

func renderAuthorCard(author string, stats map[WeekKey]*WeekStats) (string, int) {
	totalCommits := 0
	totalAdd := 0
	totalDel := 0
	for _, v := range stats {
		totalCommits += v.Commits
		totalAdd += v.Add
		totalDel += v.Del
	}

	impact := totalAdd + totalDel + totalCommits

	header := fmt.Sprintf(
		"%s\nCommits: %d   +%d / -%d   Impact: %d",
		titleStyle.Render(author),
		totalCommits,
		totalAdd,
		totalDel,
		impact,
	)

	weeklyChart, impacts := renderStackedWeeklyChart(stats)
	spark := sparkline(impacts)
	summary := fmt.Sprintf("Summary: %s", spark)

	return border.Render(
		fmt.Sprintf("%s\n\n%s\n%s", header, weeklyChart, summary),
	), impact
}

func renderTeamOverview(stats map[WeekKey]*WeekStats) string {
	if len(stats) == 0 {
		return border.Render("No matching commits found")
	}

	totalCommits := 0
	totalAdd := 0
	totalDel := 0
	for _, v := range stats {
		totalCommits += v.Commits
		totalAdd += v.Add
		totalDel += v.Del
	}

	headline := fmt.Sprintf(
		"%s\nCommits: %d   +%d / -%d",
		titleStyle.Render("All Contributors"),
		totalCommits,
		totalAdd,
		totalDel,
	)

	weeklyChart, _ := renderStackedWeeklyChart(stats)

	return border.Render(fmt.Sprintf("%s\n\n%s", headline, weeklyChart))
}

// ------------------------------------
// Main
// ------------------------------------

func main() {
	flag.Parse()

	dir := *flagDir
	args := flag.Args()
	if len(args) > 1 {
		fmt.Fprintln(os.Stderr, "Error: too many positional arguments (expected at most a repo path)")
		os.Exit(1)
	}
	if len(args) == 1 {
		if *flagDir != "." && *flagDir != args[0] {
			fmt.Fprintln(os.Stderr, "Error: repo directory specified via -dir and positional argument; please choose one")
			os.Exit(1)
		}
		dir = args[0]
	}

	var rx *regexp.Regexp
	if *flagAuthor != "" {
		var err error
		rx, err = regexp.Compile(*flagAuthor)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid --author regex: %v\n", err)
			os.Exit(1)
		}
	}

	weeks, err := parseGitLog(
		dir,
		parseDate(*flagSince),
		parseDate(*flagUntil),
		rx,
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}

	teamStats := aggregateWeeks(weeks)
	teamCard := renderTeamOverview(teamStats)

	type row struct {
		author string
		card   string
		impact int
	}
	var rows []row

	for author, stats := range weeks {
		card, impact := renderAuthorCard(author, stats)
		rows = append(rows, row{author, card, impact})
	}

	// sort by impact descending
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].impact > rows[j].impact
	})

	// limit
	if *flagLimit > 0 && *flagLimit < len(rows) {
		rows = rows[:*flagLimit]
	}

	fmt.Println(teamCard)
	fmt.Println()
	if *flagShort {
		return
	}

	for _, r := range rows {
		fmt.Println(r.card)
		fmt.Println()
	}
}
