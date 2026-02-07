package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
)

type SearchResult struct {
	FileName string
	LineNum  int
	Text     string
}

type Notifier interface {
	Notify(result SearchResult)
}
type ConsoleNotifier struct{}

func (cn ConsoleNotifier) Notify(result SearchResult) {
	fmt.Println("Found in file: ", result.FileName, " at line: ", result.LineNum, " with text: ", result.Text)
}

type FileNotifier struct {
	OutputFile string
}

func (fn FileNotifier) Notify(result SearchResult) {
	file, err := os.OpenFile(fn.OutputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer file.Close()
	_, err = file.WriteString(fmt.Sprintf("Found in file: %s at line: %d with text: %s\n", result.FileName, result.LineNum, result.Text))
	if err != nil {
		log.Fatal(err)
	}
}

func search(fileName string, searchStr string, ch chan<- SearchResult, wg *sync.WaitGroup) {
	defer wg.Done()

	file, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
		return
	}

	defer file.Close()
	scanner := bufio.NewScanner(file)
	linenum := 0

	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, searchStr) {
			ch <- SearchResult{
				FileName: fileName,
				LineNum:  linenum,
				Text:     searchStr,
			}
		}
		linenum++
	}

}

func main() {
	fmt.Println("Enter the string to search in the log files:")
	var searchStr string
	fmt.Scanln(&searchStr)
	Dir := "./test"

	var wg sync.WaitGroup

	ch := make(chan SearchResult)
	files, err := os.ReadDir(Dir)
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		wg.Add(1)
		go search(Dir+"/"+file.Name(), searchStr, ch, &wg)
	}
	go func() {
		wg.Wait()
		close(ch)
	}()
	var notifier Notifier
	notifier = FileNotifier{
		OutputFile: "search.log",
	}

	for result := range ch {
		notifier.Notify(result)
	}
}
