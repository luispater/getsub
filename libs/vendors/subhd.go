package vendors

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/luispater/getsub/common"
	"github.com/nwaples/rardecode"
	"github.com/saracen/go7z"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type SubHD struct {
}

type subHDDownAjax struct {
	Success bool   `json:"success"`
	Url     string `json:"url"`
	Msg     string `json:"msg"`
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

func (this *SubHD) DownloadFile(id string) (string, []byte, error) {
	url := fmt.Sprintf("https://subhd.tv%s", id)
	byteHtml, err := common.HttpGet(url)
	if err != nil {
		return "", nil, err
	}

	// fmt.Println(string(byteHtml))
	doc, errNewDocumentFromReader := goquery.NewDocumentFromReader(bytes.NewReader(byteHtml))
	if errNewDocumentFromReader != nil {
		return "", nil, errNewDocumentFromReader
	}
	downButton := doc.Find("#down")
	sid, hasSid := downButton.Attr("sid")
	dToken, hasDToken := downButton.Attr("dtoken1")
	if hasSid && hasDToken {
		byteJson, errHttpPost := common.HttpPost("https://subhd.tv/ajax/down_ajax", map[string]interface{}{
			"sub_id":  sid,
			"dtoken1": dToken,
		})
		if errHttpPost != nil {
			return "", nil, errHttpPost
		}
		var downAjax subHDDownAjax
		err = json.Unmarshal(byteJson, &downAjax)
		if err != nil {
			return "", nil, err
		}
		if downAjax.Success {
			byteData, errHttpGet := common.HttpGet(downAjax.Url)
			return filepath.Base(downAjax.Url), byteData, errHttpGet
		}
		return "", nil, fmt.Errorf("need caption")
	} else {
		return this.openChrome(url)
	}
}

func (this *SubHD) setCookie(name, value, domain, path string, httpOnly, secure bool) chromedp.Action {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		expr := cdp.TimeSinceEpoch(time.Now().Add(180 * 24 * time.Hour))
		success, err := network.SetCookie(name, value).
			WithExpires(&expr).
			WithDomain(domain).
			WithPath(path).
			WithHTTPOnly(httpOnly).
			WithSecure(secure).
			Do(ctx)
		if err != nil {
			return err
		}
		if !success {
			return fmt.Errorf("could not set cookie %s", name)
		}
		return nil
	})
}

func (this *SubHD) openChrome(url string) (string, []byte, error) {
	fileUrl := ""
	ctx := context.Background()
	options := []chromedp.ExecAllocatorOption{
		chromedp.Flag("headless", false),
		chromedp.Flag("hide-scrollbars", false),
		chromedp.Flag("mute-audio", true),
		chromedp.UserAgent(`Mozilla/5.0 (Windows NT 6.3; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/73.0.3683.103 Safari/537.36`),
	}
	options = append(chromedp.DefaultExecAllocatorOptions[:], options...)

	contextAllocator, cancelAllocator := chromedp.NewExecAllocator(ctx, options...)
	ctxNewContext, cancel := chromedp.NewContext(contextAllocator)

	chromedp.ListenTarget(ctxNewContext, func(ev interface{}) {
		switch ev := ev.(type) {
		case *network.EventRequestWillBeSent:
			if strings.Index(ev.Request.URL, "https://dl") != -1 && strings.Index(ev.Request.URL, ".subhd.") != -1 {
				fileUrl = ev.Request.URL
				cancel()
				cancelAllocator()
			}
		}
	})

	errRun := chromedp.Run(ctxNewContext,
		this.setCookie("ci_session", "r0sffmqur1rama56tlfp8hrkq2vlsrbm", "subhd.tv", "/", true, false),
		chromedp.Navigate(url),
		chromedp.WaitVisible(`#TencentCaptcha`),
	)
	if errRun != nil {
		if errRun.Error() != "context canceled" {
			return "", nil, errRun
		}
	}
	if fileUrl != "" {
		byteData, errHttpGet := common.HttpGet(fileUrl)
		return filepath.Base(fileUrl), byteData, errHttpGet
	}
	return "", nil, fmt.Errorf("auth error")
}

