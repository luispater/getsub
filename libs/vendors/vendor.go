package vendors

type SubtitleResult struct {
	Subtitles []Subtitle
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
	DownloadFile(id string) (string, []byte, error)
	GetArchiveFileList(filename string, archiveFile []byte) ([]string, error)
	UnArchiveFile(archiveFilename string, archiveFile []byte, filename, toFilename string) error
}
