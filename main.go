package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const (
	NameReplaceStr    = "$name"
	EpisodeReplaceStr = "$episode"
	DefaultRuleAnime  = "$name - $episode"
	DefaultRuleMovie  = "$name"
)

var (
	sourceDir      = flag.String("src", "", "source dir")
	destinationDir = flag.String("dst", "", "destination dir")
	ruleFlag       = flag.String("rule", "", "episode naming rule")
	mode           = flag.String("mode", "anime", "mode: anime or movie")
	rule           = DefaultRuleAnime

	videoSuffix = []string{
		".mkv",
		".mp4",
		".avi",
		".m2ts",
	}

	otherSuffix = []string{
		".ass",
		".srt",
		".mka",
	}

	deleteRegex = []string{
		`\[.*?\]`, //find strings between []
		`\(.*?\)`, //find strings between ()
		`【.*?】`,   //find strings between 【】
		`（.*?）`,   //find strings between （）
		`<.*?>`,   //find strings between <>
		`1080[pP]`,
		`2160[pP]`,
		`4[kK]`,
		`[Bb]lu[Rr]ay`,
		`BLURAY`,
	}

	deleteChar = []string{
		" - ",
	}

	scanner *bufio.Scanner
)

func getLine() string {
	if scanner.Scan() {
		return scanner.Text()
	}

	return ""
}

func getSplitPath(path string) (string, string) {
	index := strings.LastIndexFunc(path, func(r rune) bool {
		if r == '/' || r == '\\' {
			return true
		}

		return false
	})

	if index < 0 {
		return "", path
	}

	return path[:index], path[index+1:]
}

func deletePatterns(name string) string {
	for _, str := range deleteRegex {
		regex := regexp.MustCompile(str)
		name = regex.ReplaceAllString(name, "")
	}

	if *mode == "movie" {
		name = strings.ReplaceAll(name, ".", " ")
	}

	return name
}

func probeVideoName(name string) string {
	name, ext := getExtName(name)
	origName := name

	//delete patterns
	name = deletePatterns(name)
	name = strings.TrimSpace(name)

	if name == "" {
		name = origName

		regex := regexp.MustCompile(`\[.+?\]`)
		fields := regex.FindAllString(name, -1)

		newFields := make([]string, 0)
		for _, str := range fields {
			str = str[1 : len(str)-1]
			str = strings.TrimSpace(str)
			if str != "" {
				newFields = append(newFields, str)
			}
		}

		if len(newFields) == 0 {
			return ""
		} else if len(newFields) == 1 {
			name = newFields[0]
		} else {
			name = newFields[1]
		}
	}

	//delete chars
	for _, char := range deleteChar {
		name = strings.ReplaceAll(name, char, " ")
	}

	//delete EP number
	regex := regexp.MustCompile(`((\[?(CM|OVA|#)?\d{1,3}(v\d{1,2}|\.\d{1,2})?\]?)|(\[?第\d{1,3}(v\d{1,2}|\.\d{1,2})?[话話]\]?))`)
	name = regex.ReplaceAllString(name, "")
	name = strings.TrimSpace(name)

	if *mode == "movie" {
		ssIndex := strings.Index(name, "  ")
		if ssIndex > 0 {
			name = name[:ssIndex]
		}
	}

	return name + ext
}

