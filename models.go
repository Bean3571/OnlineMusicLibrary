package main

type Song struct {
	//ID          int    `db:"id" json:"id"`
	Group       string `db:"group" json:"group"`
	Song        string `db:"song" json:"song"`
	ReleaseDate string `db:"release_date" json:"release_date,omitempty"`
	Text        string `db:"text" json:"text,omitempty"`
	Link        string `db:"link" json:"link,omitempty"`
}

type SongShort struct {
	Group string `db:"group" json:"group"`
	Song  string `db:"song" json:"song"`
}

type SongDetail struct {
	ReleaseDate string `db:"release_date" json:"release_date,omitempty"`
	Text        string `db:"text" json:"text,omitempty"`
	Link        string `db:"link" json:"link,omitempty"`
}
