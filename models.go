package main

type Group struct {
	ID        int    `db:"id_group" json:"id_group"`
	GroupName string `db:"groupName" json:"group"`
}

type Song struct {
	ID          int    `db:"id_song" json:"id_song"`
	GroupID     int    `db:"id_group" json:"id_group"`
	GroupName   string `db:"group" json:"group"`
	SongName    string `db:"song" json:"song"`
	ReleaseDate string `db:"release_date" json:"release_date,omitempty"`
	Lyrics      string `db:"lyrics" json:"text,omitempty"`
	Link        string `db:"link" json:"link,omitempty"`
}

type SongShort struct {
	GroupName string `db:"groupName" json:"group"`
	SongName  string `db:"songName" json:"song"`
}

type SongDetail struct {
	ReleaseDate string `db:"release_date" json:"release_date,omitempty"`
	Lyrics      string `db:"lyrics" json:"text,omitempty"`
	Link        string `db:"link" json:"link,omitempty"`
}
