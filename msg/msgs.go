package msg

type LinkSet struct {
	BookListLinkChan chan string
	BookInfoLinkChan chan string
	ChapterLinkChan chan string
}

