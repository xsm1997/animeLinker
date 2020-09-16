package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
)

const (
	NameReplaceStr = "$name"
	EpisodeReplaceStr = "$episode"
)

var (
	sourceDir = flag.String("src", "", "source dir")
	destinationDir = flag.String("dst", "", "destination dir")
	rule = flag.String("rule", "$name - $episode", "episode naming rule")

	videoSuffix = []string {
		".mkv",
		".mp4",
		".avi",
		".ass",
		".srt",
		".mka",
	}

	deleteRegex = []string {
		`\[.*?\]`, //find strings between []
		`\(.*?\)`, //find strings between ()
		`【.*?】`, //find strings between 【】
		`（.*?）`, //find strings between （）
		`<.*?>`, //find strings between <>
	}

	deleteChar = []string {
		" - ",
	}
)

func isDigit(str string) bool {
	for _, s := range str {
		if !(s >= '0' && s <= '9') {
			return false
		}
	}

	return true
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
			sRune = sRune[1:len(sRune)-1]

			if isDigit(string(sRune)) {
				return s
			}

			return ""
		})
	}

	//delete chars
	for _, char := range deleteChar {
		name = strings.ReplaceAll(name, char, "")
	}
	return name
}

func getEpisode(name string) string {
	name, _ = getExtName(name)
	episode := ""

	if strings.Contains(name, "OVA") {
		episode = "OVA"
	}

	if strings.Contains(name, "CM") {
		episode = "CM"
	}

	regex := regexp.MustCompile(`\d+`) //find the last number in string
	numbersSlice := regex.FindAllString(name, -1)
	numbers := ""
	if len(numbersSlice) > 0 {
		numbers = numbersSlice[len(numbersSlice)-1]
	}

	if numbers == "" && episode == "" {
		numbers = "Unknown"
	}

	return episode + numbers
}

func deleteEpisodeName(name, episode string) string {
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
	dotIndex := strings.LastIndex(name, ".")
	if dotIndex < 0 {
		return name, ""
	}

	extname := name[dotIndex:]
	name = name[:dotIndex]

	dotIndex = strings.LastIndex(name, ".")
	if dotIndex < 0 {
		return name, extname
	}

	//handle "[02].sc.ass" etc.
	extname2 := name[dotIndex:]
	name2 := name[:dotIndex]
	suffixList := []string {
		" ", ")", "]", ">", "】", "）",
	}

	f := false
	for _, suffix := range suffixList {
		if strings.HasSuffix(name2, suffix) {
			f = true
			break
		}
	}

	if f {
		return name2, extname2+extname
	} else {
		return name, extname
	}
}

func probeDir(dir, destDir string, inner bool) {
	//get all files and directories in dir
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		fmt.Printf("Cannot read dir %s. error: %s.\n", dir, err.Error())
		os.Exit(1)
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
		}
	}

	//check video files exists
	if len(videos) > 0 {
		newVideos := make([]string, len(videos))
		episodes := make([]string, len(videos))

		for i, videoName := range videos {
			newName := probeVideoName(videoName)
			episode := getEpisode(newName)
			newName = deleteEpisodeName(newName, episode)

			newVideos[i] = newName
			episodes[i] = episode
		}

		newFilenames := make([]string, len(videos))
		for i, video := range newVideos {
			newName := *rule
			var extName string
			video, extName = getExtName(video)
			video = strings.TrimSpace(video)

			newName = strings.ReplaceAll(newName, NameReplaceStr, video)
			newName = strings.ReplaceAll(newName, EpisodeReplaceStr, episodes[i])
			newName = strings.TrimSpace(newName)
			newName += extName

			newName = strings.TrimSpace(newName)

			newFilenames[i] = newName
		}

		fmt.Printf("[DIRECTORY] %s => %s\n", dir, destDir)

		for i, newName := range newFilenames {
			oldName := videos[i]

			fmt.Printf("[VIDEO] %s => %s\n", oldName, newName)
		}

		fmt.Printf("Is that right? [Y/n] ")
		var prompt string
		fmt.Scanln(&prompt)

		linkWithNewNames := true

		if prompt == "n" || prompt == "N" {
			linkWithNewNames = false
		}

		fmt.Println("Now linking the files...")

		if !linkWithNewNames {
			_, dirName := path.Split(dir)
			dirBase, _ := path.Split(destDir)
			destDir = path.Join(dirBase, dirName)
		}

		if _, err := os.Stat(destDir); os.IsNotExist(err) {
			fmt.Printf("dest not exists, creating.\n")
			err2 := os.MkdirAll(destDir, 0666)
			if err2 != nil {
				fmt.Printf("os.MkdirAll error: %s.\n", err2.Error())
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


			if _, err := os.Stat(newPath); os.IsNotExist(err) {
				err2 := os.Link(oldPath, newPath)
				if err2 != nil {
					fmt.Printf("Link error: %s.\n", err2.Error())
				}
			} else if err == nil {
				for i:=2; i<=99; i++ {
					newName2, extName := getExtName(newName)
					newName2 += " (" + strconv.Itoa(i) + ")"
					newName2 += extName
					newPath2 := path.Join(destDir, newName2)

					if _, err2 := os.Stat(newPath2); os.IsNotExist(err2) {
						err3 := os.Link(oldPath, newPath2)
						if err3 != nil {
							fmt.Printf("Link error: %s.\n", err3.Error())
						} else {
							break
						}
					}
				}
			} else {
				fmt.Printf("os.Stat unknown error:%s.\n", err.Error())
			}
		}
	} else {
		if inner {
			return
		}
		//search for subdirectories
		for _, file := range files {
			if file.IsDir() {
				dirName := file.Name()

				fmt.Printf("Search into %s? [y/N] ", dirName)
				var prompt string
				fmt.Scanln(&prompt)

				if prompt == "y" || prompt == "Y" {
					destDir2 := probeVideoName(dirName)
					destDir2 = strings.TrimSpace(destDir2)
					destDir2 = path.Join(destDir, destDir2)

					srcDir := path.Join(dir, dirName)

					probeDir(srcDir, destDir2, true)
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

	probeDir(*sourceDir, *destinationDir, false)
}
