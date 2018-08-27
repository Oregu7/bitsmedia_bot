package helpers

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func getDocument(url string) *goquery.Document {
	// Request the HTML page.
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}
	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	return doc
}

// ParseUpdates парсит ссылки на статьи
func ParseUpdates(url string) []string {
	// Request the HTML page.
	doc := getDocument(url)
	updates := doc.Find("div.news-item .news-top > a:last-child").Map(func(_ int, s *goquery.Selection) string {
		href, _ := s.Attr("href")
		return "https://bits.media" + href
	})

	return updates
}

// ScrapArticle парсит статью
func ScrapArticle(url string) Article {
	// Request the HTML page.
	doc := getDocument(url)
	title := strings.TrimSpace(doc.Find(".article-top h1").Text())
	image, existImage := doc.Find("img.article-picture").Attr("src")
	if existImage {
		image = "https://bits.media" + image
	}
	// tags
	regxTagReplace := regexp.MustCompile(`[\s-:']+`)
	tags := doc.Find(".article-tags > a").Map(func(_ int, s *goquery.Selection) string {
		return regxTagReplace.ReplaceAllString(s.Text(), "_")
	})
	// body
	text := parseContent(doc.Find(".article-page").Contents().Slice(5, -12))
	return Article{title, image, text, tags}
}

func parseContent(block *goquery.Selection) string {
	text := ""
	block.Each(func(_ int, s *goquery.Selection) {
		frameURL, existFrame := s.Find("iframe").Attr("src")
		imageURL, existImage := s.Find("img").Attr("src")

		if existImage {
			image := "https://bits.media" + imageURL
			text += fmt.Sprintf("<figure><img src='%s'></figure>", image)
		} else if existFrame {
			text += createVideoFrame(frameURL)
		} else if goquery.NodeName(s) == "#text" {
			if len(s.Text()) > 0 {
				text += strings.TrimSpace(s.Text()) + " "
			}
		} else if goquery.NodeName(s) == "h2" {
			text += "<h3>" + strings.TrimSpace(s.Text()) + "</h3>"
		} else if goquery.NodeName(s) == "blockquote" {
			text += "<blockquote>" + strings.TrimSpace(s.Text()) + "</blockquote>"
		} else if goquery.NodeName(s) == "a" {
			href, _ := s.Attr("href")
			text += fmt.Sprintf("<a href='%s'>%s</a>", href, s.Text())
		} else if s.HasClass("twitter-tweet") {
			twitter, _ := s.Find("a").Last().Attr("href")
			text += fmt.Sprintf("<figure><iframe src='/embed/twitter?url=%s'></iframe></figure>", twitter)
		} else {
			nodeName := goquery.NodeName(s)
			ret, _ := s.Html()
			items := strings.Split(strings.TrimSpace(ret), "\n")

			if len(items) > 1 {
				ret = ""
				for _, item := range items {
					ret += strings.TrimSpace(item) + " "
				}
			}

			text += fmt.Sprintf("<%s>%s</%s>", nodeName, strings.TrimSpace(ret), nodeName)
		}
	})

	return text
}

func createVideoFrame(url string) string {
	youtubeRegx := regexp.MustCompile(`https://www.youtube.com/embed/([-\w]+)/?.*`)
	serviceName := "youtube"

	isYoutube, _ := regexp.MatchString("youtube", url)
	if isYoutube {
		match := youtubeRegx.FindStringSubmatch(url)
		shortURL := "https://www.youtube.com/watch?v=" + match[1]
		return fmt.Sprintf("<figure><iframe src='/embed/%s?url=%s'></iframe></figure>", serviceName, shortURL)
	}

	return fmt.Sprintf("<a href = '%s'>%s</a></br>", url, url)

}