func getEpisode(name string) string {
	name, _ = getExtName(name)

	exxRegex := regexp.MustCompile(`[Ee][Pp]?\d{1,3}`)
	exxStr := exxRegex.FindString(name)
	if exxStr != "" {
		exxStr = regexp.MustCompile("([EePp ]|-)").ReplaceAllString(exxStr, "")
		return exxStr
	}

	numbers := ""

	//detect [01], [OVA1], 第01話, [第01話], etc.
	regex := regexp.MustCompile(`\[((第\d{1,3}(v\d{1,2}|\.\d{1,2})?[话話])|((CM|OVA|#)?\d{1,3}(v\d{1,2}|\.\d{1,2})?))\]`)
	numbersSlice := regex.FindAllString(name, -1)
	if len(numbersSlice) > 0 {
		numbers = numbersSlice[len(numbersSlice)-1]
		numbers = regexp.MustCompile(`[\[\]第话話#]`).ReplaceAllString(numbers, "")
	}

	name = deletePatterns(name)
	name = strings.TrimSpace(name)

	if numbers == "" {
		//detect - 01, -12.5, etc.
		regex = regexp.MustCompile(`\s*-\s*((第\d{1,3}(v\d{1,2}|\.\d{1,2})?[话話])|((CM|OVA|#)?\d{1,3}(v\d{1,2}|\.\d{1,2})?))`)
		numbersSlice = regex.FindAllString(name, -1)
		if len(numbersSlice) > 0 {
			numbers = numbersSlice[len(numbersSlice)-1]
			numbers = regexp.MustCompile(`[-第话話#]`).ReplaceAllString(numbers, "")
			numbers = strings.TrimSpace(numbers)
		}
	}

	if numbers == "" {
		regex = regexp.MustCompile(`\s+((第\d{1,3}(v\d{1,2}|\.\d{1,2})?[话話])|((CM|OVA|#)?\d{1,3}(v\d{1,2}|\.\d{1,2})?))`)
		numbersSlice = regex.FindAllString(name, -1)
		if len(numbersSlice) > 0 {
			numbers = numbersSlice[len(numbersSlice)-1]
			numbers = regexp.MustCompile(`[第话話#]`).ReplaceAllString(numbers, "")
			numbers = strings.TrimSpace(numbers)
		}
	}

	return numbers
}

func getSeason(name string) string {
	name, _ = getExtName(name)

	sxxRegex := regexp.MustCompile(`[Ss]\d{1,3}}`)
	sxxStr := sxxRegex.FindString(name)
	if sxxStr != "" {
		return sxxStr
	}

	return "S01"
}

func checkExtName(ext string) bool {
	if len(ext) > 6 { //long ext names
		return false
	}

	ext = ext[1:]
	for _, c := range ext {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9') {
			return false
		}
	}

	return true
}

func getExtName(name string) (string, string) {
	extname := path.Ext(name)

	if extname == "" {
		return name, ""
	}

	if !checkExtName(extname) {
		return name, ""
	}
	name = name[:len(name)-len(extname)]

	//handle "[02].sc.ass" etc.
	extname2 := path.Ext(name)
	if extname2 == "" {
		return name, extname
	}

	if !checkExtName(extname2) {
		return name, extname
	}

	name2 := name[:len(name)-len(extname2)]

	matchList := []string{
		".sc", ".tc", ".chs", ".cht", ".en", ".jp",
	}

	f := false
	for _, match := range matchList {
		if extname2 == match {
			f = true
			break
		}
	}

	if f {
		return name2, extname2 + extname
	} else {
		return name, extname
	}
}

func getDirName(name string) string {
	dir := path.Dir(name)

	if dir == "" || dir == "." {
		index := strings.LastIndex(name, string(os.PathSeparator))
		if index >= 0 {
			dir = name[:index]
		} else {
			dir = name
		}
	}

	return dir
}

func getVideosInDir(dir string) []string {
	//get all files and directories in dir
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		fmt.Printf("Cannot read dir %s. error: %s.\n", dir, err.Error())
		return []string{}
	}

	//get video files list
	videos := make([]string, 0)
	for _, file := range files {
		if !file.IsDir() {
			filename := file.Name()

			for _, suffix := range videoSuffix {
				if strings.HasSuffix(filename, suffix) {
					videos = append(videos, filename)
				}
			}

			for _, suffix := range otherSuffix {
				if strings.HasSuffix(filename, suffix) {
					videos = append(videos, filename)
				}
			}
		}
	}

	return videos
}

func checkFileExists(file string) bool {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return false
	} else if err == nil {
		return true
	} else {
		fmt.Printf("os.Stat unknown error:%s.\n", err.Error())
		os.Exit(1)
		return false
	}
}

