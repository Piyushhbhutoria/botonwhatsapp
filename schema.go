package main

// SendText struct
type SendText struct {
	Receiver string `json:"to"`
	Message  string `json:"text"`
}

type sendBulkText struct {
	List []SendText `json:"list"`
}

type resp struct {
	Results results `json:"results"`
}

type results struct {
	Messages []messages `json:"messages"`
}

type messages struct {
	Content string `json:"content"`
}
