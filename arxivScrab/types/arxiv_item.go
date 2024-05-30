package types

type ArxivItem struct {
	ArxivLink string `json:"arxivLink" selector:"div.is-marginless > p.list-title > a"`
	Title string `json:"title" selector:"p.title"`
	Authors []string `json:"authors" selector:"p.authors a"`
	Abstract string `json:"abstract" selector:"p.abstract > span.abstract-full"`
	Date string `json:"date" selector:"p:nth-child(5)"`
	Comments string `json:"comments" selector:"p.comments span:nth-child(2)"`
}