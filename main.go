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

	"github.com/dlclark/regexp2"
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
	}

	deleteChar = []string{
		" - ",
	}

	scanner *bufio.Scanner
)

func isDigitOrDot(str string) bool {
	for _, s := range str {
		if !(s >= '0' && s <= '9' || s == '.') {
			return false
		}
	}

	return true
}

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

func probeVideoName(name string) string {
	//delete patterns
	for _, str := range deleteRegex {
		regex := regexp.MustCompile(str)
		name = regex.ReplaceAllStringFunc(name, func(s string) string {
			if strings.Contains(strings.ToUpper(s), "OVA") {
				return s
			}

			if strings.Contains(strings.ToUpper(s), "CM") {
				return s
			}

			sRune := []rune(s)
			sRune = sRune[1 : len(sRune)-1]

			if isDigitOrDot(string(sRune)) {
				return s
			}

			return ""
		})
	}

	//delete chars
	for _, char := range deleteChar {
		name = strings.ReplaceAll(name, char, " ")
	}

	//replace '.' except ext name
	name2, ext := getExtName(name)
	extValid := false

	//probe xxx.sc.ass, to get .ass
	ext2 := path.Ext(ext)
	if ext2 == "" {
		ext2 = ext
	}

	for _, validExtName := range videoSuffix {
		if ext2 == validExtName {
			extValid = true
			break
		}
	}

	if !extValid {
		for _, validExtName := range otherSuffix {
			if ext2 == validExtName {
				extValid = true
				break
			}
		}
	}

	//match any dot except dots in episode name (e.g. 12.5), but match the dot in (.5  12.) for incomplete episode name
	regex := regexp2.MustCompile(`((?<!\d+)\.(?!\d+)|(?<!\d+)\.(?=[0-9]+)|(?<=\d+)\.(?!\d+))`, 0)
	if extValid {
		name2, _ = regex.Replace(name2, " ", -1, -1)
		name2 = strings.TrimSpace(name2)
		return name2 + ext
	} else {
		name2, _ = regex.Replace(name, " ", -1, -1)
		name2 = strings.TrimSpace(name2)
		return name2
	}
}

func getEpisode(name string) string {
	name, _ = getExtName(name)

	sxxexxRegex := regexp.MustCompile(`[Ss]\d{1,3}[Ee]\d{1,3}`)
	sxxexxStr := sxxexxRegex.FindString(name)
	if sxxexxStr != "" {
		return sxxexxStr
	}

	exxRegex := regexp.MustCompile(`[Ee]\d{1,3}`)
	exxStr := exxRegex.FindString(name)
	if exxStr != "" {
		return exxStr
	}

	episode := ""

	if strings.Contains(name, "OVA") {
		episode = "OVA"
	}

	if strings.Contains(name, "CM") {
		episode = "CM"
	}

	regex := regexp.MustCompile(`(\d{1,3}\.\d{1,2}|\d{1,3})`) //find the last number in string
	numbersSlice := regex.FindAllString(name, -1)
	numbers := ""
	if len(numbersSlice) > 0 {
		numbers = numbersSlice[0]
	}

	return episode + numbers
}

func deleteEpisodeName(name, episode string) string {
	if episode == "" {
		return name
	}

	name, extName := getExtName(name)
	split := strings.Fields(name)
	newName := ""

	if len(split) == 1 {
		return name + extName
	} else {
		for _, str := range split {
			if len(str) > 0 {
				if !strings.Contains(str, episode) {
					newName += str + " "
				}
			}

		}
		return newName + extName
	}
}

func getExtName(name string) (string, string) {
	extname := path.Ext(name)
	name = name[:len(name)-len(extname)]

	//handle "[02].sc.ass" etc.
	extname2 := path.Ext(name)
	if extname2 == "" {
		return name, extname
	}
	name2 := name[:len(name)-len(extname2)]

	suffixList := []string{
		" ", ")", "]", ">", "】", "）",
	}

	matchList := []string{
		".sc", ".tc", ".chs", ".cht", ".en", ".jp",
	}

	f := false
	for _, suffix := range suffixList {
		if strings.HasSuffix(name2, suffix) {
			f = true
			break
		}
	}

	f2 := false
	for _, match := range matchList {
		if extname2 == match {
			f2 = true
			break
		}
	}

	if f || f2 {
		return name2, extname2 + extname
	} else {
		return name, extname
	}
}

func getDirName(name string) string {
	dir := path.Dir(name)

	if dir == "" || dir == "." {
		index := strings.LastIndex(name, "\\")
		dir = name[:index]
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
		newName := rule

		var extName string
		video, extName = getExtName(video)
		video = strings.TrimSpace(video)

		newName = strings.ReplaceAll(newName, NameReplaceStr, video)
		newName = strings.ReplaceAll(newName, EpisodeReplaceStr, episodes[i])
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

func manualLink(videos, episodes, names []string) (newVideos, newEpisodes []string, linkDir string) {
	var input string

	fmt.Println()
	defer fmt.Println()

	videoName, _ := getExtName(videos[0])
	videoName = strings.TrimSpace(videoName)

	fmt.Printf("Input name: [%s] ", videoName)
	input = getLine()

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

	if *mode == "anime" {
		for i, name := range names {
			_, ext := getExtName(name)
			episode := episodes[i]

			fmt.Printf("Input episode name of '%s' (# for empty, $ for not linking): [%s] ", name, episode)
			input = getLine()

			if input != "" {
				episode = input
			}

			if input == "#" {
				episode = ""
			}

			if input == "$" {
				newVideos[i] = ""
				newEpisodes[i] = ""
			} else {
				newVideos[i] = videoName + ext
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
				filename, ext := getExtName(name)

				movieName := filename

				fmt.Printf("Input movie name of '%s' ($ for not linking): [%s] ", filename, filename)
				input = getLine()

				if input != "" {
					movieName = input
				}

				if input == "$" {
					newVideos[i] = ""
					newEpisodes[i] = ""
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

	//check directory empty.
	//if not empty, ask user if want to create a sub-directory.
	//useful in linking just one movie folder.
	if !checkDirEmpty(destDir) && level == 0 {
		fmt.Println()
		fmt.Printf("Directory %s not empty. Do you need create a sub-directory in it? [Y/n] ", destDir)
		prompt = getLine()

		if prompt != "n" && prompt != "N" {
			_, dirName := getSplitPath(dir)

			destDir2 := probeVideoName(dirName)
			destDir2 = strings.TrimSpace(destDir2)
			if destDir2 == "" {
				destDir2 = "Unknown"
			}

			destDir = path.Join(destDir, destDir2)
			origDestDir = path.Join(origDestDir, destDir2)
		}
	}

	newVideos := make([]string, len(videos))
	episodes := make([]string, len(videos))

	for i, videoName := range videos {
		newName := probeVideoName(videoName)
		episode := getEpisode(newName)
		newName = deleteEpisodeName(newName, episode)

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
				var newVideoName string
				newVideos, episodes, newVideoName = manualLink(newVideos, episodes, videos)

				oldDir := getDirName(origDestDir)
				destDir = path.Join(oldDir, newVideoName)

				// remove omitted episode
				newVideos_, episodes_, videos_ := newVideos, episodes, videos
				newVideos, episodes, videos = make([]string, 0), make([]string, 0), make([]string, 0)

				for i, newVideo := range newVideos_ {
					if newVideo == "" {
						continue
					}

					newVideos = append(newVideos, newVideo)
					episodes = append(episodes, episodes_[i])
					videos = append(videos, videos_[i])
				}

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

		//sort by modtime desc
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
