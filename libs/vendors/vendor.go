package vendors

type SubtitleResult struct {
	Subtitles   []Subtitle
}

type Subtitle struct {
	Id          string
	Title       string
	Author      string
	PublishTime string
	Extension   []SubtitleExtension
}

type SubtitleExtension struct {
	Name  string
	Value string
}

type Vendor interface {
	Init() error
	Search(keyword string) (*SubtitleResult, error)
	DownloadFile(Id, filePath string) ([]byte, error)
	UnArchiveFile(archiveFilePath, filename, toFilename string) ([]byte, error)
}
