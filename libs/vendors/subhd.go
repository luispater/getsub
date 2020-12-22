package vendors

import (
	"bytes"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/luispater/getsub/common"
	"math"
	"regexp"
	"strconv"
	"strings"
)

type SubHD struct {
}

func (this *SubHD) Init() error {
	return nil
}

func (this *SubHD) Search(keyword string) (*SubtitleResult, error) {
	result := new(SubtitleResult)
	result.Subtitles = make([]Subtitle, 0)

	for i := 1; i <= 100; i++ {
		var url string
		if i == 1 {
			url = fmt.Sprintf("https://subhd.tv/search/%s", keyword)
		} else {
			url = fmt.Sprintf("https://subhd.tv/search/%s/%d", keyword, i)
		}

		byteHtml, err := common.HttpGet(url)
		if err != nil {
			return nil, err
		}

		doc, errNewDocumentFromReader := goquery.NewDocumentFromReader(bytes.NewReader(byteHtml))
		if errNewDocumentFromReader != nil {
			return nil, errNewDocumentFromReader
		}

		var totalRecord int64
		doc.Find("small").Find("b").Each(func(i int, s *goquery.Selection) {
			if i == 0 {
				totalRecord, err = strconv.ParseInt(s.Text(), 10, 64)
			}
		})

		doc.Find(".mb-4.bg-white").Each(func(i int, s *goquery.Selection) {
			var subtitle Subtitle
			titleLink := s.Find("a[data-toggle='tooltip']")
			title := titleLink.Text()
			matchFile, exist := titleLink.Attr("title")
			if !exist {

			}
			href, exist2 := titleLink.Attr("href")
			if !exist2 {

			}
			var language string
			var publisher string
			var publishTime string
			s.Find(".pt-1.text-secondary").Each(func(i int, s *goquery.Selection) {
				if i == 0 {
					language = s.Text()
					language = regexp.MustCompile(`\t`).ReplaceAllString(language, "")
					language = strings.TrimSpace(language)
				} else if i == 1 {
					publish := s.Text()
					publish = regexp.MustCompile(`\t`).ReplaceAllString(publish, "")
					publish = regexp.MustCompile(`\s+`).ReplaceAllString(publish, " ")
					publish = strings.TrimSpace(publish)
					arrayPublish := strings.SplitN(publish, " ", 3)
					if len(arrayPublish) > 2 {
						publisher = arrayPublish[0]
						publishTime = arrayPublish[2]
					} else {
						publisher = arrayPublish[0]
					}
				}
			})
			group := s.Find(".float-right.py-1.px-2.rounded-sm").Text()

			subtitle.Id = href
			subtitle.Title = title
			subtitle.Author = publisher
			subtitle.PublishTime = publishTime
			subtitle.Extension = make([]SubtitleExtension, 0)
			subtitle.Extension = append(subtitle.Extension, SubtitleExtension{Name: "语言", Value: language})
			subtitle.Extension = append(subtitle.Extension, SubtitleExtension{Name: "字幕组", Value: group})
			subtitle.Extension = append(subtitle.Extension, SubtitleExtension{Name: "视频版本", Value: matchFile})
			result.Subtitles = append(result.Subtitles, subtitle)
		})
		if int64(math.Ceil(float64(totalRecord)/20.0)) <= int64(i) {
			break
		}
	}

	return result, nil
}

func (this *SubHD) DownloadFile(Id, filePath string) ([]byte, error) {
	return nil, nil
}

func (this *SubHD) UnArchiveFile(archiveFilePath, filename, toFilename string) ([]byte, error) {
	return nil, nil
}
