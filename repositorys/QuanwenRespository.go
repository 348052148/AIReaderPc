package repositorys

import (
	"test/gopanc/entitys"
	"fmt"
	"os"
)

func SaveBooInfo(info entitys.BookInfo)  {
	file,err := os.OpenFile("data.sql", os.O_CREATE | os.O_APPEND, 0777)
	defer file.Close()
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(file, "insert into w_books (book_id,title,author, status, cover, detail, chapter_link, source) values ('%s','%s','%s','%s','%s','%s','%s', '%s'); \n",
		info.BookId, info.Title, info.Author, info.Status, info.Cover, info.Detail, info.ChapterLink, "全文网")
}

func SaveChapters(chapters []entitys.Chapter)  {
	file,err := os.OpenFile("data.sql", os.O_CREATE | os.O_APPEND, 0777)
	defer file.Close()
	if err != nil {
		panic(err)
	}
	for _,chapter := range chapters {
		fmt.Fprintf(file, "insert into w_book_chapters (title,`index`, content_link, book_id) values ('%s', %d, '%s', '%s'); \n",
			chapter.Title, chapter.Index, chapter.ContentLink, chapter.BookId)
	}
}

func SaveClassfly(classfly entitys.Classfly)  {
	fmt.Printf("Classfly %v", classfly)
}
