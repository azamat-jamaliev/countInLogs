package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
)

var debugMode bool

func assert(e error, a ...interface{}) {
	if e != nil {
		color.New(color.FgRed).Println("ERROR:", e.Error())
		if a != nil {
			color.New(color.FgRed).Println("\t Other details:", a)
		}
		panic(e)
	}
}
func printDebug(a ...interface{}) {
	if debugMode {
		fmt.Println("DEBUG:", a)
	}
}
func TrimAll(str string) string {
	return strings.Trim(strings.Trim(strings.Trim(strings.TrimSpace(str), "\n"), "\r"), "\t")
}
func loadIDsToSearch(idsFile string, searchIds *map[string]int) {
	inputIds, err := ioutil.ReadFile(idsFile)
	assert(err, " Reading File with Ids: ", idsFile)
	fileContent := string(inputIds)
	lines := strings.Split(fileContent, "\n")
	for _, line := range lines {
		lnKey := TrimAll(line)
		printDebug("lnKey=", lnKey)
		if len(lnKey) > 2 {
			(*searchIds)[lnKey] = 0
		}
	}
	color.New(color.FgYellow).Println("Number of IDs to search =", len(*searchIds))
}

func main() {
	fmt.Println("********************************************************************************")
	fmt.Println("*                              countInLogs                                     *")
	fmt.Println("********************************************************************************")

	idsFile := flag.String("count_from_file", "", "file with IDs or other lines to serach in the files\nEXAMPLE: -count_from_file ./Ids.txt")
	logsFolder := flag.String("logs_dir", "./log_files", "folder with log files")
	searchRegex := flag.String("search", `\d{19}`, fmt.Sprintf("regex to serach\nEXAMPLE: to count URLS in the log files use:\n -search '%s'\n", `https?:\/\/(www\.)?[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b([-a-zA-Z0-9()@:%_\+.~#?&//=]*)`))

	debugModePt := flag.Bool("debug", false, "enable debug mode with additional output")
	greaterThan := flag.Int("greater_than", 0, "display only records with count greater than the value")
	maxFiles := flag.Int("max_files", 4, "max number of simultaneously opened files, it's recomended to set `max_files` to number of CPU cores (or less)")
	flag.Parse()

	debugMode = *debugModePt
	fmt.Printf("DEBUG MODE:%v\n", debugMode)

	startTime := time.Now()
	color.New(color.FgYellow).Println("NOTE: Starting at: ", startTime.Format(time.RFC850))

	countFromFile := *idsFile != ""
	searchIds := map[string]int{}
	foundIds := map[string]int{"test": 0}
	var resultsMap *map[string]int
	if countFromFile {
		loadIDsToSearch(*idsFile, &searchIds)
		resultsMap = &searchIds
	} else {
		resultsMap = &foundIds
	}

	files, err := ioutil.ReadDir(*logsFolder)
	assert(err, "Reading files from directory: ", *logsFolder)
	totalNumOfFiles := 0
	var m sync.Mutex
	var wg sync.WaitGroup
	var wgFileLimit sync.WaitGroup

	for i := 0; i < len(files); i++ {
		f := files[i]
		if !f.IsDir() {
			for j := 0; j < *maxFiles && i+j < len(files); j++ {
				f = files[i+j]

				totalNumOfFiles++
				chanIds := make(chan string)
				fmt.Println("Number: ", i, " of ", len(files), " name:", f.Name())

				wg.Add(1)
				wgFileLimit.Add(1)
				go Process(path.Join(*logsFolder, f.Name()), *searchRegex, &wg, &wgFileLimit, chanIds)

				wg.Add(1)
				go func(input <-chan string) {
					defer wg.Done()
					for strId := range input {
						m.Lock()
						foundIds[strId]++
						m.Unlock()
					}
				}(chanIds)
			}
			wgFileLimit.Wait()
			i += *maxFiles - 1
		}
	}

	wg.Wait()

	fmt.Println("VVVVVVVVVVVVVVVVVVVVVVVV - Results - VVVVVVVVVVVVVVVVVVVVVVVV")
	if countFromFile { //copy the data to searchIds only we need counts matching to file
		for idToReport, countToReport := range foundIds {
			printDebug("countFromFile section ", idToReport, " found ", countToReport, " times")
			if _, ok := searchIds[idToReport]; ok {
				searchIds[idToReport] = foundIds[idToReport]
			}
		}
	}

	keys := make([]string, 0, len(*resultsMap))

	for key := range *resultsMap {
		keys = append(keys, key)
	}

	sort.SliceStable(keys, func(i, j int) bool {
		return (*resultsMap)[keys[i]] < (*resultsMap)[keys[j]]
	})

	for _, key := range keys {
		if (*resultsMap)[key] > *greaterThan {
			color.New(color.FgGreen).Println("", key, "\t\t\tfound ", (*resultsMap)[key], " times")
		}
	}

	endTime := time.Now()
	color.New(color.FgYellow).Printf("Start Time: %v EndTime: %v \nDuration: %v \n", startTime, endTime, endTime.Sub(startTime))

}

func Process(filePath, searchRegex string, wg *sync.WaitGroup, wgFileLimit *sync.WaitGroup, output chan<- string) {
	defer wg.Done()
	file, err := os.Open(filePath)
	assert(err, "Openning File: ", filePath)
	defer file.Close()
	defer wgFileLimit.Done()

	r := bufio.NewReader(file)
	lineNum := 0

	var wgCloseChan sync.WaitGroup
	for {
		lineNum++
		nextUntillNewline, err := r.ReadBytes('\n')
		if err == io.EOF {
			break
		}
		assert(err, "Reading line #", lineNum, " from file:", filePath)

		wgCloseChan.Add(1)
		go func(line string, output2 chan<- string) {
			r := regexp.MustCompile(searchRegex)
			for _, match := range r.FindAllString(line, -1) {
				// printDebug("FOUND match=", match)
				output2 <- TrimAll(match)
			}
			wgCloseChan.Done()
		}(string(nextUntillNewline), output)
	}
	go func() {
		wgCloseChan.Wait()
		close(output)
	}()
}
