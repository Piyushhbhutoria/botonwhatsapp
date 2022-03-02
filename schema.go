package main

type resp struct {
	Results results `json:"results"`
}

type results struct {
	Messages []messages `json:"messages"`
}

type messages struct {
	Content string `json:"content"`
}