func checkDirEmpty(dir string) bool {
	if checkFileExists(dir) {
		//dir exists
		files, err := ioutil.ReadDir(dir)
		if err != nil {
			fmt.Printf("Cannot read dir %s. error: %s.\n", dir, err.Error())
			os.Exit(1)
			return false
		}

		if len(files) > 0 {
			return false
		} else {
			return true
		}
	} else {
		return true
	}
}

func getVideosCount(videos []string) int {
	count := 0

	for _, name := range videos {
		extname := path.Ext(name)

		validExt := false
		for _, suffix := range videoSuffix {
			if suffix == extname {
				validExt = true
				break
			}
		}

		if validExt {
			count++
		}
	}

	return count
}

func generatesVideoNames(videos, episodes []string, manual bool) (newFilenames []string) {
	newFilenames = make([]string, len(videos))

	for i, video := range videos {
		if episodes[i] == "" && *mode == "anime" || episodes[i] == "$$$$$" {
			newFilenames[i] = "(Not linking)"
			continue
		}

		newName := rule

		var extName string
		video, extName = getExtName(video)
		video = strings.TrimSpace(video)
		episode := episodes[i]

		newName = strings.ReplaceAll(newName, NameReplaceStr, video)
		newName = strings.ReplaceAll(newName, EpisodeReplaceStr, episode)
		newName = strings.TrimSpace(newName)
		newName += extName

		newName = strings.TrimSpace(newName)

		if *mode == "movie" && getVideosCount(videos) > 1 && !manual {
			newName = path.Join(video, newName)
		}

		newFilenames[i] = newName
	}

	return
}

func manualLink(videos, episodes, names []string, origLinkDir string) (newVideos, newEpisodes []string, linkDir string) {
	var input string

	fmt.Println()
	defer fmt.Println()

	fmt.Printf("Input name: [%s] ", origLinkDir)
	input = getLine()

	videoName := origLinkDir

	if input != "" {
		videoName = input
	}

	linkDir = videoName

	fmt.Printf("Input link directory: [%s] ", videoName)
	input = getLine()

	if input != "" {
		linkDir = input
	}

	newVideos = make([]string, len(videos))
	newEpisodes = make([]string, len(episodes))

	season := "S01"
	seasonPrompt := true

	if *mode == "anime" {
		for i, name := range names {
			_, ext := getExtName(name)
			episode := episodes[i]

			if episode == "" {
				episode = "$"
			}

			fmt.Printf("Input episode name of '%s' ($ for not linking): [%s] ", name, episode)
			input = getLine()

			if input != "" {
				episode = input
			}

			if episode != "$" {
				if seasonPrompt {
					fmt.Printf("Input season name of '%s' (# for empty, ! for all %s): [%s] ", name, season, season)
					input2 := getLine()

					if input2 == "!" {
						seasonPrompt = false
					} else if input2 == "#" {
						season = ""
					} else if input2 != "" {
						season = input2
					}
				}

			}

			newVideos[i] = videoName + ext
			if season != "" {
				newVideos[i] = path.Join(season, newVideos[i])
			}

			if episode == "$" {
				newEpisodes[i] = ""
			} else {
				newEpisodes[i] = episode
			}

		}
	} else if *mode == "movie" {
		if getVideosCount(names) <= 1 {
			for i, name := range names {
				_, ext := getExtName(name)
				newVideos[i] = videoName + ext
				newEpisodes[i] = ""
			}
		} else {
			for i, name := range names {
				_, ext := getExtName(name)

				movieName := probeVideoName(name)
				movieName, _ = getExtName(movieName)

				if episodes[i] == "$$$$$" {
					movieName = "$"
				}

				fmt.Printf("Input movie name of '%s' ($ for not linking): [%s] ", name, movieName)
				input = getLine()

				if input != "" {
					movieName = input
				}

				if movieName == "$" {
					newVideos[i] = videos[i]
					newEpisodes[i] = "$$$$$"
					continue
				}

				movieLinkDir := movieName

				fmt.Printf("Input link directory of '%s' (# for top directory): [%s] ", name, movieName)
				input = getLine()

				if input != "" {
					movieLinkDir = input
				}

				if input == "#" {
					newVideos[i] = movieName + ext
				} else {
					newVideos[i] = path.Join(movieLinkDir, movieName+ext)
				}

				newEpisodes[i] = ""
			}
		}
	}

	return
}

