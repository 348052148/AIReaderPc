package engine

import (
	"test/gopanc/parser"

	"test/gopanc/worker"
	"test/gopanc/msg"
	"fmt"
	//"test/gopanc/repositorys"
	//"test/gopanc/repositorys"
	"sync"
	"test/gopanc/repositorys"
)

type WorkerEngine struct {
	Parser *parser.QuanwenParser
	Workers *worker.Worker
}

func NewWorkerEngine(parser *parser.QuanwenParser) *WorkerEngine  {
	linkSet := &msg.LinkSet{
		BookListLinkChan:make(chan string),
		BookInfoLinkChan:make(chan string),
		ChapterLinkChan:make(chan string),
	}
	parser.SetLinkSet(linkSet)
	engine := &WorkerEngine{
		Parser:parser,
		Workers:worker.NewWorker(),
	}
	//parser.SetEngine(engine)
	return engine
}

func fanIn(done chan interface{}, inStreams ...<-chan interface{}) <-chan interface{} {
	outStream := make(chan interface{})
	wg := &sync.WaitGroup{}
	for _,inStream := range inStreams {
		wg.Add(1)
		go func(in <- chan interface{}) {
			defer wg.Done()
			for v := range in{
				select {
				case outStream <- v:
				case <-done:
					return
				}
			}
		}(inStream)
	}
	//
	go func() {
		wg.Wait()
		fmt.Println("close ALL")
		close(outStream)
	}()
	return outStream
}

func (e *WorkerEngine)Run1(spage, epage int)  {
	var resultChanList []<-chan interface{}
	for page := spage; page <= epage; page++ {
		resultChan := make(chan interface{})
		parame := make(map[string]interface{})
		parame["page"] = page
		syncTask := &worker.Task{Func: func(task *worker.Task) {
			page := task.Parame["page"].(int)
			fmt.Printf("Page Get Page %d \n", page)
			url := fmt.Sprintf("http://www.quanshuwang.com/list/1_%d.html", page)
			classfly, err := e.Parser.ParserClassflysBooks(url)
			if err != nil {
				task.ReTry()
				return
			}
			fmt.Println("ClassFly :" + url)
			//var arr []interface{}
			for i, book := range classfly.Books {
				task.Result.(chan interface{}) <- struct {
					Index int
					Link  string
				}{
					i, book.Link,
				}
			}
			close(task.Result.(chan interface{}))
			fmt.Printf("close:  %s \n", page)
		}, Result: resultChan, Parame:parame}
		//存放入ARRAYe
		resultChanList = append(resultChanList, resultChan)
		e.Workers.AddTask(syncTask)
	}

	resultChapterLinkChan := make(chan struct {
		BookId string
		ChapterLink string
	})
	go func() {
		wg := &sync.WaitGroup{}
		for v := range fanIn(nil, resultChanList...) {
			parame := make(map[string]interface{})
			parame["p"] = v
			wg.Add(1)
			e.Workers.AddTask(&worker.Task{Func: func(task *worker.Task) {
				parame := task.Parame["p"].(struct {
					Index int
					Link  string
				})
				bookInfo, err := e.Parser.ParserBookInfo(parame.Link)
				if err != nil {
					task.ReTry()
					return
				}
				//如果解析失败- 失败原因。网页无内容返回
				if bookInfo.Title == "" {
					wg.Done()
					return
				}
				resultChapterLinkChan <- struct {
					BookId string
					ChapterLink string
				}{
					BookId:bookInfo.BookId,
					ChapterLink:bookInfo.ChapterLink,
				}
				repositorys.SaveBooInfo(bookInfo)
				//fmt.Printf("BOOKINFOLink %d %s \n", parame.Index, parame.Link)
				wg.Done()
			},Parame:parame})

		}
		go func() {
			wg.Wait()
			fmt.Println("Close ChapterLinkChan")
			defer close(resultChapterLinkChan)
		}()
	}()
	go func() {
		for v := range resultChapterLinkChan {
			parame := make(map[string]interface{})
			parame["p"] = v
			e.Workers.AddTask(&worker.Task{Func: func(task *worker.Task) {
				link :=task.Parame["p"].(struct{
					BookId string
					ChapterLink string
				}).ChapterLink
				bookId :=task.Parame["p"].(struct{
					BookId string
					ChapterLink string
				}).BookId
				chapters, err := e.Parser.ParserChapters(link, bookId)
				if err != nil {
					task.ReTry()
					return
				}
				repositorys.SaveChapters(chapters)
				//fmt.Println("ChapterLink " + link)
			},Parame:parame})
		}
	}()
	e.Workers.Run()
	e.Workers.WaitProcess();
}

