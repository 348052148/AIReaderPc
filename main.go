package main

import (
	"test/gopanc/engine"
	"test/gopanc/parser"
	"flag"
)

func main() {
	spage := flag.Int("spage", 1, "开始页")
	epage := flag.Int("epage", 20, "结束页")
	flag.Parse()
	engine.NewWorkerEngine(parser.NewQuanwenParser()).Run1(*spage, *epage)
	//fmt.Println(ParserBookInfo(Request("http://www.quanshuwang.com/book_171170.html")))
	//fmt.Println(parser.ParserChapters(parser.Request("http://www.quanshuwang.com/book/171/171170")))
	//classflyList := ParserClassflys(Request("http://www.quanshuwang.com/list/1_1.html"))
	//for _,classfly :=range classflyList {
	//	Request(classfly.Href)
	//}
}
