package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/user"
	"strings"
)

func scan(folder string) {
	repositories := recursiveScanFolders(folder)
	filePath := getFilePath()
	addNewRepositories(repositories, filePath)
}

func recursiveScanFolders(folder string) []string {
	return scanGitFolders(make([]string, 0), folder)
}

func scanGitFolders(folders []string, folder string) []string {
	folder = strings.TrimSuffix(folder, "/")
	f, err := os.Open(folder)
	if err != nil {
		log.Fatal(err)
	}
	files, err := f.Readdir(-1)
	if err != nil {
		log.Fatal(err)
	}
	err = f.Close()
	if err != nil {
		log.Fatal(err)
	}
	var path string
	for _, file := range files {
		if file.IsDir() {
			path = folder + "/" + file.Name()
			if file.Name() == ".git" {
				path = strings.TrimSuffix(path, "/.git")
				fmt.Println(path)
				folders = append(folders, path)
				continue
			}
			if file.Name() == "vendor" || file.Name() == "node_modules" {
				continue
			}
			folders = scanGitFolders(folders, path)
		}
	}
	return folders
}

func getFilePath() string {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	return usr.HomeDir + "\\.my_git_stats"
}

func addNewRepositories(repositories []string, filePath string) {
	existRepos := resolveFile(filePath)
	repos := addNewRepos(existRepos, repositories)
	writeBackToFile(repos, filePath)
}

func resolveFile(filePath string) []string {
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_RDWR, 0755)
	if err != nil {
		if os.IsNotExist(err) {
			_, err := os.Create(filePath)
			if err != nil {
				panic(err)
			}
		} else {
			panic(err)
		}
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			panic(err)
		}
	}(f)
	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err = scanner.Err(); err != nil {
		if err != io.EOF {
			//log.Fatal(err)
			panic(err)
		}
	}
	return lines
}

func addNewRepos(existRepos []string, newRepos []string) []string {
	for _, repo := range newRepos {
		if !containRepo(existRepos, repo) {
			existRepos = append(existRepos, repo)
		}
	}
	return existRepos
}

func containRepo(repoList []string, repo string) bool {
	for _, existRepo := range repoList {
		if existRepo == repo {
			return true
		}
	}
	return false
}

func writeBackToFile(repos []string, filePath string) {
	text := strings.Join(repos, "\n")
	err := os.WriteFile(filePath, []byte(text), 0755)
	if err != nil {
		panic(err)
	}
}
