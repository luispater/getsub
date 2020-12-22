package main

import (
	"errors"
	"fmt"
	"github.com/luispater/getsub/libs/vendors"
	"github.com/manifoldco/promptui"
	"strconv"
)

func main() {

	subHd := new(vendors.SubHD)
	result, err := subHd.Search("The.Mandalorian.S02E08")
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
	prompt := promptui.Prompt{
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

	fmt.Printf("You choose %q\n", promptResult)
}