func probeDirInner(dir, destDir string, videos []string, level int, origDestDir string) {
	var prompt string

	if videos == nil {
		videos = getVideosInDir(dir)
	}

	if len(videos) == 0 {
		fmt.Println("No videos found.")
		fmt.Println()
		return
	}

	_, dirName := getSplitPath(dir)
	animeName := probeVideoName(dirName)
	animeName = strings.TrimSpace(animeName)
	if animeName == "" {
		animeName = "Unknown"
	}

	//check directory empty.
	//if not empty, ask user if he wants to create a subdirectory.
	//useful in linking just one movie folder.
	if !checkDirEmpty(destDir) && level == 0 {
		fmt.Println()
		fmt.Printf("Directory %s not empty. Do you need create a sub-directory in it? [Y/n] ", destDir)
		prompt = getLine()

		if prompt != "n" && prompt != "N" {
			destDir = path.Join(destDir, animeName)
			origDestDir = path.Join(origDestDir, animeName)
		}
	}

	newVideos := make([]string, len(videos))
	episodes := make([]string, len(videos))

	for i, videoName := range videos {
		_, ext := getExtName(videoName)
		newName := animeName + ext
		episode := getEpisode(videoName)

		if *mode == "anime" {
			season := getSeason(videoName)
			newName = path.Join(season, newName)
		}

		newVideos[i] = newName
		episodes[i] = episode
	}

	linkWithNewNames := true

	var newFilenames []string

	flagManualLink := false

	for {
		if checkFileExists(destDir) {
			fmt.Printf("[WARNING] Directory '%s' already exists!\n", destDir)
		}

		newFilenames = generatesVideoNames(newVideos, episodes, flagManualLink)

		fmt.Println()

		fmt.Printf("[DIRECTORY] %s => %s\n", dir, destDir)

		for i, newName := range newFilenames {
			oldName := videos[i]

			fmt.Printf("[VIDEO] %s => %s\n", oldName, newName)
		}

		fmt.Println()

		prompt = ""
		for prompt != "y" && prompt != "n" && prompt != "Y" && prompt != "N" {
			fmt.Printf("Is that right? [Y/n] ")

			prompt = getLine()

			if prompt == "n" || prompt == "N" {
				linkWithNewNames = false
			}
		}

		if prompt == "y" || prompt == "Y" {
			break
		}

		if !linkWithNewNames {
			prompt = ""
			fmt.Printf("Manually edit file names? [Y/n] ")
			prompt = getLine()

			if prompt == "n" || prompt == "N" {
				prompt = ""
				fmt.Printf("Link with original names? [Y/n] ")
				prompt = getLine()

				if prompt == "n" || prompt == "N" {
					return
				} else {
					newVideos = videos
					destDir = origDestDir

					linkWithNewNames = true
				}
			} else {
				var linkDir string
				_, linkDir = getSplitPath(destDir)
				newVideos, episodes, linkDir = manualLink(newVideos, episodes, videos, linkDir)

				oldDir := getDirName(origDestDir)
				destDir = path.Join(oldDir, linkDir)

				linkWithNewNames = true
				flagManualLink = true
			}
		}
	}

	fmt.Println("Now linking the files...")

	if !linkWithNewNames {
		_, dirName := getSplitPath(dir)
		dirBase, _ := getSplitPath(destDir)
		destDir = path.Join(dirBase, dirName)
	}

	if _, err := os.Stat(destDir); os.IsNotExist(err) {
		fmt.Printf("dest not exists, creating.\n")
		err2 := os.MkdirAll(destDir, 0777)
		if err2 != nil {
			fmt.Printf("os.MkdirAll error: %s.\n", err2.Error())
			os.Exit(1)
			return
		}
	}

	for i, newName := range newFilenames {
		if episodes[i] == "" && *mode == "anime" || episodes[i] == "$$$$$" {
			//omitted video
			continue
		}

		oldName := videos[i]

		oldPath := path.Join(dir, oldName)
		newPath := ""
		if linkWithNewNames {
			newPath = path.Join(destDir, newName)
		} else {
			newPath = path.Join(destDir, oldName)
		}

		newPathDir, _ := getSplitPath(newPath)
		if _, err := os.Stat(newPathDir); os.IsNotExist(err) {
			err2 := os.MkdirAll(newPathDir, 0777)
			if err2 != nil {
				fmt.Printf("os.MkdirAll error: %s.\n", err2.Error())
				os.Exit(1)
				return
			}
		}

		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			err2 := os.Link(oldPath, newPath)
			if err2 != nil {
				fmt.Printf("Link error: %s.\n", err2.Error())
				os.Exit(1)
				return
			}
		} else if err == nil {
			for i := 2; i <= 99; i++ {
				var newName2, extName string
				if linkWithNewNames {
					newName2, extName = getExtName(newName)
				} else {
					newName2, extName = getExtName(oldName)
				}

				newName2 += " (" + strconv.Itoa(i) + ")"
				newName2 += extName
				newPath2 := path.Join(destDir, newName2)

				if _, err2 := os.Stat(newPath2); os.IsNotExist(err2) {
					err3 := os.Link(oldPath, newPath2)
					if err3 != nil {
						fmt.Printf("Link error: %s.\n", err3.Error())
						os.Exit(1)
						return
					} else {
						break
					}
				}
			}
		} else {
			fmt.Printf("os.Stat unknown error:%s.\n", err.Error())
			os.Exit(1)
			return
		}
	}
}

