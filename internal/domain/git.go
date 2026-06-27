package domain

import "time"

type TreeEntry struct {
	Name  string
	Path  string
	IsDir bool
	Size  int64
	Mode  string
}

type Commit struct {
	Hash      string
	ShortHash string
	Message   string
	Author    string
	Email     string
	When      time.Time
}

type FileBlob struct {
	Path     string
	Content  []byte
	IsBinary bool
	Size     int64
}

type FileDiff struct {
	Path    string
	Patch   string
	Added   int
	Deleted int
}
