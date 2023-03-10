package main

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"sort"
	"time"
)

const daysInLastSixMonths = 183
const outOfRange = 99999
const weekInLastSixMonths = 26
const viewsColsNum = 28

type column []int

func stats(email string) {
	commits := getCommitsInfo(email)
	render(commits)
}

func getCommitsInfo(email string) map[int]int {
	filePath := getFilePath()
	repos := resolveFile(filePath)
	daysInMap := daysInLastSixMonths
	commits := make(map[int]int, daysInMap)
	for i := 0; i < daysInMap; i++ {
		commits[i] = 0
	}
	for _, repo := range repos {
		commits = fillCommits(commits, repo, email)
	}
	return commits
}

func fillCommits(commits map[int]int, path string, email string) map[int]int {
	repo, err := git.PlainOpen(path)
	if err != nil {
		panic(err)
	}
	ref, err := repo.Head()
	if err != nil {
		panic(err)
	}
	iterator, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		panic(err)
	}
	offset := calcOffset()
	err = iterator.ForEach(func(commit *object.Commit) error {
		daysAgo := countDaySinceDate(commit.Author.When) + offset
		if commit.Author.Email != email {
			return nil
		}
		if daysAgo != outOfRange {
			commits[daysAgo]++
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	return commits
}

func calcOffset() int {
	var offset int
	weekday := time.Now().Weekday()
	switch weekday {
	case time.Sunday:
		offset = 7
	case time.Monday:
		offset = 6
	case time.Tuesday:
		offset = 5
	case time.Wednesday:
		offset = 4
	case time.Thursday:
		offset = 3
	case time.Friday:
		offset = 2
	case time.Saturday:
		offset = 1
	}
	return offset
}

func countDaySinceDate(date time.Time) int {
	days := 0
	now := getBeginningOfTheDay(time.Now())
	for date.Before(now) {
		date = date.Add(time.Hour * 24)
		days++
		if days > daysInLastSixMonths {
			return outOfRange
		}
	}
	return days
}

func getBeginningOfTheDay(t time.Time) time.Time {
	year, month, day := t.Date()
	beginningOfTheDay := time.Date(year, month, day, 0, 0, 0, 0, t.Location())
	return beginningOfTheDay
}

func render(commits map[int]int) {
	keys := sortKeys(commits)
	cols := generateCols(keys, commits)
	renderStats(cols)
}

func sortKeys(commits map[int]int) []int {
	var keys []int
	for key := range commits {
		keys = append(keys, key)
	}
	sort.Ints(keys)
	return keys
}

func generateCols(keys []int, commits map[int]int) map[int]column {
	cols := make(map[int]column)
	col := column{}
	for _, key := range keys {
		week := key / 7
		dayInWeek := key % 7
		if dayInWeek == 0 {
			col = column{}
		}
		col = append(col, commits[key])
		if dayInWeek == 6 {
			cols[week] = col
		}
	}
	return cols
}

func renderStats(cols map[int]column) {
	renderMonths(calcOffset())
	for j := 6; j >= 0; j-- {
		for i := weekInLastSixMonths + 1; i >= 0; i-- {
			if i == weekInLastSixMonths+1 {
				renderWeekday(j)
			}
			if col, ok := cols[i]; ok {
				if i == 0 && j == calcOffset()-1 {
					renderSingleCell(col[j], true)
					continue
				} else {
					if len(col) > j {
						renderSingleCell(col[j], false)
						continue
					}
				}
			}
			renderSingleCell(0, false)
		}
		fmt.Printf("\n")
	}
}

func renderMonths(offset int) {
	week := getBeginningOfTheDay(time.Now()).Add(time.Duration(offset) * time.Hour * 24).Add(-(viewsColsNum * time.Hour * 24 * 7))
	month := week.Month()
	fmt.Printf(" ")
	//fmt.Printf("%s ", month.String()[:3])
	for {
		if week.Month() != month {
			fmt.Printf(" %s", week.Month().String()[:3])
			month = week.Month()
		} else {
			fmt.Printf("    ")
		}
		week = week.Add(7 * time.Hour * 24)
		if week.After(time.Now()) {
			break
		}
	}
	fmt.Printf("\n")
}

func renderWeekday(weekday int) {
	str := "     "
	switch weekday {
	case 5:
		str = " Mon "
	case 3:
		str = " Wed "
	case 1:
		str = " Fri "
	}
	fmt.Printf(str)
}

func renderSingleCell(val int, isToday bool) {
	escape := "\033[0;37;30m"
	switch {
	case val > 0 && val < 5:
		escape = "\033[1;30;47m"
	case val >= 5 && val < 10:
		escape = "\033[1;30;43m"
	case val >= 10:
		escape = "\033[1;30;42m"
	}
	if isToday {
		escape = "\033[1;37;45m"
	}
	if val == 0 {
		fmt.Printf(escape + "  - " + "\033[0m")
		return
	}
	str := "  %d "
	switch {
	case val >= 10:
		str = " %d "
	case val >= 100:
		str = "%d "
	}
	fmt.Printf(escape+str+"\033[0m", val)
}