func probeDir(dir, destDir string) {
	videos := getVideosInDir(dir)

	//check video files exists
	if len(videos) > 0 {
		probeDirInner(dir, destDir, videos, 0, destDir)
	} else {
		//search for subdirectories
		files, err := ioutil.ReadDir(dir)
		if err != nil {
			fmt.Printf("Cannot read dir %s. error: %s.\n", dir, err.Error())
			os.Exit(1)
			return
		}

		//sort by modify time desc
		sort.Slice(files, func(i, j int) bool {
			return files[i].ModTime().After(files[j].ModTime())
		})

		for _, file := range files {
			if file.IsDir() {
				dirName := file.Name()

				fmt.Printf("Search into %s? [y/N] ", dirName)
				var prompt string
				prompt = getLine()

				if prompt == "y" || prompt == "Y" {
					destDir2 := probeVideoName(dirName)
					destDir2 = strings.TrimSpace(destDir2)
					if destDir2 == "" {
						destDir2 = "Unknown"
					}
					destDir2 = path.Join(destDir, destDir2)
					origDestDir := path.Join(destDir, dirName)

					srcDir := path.Join(dir, dirName)

					probeDirInner(srcDir, destDir2, nil, 1, origDestDir)
				}
			}
		}
	}
}

func main() {
	flag.Parse()

	if *sourceDir == "" {
		fmt.Println("src must not be empty")
		os.Exit(1)
	}

	if *destinationDir == "" {
		fmt.Println("dst must not be empty")
		os.Exit(1)
	}

	if *mode != "anime" && *mode != "movie" {
		fmt.Println("mode must be anime or movie")
		os.Exit(1)
	}

	if *ruleFlag == "" {
		if *mode == "movie" {
			rule = DefaultRuleMovie
		}
	} else {
		rule = *ruleFlag
	}

	scanner = bufio.NewScanner(os.Stdin)

	probeDir(*sourceDir, *destinationDir)
}