func (this *SubHD) GetArchiveFileList(filename string, archiveFile []byte) ([]string, error) {
	filenames := make([]string, 0)
	fileExt := filepath.Ext(filename)
	switch fileExt {
	case ".zip":
		r, err := zip.NewReader(bytes.NewReader(archiveFile), int64(len(archiveFile)))
		if err != nil {
			return nil, err
		}
		for i := 0; i < len(r.File); i++ {
			var content []byte
			if r.File[i].FileHeader.Flags>>11 != 1 {
				length := bytes.NewReader([]byte(r.File[i].Name))
				decoder := transform.NewReader(length, simplifiedchinese.GB18030.NewDecoder())
				content, err = ioutil.ReadAll(decoder)
				if err != nil {
					continue
				}
			} else {
				content = []byte(r.File[i].Name)
			}

			filenames = append(filenames, string(content))
		}
	case ".7z":
		r, err := go7z.NewReader(bytes.NewReader(archiveFile), int64(len(archiveFile)))
		if err != nil {
			return nil, err
		}
		for {
			hdr, errNext := r.Next()
			if errNext == io.EOF {
				break
			}
			if errNext != nil {
				continue
			}
			filenames = append(filenames, hdr.Name)
		}
	case ".rar":
		r, err := rardecode.NewReader(bytes.NewReader(archiveFile), "")
		if err != nil {
			return nil, err
		}
		for {
			hdr, errNext := r.Next()
			if errNext == io.EOF {
				break
			}
			if errNext != nil {
				continue
			}
			filenames = append(filenames, hdr.Name)
		}
	}
	return filenames, nil
}

func (this *SubHD) UnArchiveFile(archiveFilename string, archiveFile []byte, filename, toFilename string) error {
	fileExt := filepath.Ext(archiveFilename)

	toFilenameFileExt := filepath.Ext(toFilename)
	toFilename = toFilename[0 : len(toFilename)-len(toFilenameFileExt)]
	zipFileExt := filepath.Ext(filename)
	toFilename = fmt.Sprintf("%s%s", toFilename, zipFileExt)

	switch fileExt {
	case ".zip":
		r, err := zip.NewReader(bytes.NewReader(archiveFile), int64(len(archiveFile)))
		if err != nil {
			return err
		}
		var file *zip.File
		for i := 0; i < len(r.File); i++ {
			var content []byte
			if r.File[i].FileHeader.Flags>>11 != 1 {
				length := bytes.NewReader([]byte(r.File[i].Name))
				decoder := transform.NewReader(length, simplifiedchinese.GB18030.NewDecoder())
				content, err = ioutil.ReadAll(decoder)
				if err != nil {
					continue
				}
			} else {
				content = []byte(r.File[i].Name)
			}
			if string(content) == filename {
				file = r.File[i]
				break
			}
		}
		if file == nil {
			return fmt.Errorf("file not exist")
		}

		rc, errOpen := file.Open()
		if errOpen != nil {
			return errOpen
		}
		f, errOpenFile := os.OpenFile(toFilename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if errOpenFile != nil {
			return errOpenFile
		}
		defer func() {
			if errClose := f.Close(); errClose != nil {
				panic(errClose)
			}
		}()

		_, err = io.Copy(f, rc)
		if err != nil {
			return err
		}
		return nil
	case ".7z":
		r, err := go7z.NewReader(bytes.NewReader(archiveFile), int64(len(archiveFile)))
		if err != nil {
			return err
		}
		f, errOpenFile := os.OpenFile(toFilename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if errOpenFile != nil {
			return errOpenFile
		}
		defer func() {
			if errClose := f.Close(); errClose != nil {
				panic(errClose)
			}
		}()

		for {
			hdr, errNext := r.Next()
			if errNext == io.EOF {
				break
			}
			if errNext != nil {
				continue
			}
			if hdr.Name == filename {
				if _, errCopy := io.Copy(f, r); errCopy != nil {
					return errCopy
				}
				return nil
			}
		}
	case ".rar":
		r, err := rardecode.NewReader(bytes.NewReader(archiveFile), "")
		if err != nil {
			return err
		}
		f, errOpenFile := os.OpenFile(toFilename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if errOpenFile != nil {
			return errOpenFile
		}
		defer func() {
			if errClose := f.Close(); errClose != nil {
				panic(errClose)
			}
		}()

		for {
			hdr, errNext := r.Next()
			if errNext == io.EOF {
				break
			}
			if errNext != nil {
				continue
			}
			if hdr.Name == filename {
				if _, errCopy := io.Copy(f, r); errCopy != nil {
					return errCopy
				}
				return nil
			}
		}

	}

	return fmt.Errorf("file not exist")
}