func (e *WorkerEngine)Run()  {
	//同步任务
	var resultChanList []<-chan interface{}
	for page := 1; page <= 1; page++ {
		resultChan := make(chan interface{})
		parame := make(map[string]interface{})
		parame["page"] = page
		syncTask := &worker.Task{Func: func(task *worker.Task) {
			page := task.Parame["page"].(int)
			fmt.Printf("Page Get Page %d \n", page)
			url := fmt.Sprintf("http://www.quanshuwang.com/list/1_%d.html", page)
			classfly, err := e.Parser.ParserClassflysBooks(url)
			if err != nil {
				task.ReTry()
				return
			}
			fmt.Println("ClassFly :" + url)
			//var arr []interface{}
			for i, book := range classfly.Books {
				task.Result.(chan interface{}) <- struct {
					Index int
					Link  string
				}{
					i, book.Link,
				}
			}
			close(task.Result.(chan interface{}))
		}, Result: resultChan, Parame:parame}
		//存放入ARRAY
		resultChanList = append(resultChanList, resultChan)
		e.Workers.AddTask(syncTask)
	}

	resultChanIn := fanIn(nil, resultChanList...)
	//Channel

	resTask := &worker.Task{Func: func(task *worker.Task) {
		for v := range resultChanIn {
			fmt.Println(v)
			p := make(map[string]interface{})
			p["link"] = v
			e.Workers.AddSyncTask(&worker.Task{Func: func(task *worker.Task) {
				parame := task.Parame["link"].(struct {
					Index int
					Link  string
				})
				bookInfo, err := e.Parser.ParserBookInfo(parame.Link)
				if err != nil {
					task.ReTry()
					return
				}
				repositorys.SaveBooInfo(bookInfo)
				fmt.Printf("BOOKINFOLink %d %s \n", parame.Index, parame.Link)
			},Parame:p})
		}
	}}
	e.Workers.AddTask(resTask)

	fmt.Println("Wait resTask")

	//解析Chapter
	//parame := make(map[string]interface{})
	//parame["link"] = bookInfo.ChapterLink
	//e.Workers.AddTask(&worker.Task{Func: func(task *worker.Task) {
	//	chapters, err := e.Parser.ParserChapters(task.Parame["link"].(string))
	//	if err != nil {
	//		e.Workers.AddTask(task)
	//		return
	//	}
	//	repositorys.SaveChapters(chapters)
	//	fmt.Println("ChapterLink " + task.Parame["link"].(string))
	//
	//}, Parame: parame})

	e.Workers.Run()
	//fmt.Println("ADDCLSSFLY")
	//for v := range syncTask.Result {
	//	parame := make(map[string]interface{})
	//	parame["link"] = v.(struct {
	//		Index int
	//		Link  string
	//	}).Link
	//	parame["index"] = v.(struct {
	//		Index int
	//		Link  string
	//	}).Index
	//	e.Workers.AddTask(&worker.Task{Func: func(task *worker.Task) {
	//		bookInfo, err := e.Parser.ParserBookInfo(task.Parame["link"].(string))
	//		if err != nil {
	//			e.Workers.AddTask(task)
	//			return
	//		}
	//		repositorys.SaveBooInfo(bookInfo)
	//		fmt.Printf("BOOKINFOLink %d %s \n", task.Parame["index"], task.Parame["link"].(string))
	//		//解析Chapter
	//		parame := make(map[string]interface{})
	//		parame["link"] = bookInfo.ChapterLink
	//		e.Workers.AddTask(&worker.Task{Func: func(task *worker.Task) {
	//			chapters, err := e.Parser.ParserChapters(task.Parame["link"].(string))
	//			if err != nil {
	//				e.Workers.AddTask(task)
	//				return
	//			}
	//			repositorys.SaveChapters(chapters)
	//			fmt.Println("ChapterLink " + task.Parame["link"].(string))
	//
	//		}, Parame: parame})
	//	},
	//		Parame: parame,
	//	})
	//}
	fmt.Println("--------------------WAIT----------------------")
	e.Workers.WaitProcess()
}
