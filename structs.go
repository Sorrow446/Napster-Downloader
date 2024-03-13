package main

type WriteCounter struct {
	Total      int64
	TotalStr   string
	Downloaded int64
	Percentage int
	StartTime  int64
}

type Config struct {
	AlbumTemplate string
	Email         string
	Format        int
	OutPath       string
	Password      string
	TrackTemplate string
	KeepCover     bool
	Urls          []string
	UseFFmpegEnvVar bool
	FFmpegNameStr   string
}

type Args struct {
	Urls    []string `arg:"positional, required"`
	Format  int      `arg:"-f" default:"-1" help:"Download format.\n\t\t\t 1 = 64 Kbps AAC, 2 = 128 Kbps MP3, 3 = 192 Kbps AAC, 4 = 320 Kbps AAC, 5 = FLAC."`
	OutPath string   `arg:"-o" help:"Where to download to. Path will be made if it doesn't already exist."`
}