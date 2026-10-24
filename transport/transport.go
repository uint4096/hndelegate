package transport

type Post struct {
	Title         string
	ArticleURL    string
	DiscussionURL string
}

type Transport interface {
	Send(msg Post) error
	Alert(msg string)
}
