package entitys

import (
	"crypto/md5"
	"fmt"
)

//书籍描述
type BookDetail struct {
	Title string
	Link string
	Cover string
}
//分类
type Classfly struct {
	Title string
	Books []BookDetail
	Cover string
}
//书籍信息
type BookInfo struct {
	BookId string
	Title string
	Author string
	Status string
	Cover string
	Detail string
	ChapterLink string
}

func Md5(str string) string {
	data := []byte(str)
	has := md5.Sum(data)
	md5str := fmt.Sprintf("%x", has)
	return md5str
}

//章节信息
type Chapter struct {
	Title string
	Index int
	ContentLink string
	BookId string
}