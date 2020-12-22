package main

import (
	"errors"
	"fmt"
	"github.com/luispater/getsub/libs/vendors"
	"github.com/manifoldco/promptui"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
)

func main() {
	defaultVideoName := filepath.Base(os.Args[1])

	videoNameReg, _ := regexp.Compile(`S\d+E\d+`)
	videoNameIndex := videoNameReg.FindStringIndex(defaultVideoName)
	if len(videoNameIndex) > 0 {
		defaultVideoName = defaultVideoName[0:videoNameIndex[1]]
	}

	prompt := promptui.Prompt{
		Label:   "视频名称",
		Default: defaultVideoName,
	}

	keyword, err := prompt.Run()

	if err != nil {
		fmt.Println("")
		return
	}

	subHd := new(vendors.SubHD)
	result, err := subHd.Search(keyword)
	if err != nil {
		panic(err)
	}
	fmt.Print("====================\n")
	for i := 0; i < len(result.Subtitles); i++ {
		fmt.Printf("ID: %d\n", i+1)
		fmt.Printf("标题: %s\n", result.Subtitles[i].Title)
		fmt.Printf("发布者: %s\n", result.Subtitles[i].Author)
		fmt.Printf("发布时间: %s\n", result.Subtitles[i].PublishTime)
		for j := 0; j < len(result.Subtitles[i].Extension); j++ {
			fmt.Printf("%s: %s\n", result.Subtitles[i].Extension[j].Name, result.Subtitles[i].Extension[j].Value)
		}
		fmt.Print("====================\n")
	}
	fmt.Println("")
	prompt = promptui.Prompt{
		Label: "要下载的字幕ID",
		Validate: func(input string) error {
			index, errParseInt := strconv.ParseInt(input, 10, 64)
			if errParseInt != nil {
				return errors.New("请输入正确的ID")
			}

			if index > int64(len(result.Subtitles)) {
				return errors.New("请输入正确的ID")
			}
			return nil
		},
	}

	promptResult, err := prompt.Run()

	if err != nil {
		fmt.Println("")
		return
	}

	promptIndex, err := strconv.ParseInt(promptResult, 10, 64)
	if err != nil {
		panic(err)
	}

	filename, byteArchiveFile, err := subHd.DownloadFile(result.Subtitles[promptIndex-1].Id)
	if err != nil {
		panic(err)
	}

	filenames, err := subHd.GetArchiveFileList(filename, byteArchiveFile)
	for i := 0; i < len(filenames); i++ {
		fmt.Printf("%d: %s\n", i+1, filepath.Base(filenames[i]))
	}

	prompt = promptui.Prompt{
		Label: "要保存的字幕ID",
		Validate: func(input string) error {
			index, errParseInt := strconv.ParseInt(input, 10, 64)
			if errParseInt != nil {
				return errors.New("请输入正确的ID")
			}

			if index > int64(len(filenames)) {
				return errors.New("请输入正确的ID")
			}
			return nil
		},
	}

	promptResult, err = prompt.Run()

	if err != nil {
		fmt.Println("")
		return
	}

	promptIndex, err = strconv.ParseInt(promptResult, 10, 64)
	if err != nil {
		panic(err)
	}

	err = subHd.UnArchiveFile(filename, byteArchiveFile, filenames[promptIndex-1], os.Args[1])
	if err != nil {
		panic(err)
	}

	fmt.Println("")
	fmt.Println("下载完成")

	// fmt.Println(string(byteArchiveFile), err)

	// strPwd, err := os.Getwd()
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Printf("You choose %q\n", promptResult)
}
