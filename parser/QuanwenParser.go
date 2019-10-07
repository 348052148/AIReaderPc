package parser

import (
	"net/http"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/text/transform"
	"golang.org/x/text/encoding/simplifiedchinese"
	"io"
	"test/gopanc/entitys"
	"time"
	//"test/gopanc/engine"
	"test/gopanc/msg"
	"fmt"
)

type QuanwenParser struct {
	linkSet *msg.LinkSet
}

func NewQuanwenParser() *QuanwenParser {
	return &QuanwenParser{}
}

func (parser *QuanwenParser)SetLinkSet(linkSet *msg.LinkSet)  {
	parser.linkSet = linkSet
}

func (parser *QuanwenParser)Request(url string) (io.ReadCloser, error) {
	client := http.Client{Timeout:time.Second * 5}
	request, _ := http.NewRequest("GET", url, nil)
	request.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/77.0.3865.90 Safari/537.36")
	request.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3")
	res, err :=client.Do(request)
	//res, err := client.Get(url)
	if err != nil {
		//fmt.Printf("status: time out %v", err)
		return nil, err
	}
	if res.StatusCode != 200 {
		return nil, err
	}
	return res.Body, nil
}

func (parser *QuanwenParser)ParserClassflysBooks(url string) (entitys.Classfly,error) {
	body, reqErr := parser.Request(url)
	//defer body.Close()
	if reqErr != nil {
		fmt.Println("ClassFly TIME OUT" + url)
		return entitys.Classfly{}, reqErr
	}
	//defer body.Close()
	bytes := transform.NewReader(body ,simplifiedchinese.GBK.NewDecoder())
	doc, err := goquery.NewDocumentFromReader(bytes)
	if err != nil {
		fmt.Println("ClassFly BAN OUT" + url)
		return entitys.Classfly{}, err
	}
	var bookDetailList []entitys.BookDetail
	// Find the review items
	doc.Find(".seeWell li").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the band and title
		title := s.Find("span>a").Eq(0).Text()
		href,_ := s.Find(".mr10").Attr("href")
		cover,_ := s.Find("a>img").Attr("src")
		//fmt.Printf("书名： %s , 封面: %s , 连接 %s\n",  title, cover, href)
		bookDetailList = append(bookDetailList, entitys.BookDetail{
			Title:title,
			Link:href,
			Cover:cover,
		})
	})
	return entitys.Classfly{
		Title:"",
		Books:bookDetailList,
		Cover: "",
	}, nil;
}


func (parser *QuanwenParser) ParserBookInfo(url string) (entitys.BookInfo, error) {
	body, reqErr := parser.Request(url)
	//defer body.Close()
	if reqErr != nil {
		fmt.Println("BOOK TIME OUT" + url)
		return entitys.BookInfo{}, reqErr
	}
	//defer body.Close()
	bytes := transform.NewReader(body ,simplifiedchinese.GBK.NewDecoder())
	doc, err := goquery.NewDocumentFromReader(bytes)
	if err != nil {
		fmt.Println("BOOK BAN TIME OUT" + url)
		return entitys.BookInfo{}, err
	}
	title := doc.Find(".b-info h1").Eq(0).Text()
	detail := doc.Find(".b-info .infoDetail #waa").Eq(0).Text()
	cover,_ := doc.Find(".detail .mr11>img").Eq(0).Attr("src")
	href,_ := doc.Find(".detail .mr11").Eq(0).Attr("href")
	var author,status string
	doc.Find(".author .bookDetail dl").Each(func(i int, s *goquery.Selection) {
		if i == 0 {
			status = s.Find("dd").Text()
		}
		if i == 1 {
			author = s.Find("dd").Text()
		}
	})
	//
	if title == "" || href == "" {
		fmt.Println("BOOK Parser Faill : " + url)
	}
	return entitys.BookInfo{
		BookId:entitys.Md5(title+author+status),
		Title:       title,
		Author:      author,
		Detail:      detail,
		Cover:       cover,
		Status:      status,
		ChapterLink: href,
	},nil
}

func (parser *QuanwenParser)ParserChapters(url string, bookId string) ([]entitys.Chapter, error) {
	body, reqErr := parser.Request(url)
	//defer body.Close()
	if reqErr != nil {
		fmt.Println("Chapter TIME OUT"+ url)
		return []entitys.Chapter{}, reqErr
	}
	//defer body.Close()
	bytes := transform.NewReader(body ,simplifiedchinese.GBK.NewDecoder())
	doc, err := goquery.NewDocumentFromReader(bytes)
	if err != nil {
		fmt.Println("Chapter BAN TIME OUT" + url)
		return []entitys.Chapter{}, err
	}
	var ChapterList []entitys.Chapter
	doc.Find(".chapterSo .chapterNum ul li>a").Each(func(i int, s *goquery.Selection) {
		link,_ := s.Attr("href")
		//fmt.Printf("title %s link : %s \n",s.Text(), link)
		ChapterList= append(ChapterList, entitys.Chapter{
			BookId:bookId,
			Title:s.Text(),
			Index:i,
			ContentLink:link,
		})
	})
	return ChapterList, nil
}
