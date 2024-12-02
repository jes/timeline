package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
	"text/template"
)

type Event struct {
	StartYear int
	EndYear   int
	Text      string
	IsBC      bool
	HasAD     bool
}

type TemplateData struct {
	Events       []Event
	EarliestYear int
	LatestYear   int
	TotalYears   int
}

func parseYear(year string) (int, bool, bool) {
	isBC := strings.HasSuffix(year, "BC")
	hasAD := strings.HasSuffix(year, "AD")

	// Remove BC/AD suffix
	year = strings.TrimSuffix(year, "BC")
	year = strings.TrimSuffix(year, "AD")

	var y int
	fmt.Sscanf(year, "%d", &y)
	if isBC {
		y = -y
	}
	return y, isBC, hasAD
}

func parseTimeline(filename string) ([]Event, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var events []Event
	scanner := bufio.NewScanner(file)
	re := regexp.MustCompile(`^(\d+(?:\s*[AB]C)?(?:-\d+(?:\s*[AB]C)?)?)\s*:\s*(.+)$`)

	for scanner.Scan() {
		line := scanner.Text()
		matches := re.FindStringSubmatch(line)
		if matches == nil {
			continue
		}

		dateStr := matches[1]
		text := matches[2]

		var event Event
		event.Text = text

		// Handle date ranges
		if strings.Contains(dateStr, "-") {
			parts := strings.Split(dateStr, "-")
			start, isBC, hasAD := parseYear(strings.TrimSpace(parts[0]))
			end, _, _ := parseYear(strings.TrimSpace(parts[1]))
			event.StartYear = start
			event.EndYear = end
			event.IsBC = isBC
			event.HasAD = hasAD
		} else {
			year, isBC, hasAD := parseYear(strings.TrimSpace(dateStr))
			event.StartYear = year
			event.EndYear = year
			event.IsBC = isBC
			event.HasAD = hasAD
		}

		events = append(events, event)
	}

	return events, scanner.Err()
}

func findEarliestYear(events []Event) int {
	if len(events) == 0 {
		return 0
	}
	earliest := events[0].StartYear
	for _, e := range events {
		if e.StartYear < earliest {
			earliest = e.StartYear
		}
	}
	return earliest
}

func main() {
	events, err := parseTimeline("timeline.mw")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading timeline: %v\n", err)
		os.Exit(1)
	}

	earliestYear := findEarliestYear(events)
	latestYear := events[len(events)-1].EndYear
	totalYears := latestYear - earliestYear

	data := TemplateData{
		Events:       events,
		EarliestYear: earliestYear,
		LatestYear:   latestYear,
		TotalYears:   totalYears,
	}

	funcMap := template.FuncMap{
		"subtract": func(a, b int) int { return a - b },
		"abs": func(a int) int {
			if a < 0 {
				return -a
			}
			return a
		},
		"multiply": func(a, b int) int { return a * b },
	}

	tmpl, err := template.New("timeline").Funcs(funcMap).ParseFiles("timeline.html.tmpl")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing template: %v\n", err)
		os.Exit(1)
	}

	err = tmpl.ExecuteTemplate(os.Stdout, "timeline.html.tmpl", data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing template: %v\n", err)
		os.Exit(1)
	}
}
