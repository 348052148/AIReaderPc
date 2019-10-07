package engine

import (
	"test/gopanc/repositorys"
	"test/gopanc/msg"
	"test/gopanc/parser"
	"test/gopanc/worker"
	"fmt"
)

type PipelineEngine struct {
	LinkSet *msg.LinkSet
	Parser *parser.QuanwenParser
	Workers *worker.Worker
}


func NewPipelineEngine(parser *parser.QuanwenParser) *PipelineEngine  {
	linkSet := &msg.LinkSet{
		BookListLinkChan:make(chan string),
		BookInfoLinkChan:make(chan string),
		ChapterLinkChan:make(chan string),
	}
	parser.SetLinkSet(linkSet)
	engine := &PipelineEngine{
		LinkSet:linkSet,
		Parser:parser,
		Workers:worker.NewWorker(),
	}
	//parser.SetEngine(engine)
	return engine
}

//解析分类下书籍列表
func (e *PipelineEngine)ParserClassflyBoosWorker(url string) {
	classfly,_ := e.Parser.ParserClassflysBooks(url)
	for _, book := range classfly.Books {
		e.LinkSet.BookInfoLinkChan <- book.Link
	}
	//不应该去关闭通道
	//defer close( e.LinkSet.BookInfoLinkChan)
}

//解析书籍信息
func (e *PipelineEngine)ParserBookInfoWorker() {
	//处理channel\
	for  link := range  e.LinkSet.BookInfoLinkChan{
		bookInfo,err := e.Parser.ParserBookInfo(link)
		if err != nil {
			continue
		}
		fmt.Println("BOOK SUCESS:" + link)
		//repositorys.SaveBooInfo(bookInfo)
		e.LinkSet.ChapterLinkChan<- bookInfo.ChapterLink
		//time.Sleep(time.Millisecond * 500)
	}
	defer close( e.LinkSet.ChapterLinkChan)
}

//解析章节目录
func (e *PipelineEngine)ParserChaptersWorker()  {
	var index = 0;
	for link := range e.LinkSet.ChapterLinkChan {
		chapters, err := e.Parser.ParserChapters(link, "")
		if err != nil {
			continue
		}
		fmt.Println("Chapter SUCESS:" + link)
		repositorys.SaveChapters(chapters)
		index++
	}
}

//channel流水线版本
func (e *PipelineEngine)Run()  {
	e.Workers.Run()
	e.Workers.AddTask(&worker.Task{Func: func(task *worker.Task) {
		e.ParserClassflyBoosWorker("http://www.quanshuwang.com/list/1_1.html")
		defer fmt.Println("----------------------Task Classfly End")
	}})
	e.Workers.AddTask(&worker.Task{Func: func(task *worker.Task) {
		e.ParserBookInfoWorker()
		defer fmt.Println("----------------------Task BookInfo End")
	}})
	e.Workers.AddTask(&worker.Task{Func: func(task *worker.Task) {
		e.ParserChaptersWorker()
		defer fmt.Println("----------------------Task Chapters End")
	}})
	e.Workers.WaitProcess()
	fmt.Println("------------------WAIT ENDS")
	e.Workers.Shutdown()

}