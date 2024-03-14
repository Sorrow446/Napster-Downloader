package main

import (
	"bytes"
	"encoding/json"
	"sort"
	"errors"
	"fmt"
	"html"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
	"main/api"
	"main/utils"
	"os/exec"
	"github.com/Sorrow446/go-mp4tag"
	"github.com/alexflint/go-arg"
	"github.com/bogem/id3v2"
	"github.com/dustin/go-humanize"
	"github.com/go-flac/flacpicture"
	"github.com/go-flac/flacvorbis"
	"github.com/go-flac/go-flac"
	"github.com/grafov/m3u8"
)

const (
	defTrackTemplate = "{{.trackPad}}. {{.title}}"
	defAlbumTemplate = "{{.albumArtist}} - {{.album}}"
	coverURL      = "http://direct-ns.rhapsody.com/imageserver/v2/albums/%s/images/600x600.jpg"
)

var regexStrings = [2]string{
	`^http(?:s|)://(?:play|web).napster.com/album/([aA]lb.\d+)(?:/|)$`,
	`^http(?:s|)://(?:play|web).napster.com/video/([a-z\d]+)(?:/|)$`,
}


func (wc *WriteCounter) Write(p []byte) (int, error) {
	var speed int64 = 0
	n := len(p)
	wc.Downloaded += int64(n)
	percentage := float64(wc.Downloaded) / float64(wc.Total) * float64(100)
	wc.Percentage = int(percentage)
	toDivideBy := time.Now().UnixMilli() - wc.StartTime
	if toDivideBy > 0 {
		speed = int64(wc.Downloaded) / toDivideBy * 1000
	}
	fmt.Printf("\r%d%% @ %s/s, %s/%s ", wc.Percentage, humanize.Bytes(uint64(speed)),
		humanize.Bytes(uint64(wc.Downloaded)), wc.TotalStr)
	return n, nil
}

func handleErr(err error, shouldPanic bool) {
	// pc := make([]uintptr, 1)
	// n := runtime.Callers(2, pc)
	// frames := runtime.CallersFrames(pc[:n])
	// frame, _ := frames.Next()
	// e := fmt.Sprintf(
	// 	"An error occured on line %d in function %s in file %s.\n%s",
	// 	frame.Line, frame.Function, frame.File, err.Error())
	// fmt.Println(e)
	if shouldPanic {
		panic(err)
	} else {
		fmt.Println(err.Error())
	}
}

func parseConfig() (*Config, error) {
	cfg, err := readConfig()
	if err != nil {
		return nil, err
	}
	args := parseArgs()
	if args.Format != -1 {
		cfg.Format = args.Format
	}
	if !(cfg.Format >= 1 && cfg.Format <= 5) {
		return nil, errors.New("Format must be between 1 and 5.")
	}
	if args.OutPath != "" {
		cfg.OutPath = args.OutPath
	}
	if strings.TrimSpace(cfg.OutPath) == "" {
		cfg.OutPath = "Napster downloads"
	}

	if strings.TrimSpace(cfg.AlbumTemplate) == "" {
		cfg.AlbumTemplate = defAlbumTemplate
	}

	if strings.TrimSpace(cfg.TrackTemplate) == "" {
		cfg.TrackTemplate = defTrackTemplate
	}

	if cfg.UseFFmpegEnvVar {
		cfg.FFmpegNameStr = "ffmpeg"
	} else {
		cfg.FFmpegNameStr = "./ffmpeg"
	}

	cfg.Urls, err = utils.ProcessUrls(args.Urls)
	if err != nil {
		fmt.Println("Failed to process URLs.")
		return nil, err
	}
	return cfg, nil
}

func readConfig() (*Config, error) {
	data, err := os.ReadFile("config.json")
	if err != nil {
		return nil, err
	}
	var obj Config
	err = json.Unmarshal(data, &obj)
	if err != nil {
		return nil, err
	}
	return &obj, nil
}

func parseArgs() *Args {
	var args Args
	arg.MustParse(&args)
	return &args
}

func checkUrl(_url string) (string, int) {
	splitUrl := strings.SplitN(_url, "?", 2)
	if len(splitUrl) > 1 {
		_url = splitUrl[0]
	}
	for idx, regexString := range regexStrings {
		regex := regexp.MustCompile(regexString)
		match := regex.FindStringSubmatch(_url)
		if match != nil {
			return match[1], idx
		}
	}
	return "", -1
}

func parseAlbumMeta(meta *api.Album) map[string]string {
	released := meta.OriginallyReleased
	if released != "" {
		released = released[:4]
	}

	parsedMeta := map[string]string{
		"album":       meta.Name,
		"albumArtist": meta.ArtistName,
		"label":       meta.Label,
		"copyright":   meta.Copyright,
		"upc":         meta.UPC,
		"year":        released,
	}
	return parsedMeta
}

func parseTemplate(templateText string, tags map[string]string, isAlbum bool) string {
	var defTemplate string
	if isAlbum {
		defTemplate = defAlbumTemplate
	} else {
		defTemplate = defTrackTemplate
	}
	var buffer bytes.Buffer
	for {
		err := template.Must(
			template.New("").Parse(templateText)).Execute(&buffer, tags)
		if err == nil {
			break
		}
		fmt.Println(
			"Failed to parse template; default one will be used instead.")
		templateText = defTemplate
		buffer.Reset()
	}
	return html.UnescapeString(buffer.String())
}

func parseTrackMeta(meta *api.Track, albMeta map[string]string, trackNum, trackTotal int) map[string]string {
	albMeta["artist"] = meta.ArtistName
	albMeta["isrc"] = meta.ISRC
	albMeta["title"] = meta.Name
	albMeta["track"] = strconv.Itoa(trackNum)
	albMeta["trackPad"] = fmt.Sprintf("%02d", trackNum)
	albMeta["trackTotal"] = strconv.Itoa(trackTotal)
	return albMeta
}

func getNextBestFmt(wantedFmt int, formats map[int]*api.Format) *api.Format {
	for i := wantedFmt - 1; i > 0; i-- {
		format, ok := formats[i]
		if ok {
			return format
		}
	}
	return nil
}

func selectFormat(wantedFmt int, track *api.Track) *api.Format {
	parsedFmts := map[int]*api.Format{}
	if len(track.LosslessFormats) > 0 {
		parsedFmts[5] = track.LosslessFormats[0]
	}
	for _, format := range track.Formats {
		switch formatName := format.Name; {
		case formatName == "AAC PLUS":
			parsedFmts[1] = format
		case formatName == "MP3":
			// Assuming always 128 Kbps.
			parsedFmts[2] = format
		case formatName == "AAC":
			bitrate := format.Bitrate
			if bitrate == 192 {
				parsedFmts[3] = format
			} else if bitrate == 320 {
				parsedFmts[4] = format
			}
		}
	}
	selectedFmt, ok := parsedFmts[wantedFmt]
	if !ok {
		selectedFmt = getNextBestFmt(wantedFmt, parsedFmts)
		fmt.Println("Unavailable in your chosen format; will use next best.")
	}
	return selectedFmt
}

func parseSpecs(format *api.Format) (string, string) {
	var (
		specs string
		ext   string
	)

	bitrateStr := strconv.Itoa(format.Bitrate)
	switch format.Name {
	case "AAC", "AAC PLUS":
		specs = bitrateStr + " Kbps AAC"
		ext = ".m4a"
	case "MP3":
		specs =  bitrateStr + " Kbps MP3"
		ext = ".mp3"
	case "FLAC":
		specs = fmt.Sprintf(
			"%d-bit / %s kHz FLAC", format.SampleBits, utils.FormatFreq(format.Bitrate))
		ext = ".flac"
	}
	return specs, ext
}

func downloadCover(client *api.Client, albumID, path string) error {
	_url := fmt.Sprintf(coverURL, albumID)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0777)
	if err != nil {
		return err
	}
	defer f.Close()
	req, err := client.Client.Get(_url)
	if err != nil {
		return err
	}
	defer req.Body.Close()
	if req.StatusCode != http.StatusOK {
		return errors.New(req.Status)
	}
	_, err = io.Copy(f, req.Body)
	return err
}

func downloadTrack(client *api.Client, trackPath, url string) error {
	f, err := os.OpenFile(trackPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0777)
	if err != nil {
		return err
	}
	defer f.Close()
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Range", "bytes=0-")
	do, err := client.Client.Do(req)
	if err != nil {
		return err
	}
	defer do.Body.Close()
	if do.StatusCode != http.StatusOK && do.StatusCode != http.StatusPartialContent {
		return errors.New(do.Status)
	}
	totalBytes := do.ContentLength
	counter := &WriteCounter{
		Total:     totalBytes,
		TotalStr:  humanize.Bytes(uint64(totalBytes)),
		StartTime: time.Now().UnixMilli(),
	}
	_, err = io.Copy(f, io.TeeReader(do.Body, counter))
	fmt.Println("")
	return err
}

func writeMP4Tags(trackPath string, tags map[string]string, coverData []byte) error {
	trackNum, err := strconv.Atoi(tags["track"])
	if err != nil {
		return err
	}
	trackTotal, err := strconv.Atoi(tags["trackTotal"])
	if err != nil {
		return err
	}

	year, err := strconv.ParseInt(tags["year"], 10, 32)
	if err != nil {
		return err
	}

	t := &mp4tag.MP4Tags{
		Album:       tags["album"],
		AlbumArtist: tags["albumArtist"],
		Copyright:   tags["copyright"],
		Custom: map[string]string{
			"ISRC": tags["isrc"],
			"UPC":  tags["upc"],
			"LABEL": tags["label"],
		},
		Title:       tags["title"],
		TrackNumber: int16(trackNum),
		TrackTotal:  int16(trackTotal),
		Year:        int32(year),
	}
	if coverData != nil {
		t.Pictures = []*mp4tag.MP4Picture{&mp4tag.MP4Picture{Data: coverData}}
	}

	mp4, err := mp4tag.Open(trackPath)
	if err != nil {
		panic(err)
	}
	defer mp4.Close()

	return mp4.Write(t, []string{})
}

func writeMP3Tags(trackPath string, tags map[string]string, coverData []byte) error {
	tags["track"] += "/" + tags["trackTotal"]
	resolve := map[string]string{
		"album":       "TALB",
		"artist":      "TPE1",
		"albumArtist": "TPE2",
		"copyright":   "TCOP",
		"isrc":        "TSRC",
		"label":       "TPUB",
		"title":       "TIT2",
		"track":       "TRCK",
		"year":        "TYER",
	}
	tag, err := id3v2.Open(trackPath, id3v2.Options{Parse: true})
	if err != nil {
		return err
	}
	defer tag.Close()
	for k, v := range tags {
		resolved, ok := resolve[k]
		if ok {
			tag.AddTextFrame(resolved, tag.DefaultEncoding(), v)
		}
	}
	if coverData != nil {
		imgFrame := id3v2.PictureFrame{
			Encoding:    id3v2.EncodingUTF8,
			MimeType:    "image/jpeg",
			PictureType: id3v2.PTFrontCover,
			Picture:     coverData,
		}
		tag.AddAttachedPicture(imgFrame)
	}
	return tag.Save()
}

func writeFLACTags(trackPath string, tags map[string]string, coverData []byte) error {
	f, err := flac.ParseFile(trackPath)
	if err != nil {
		return err
	}
	tag := flacvorbis.New()
	for k, v := range tags {
		tag.Add(strings.ToUpper(k), v)
	}
	tagMeta := tag.Marshal()
	f.Meta = append(f.Meta, &tagMeta)
	if coverData != nil {
		picture, err := flacpicture.NewFromImageData(
			flacpicture.PictureTypeFrontCover, "", coverData, "image/jpeg",
		)
		if err != nil {
			return err
		}
		pictureMeta := picture.Marshal()
		f.Meta = append(f.Meta, &pictureMeta)
	}
	return f.Save(trackPath)
}

func writeTags(trackPath, coverPath, ext string, tags map[string]string) error {
	var (
		err        error
		coverData []byte
	)
	if coverPath != "" {
		coverData, err = os.ReadFile(coverPath)
		if err != nil {
			return err
		}
	}
	delete(tags, "trackPad")
	switch {
	case ext == ".m4a":
		err = writeMP4Tags(trackPath, tags, coverData)
	case ext == ".mp3":
		err = writeMP3Tags(trackPath, tags, coverData)
	case ext == ".flac":
		err = writeFLACTags(trackPath, tags, coverData)
	}
	return err
}

func processAlbum(client *api.Client, id string, config *Config) error {
	albumMeta, err := client.GetAlbumMeta(id)
	if err != nil {
		fmt.Println("Failed to get album metadata.")
		return err
	}
	parsedAlbMeta := parseAlbumMeta(albumMeta)
	albumFolder := parseTemplate(config.AlbumTemplate, parsedAlbMeta, true)
	fmt.Println(parsedAlbMeta["albumArtist"] + " - " + parsedAlbMeta["album"])
	if len(albumFolder) > 140 {
		fmt.Println("Album folder was chopped as it exceeds 140 characters.")
		albumFolder = albumFolder[:140]
	}
	sanAlbumFolder := utils.Sanitise(albumFolder, true)
	albumPath := filepath.Join(config.OutPath, strings.TrimSuffix(sanAlbumFolder, "."))
	err = utils.MakeDirs(albumPath)
	if err != nil {
		return err
	}
	albTracksMeta, err := client.GetAlbTracksMeta(id)
	if err != nil {
		return err	
	}
	trackTotal := len(albTracksMeta)

	coverPath := filepath.Join(albumPath, "cover.jpg")
	err = downloadCover(client, id, coverPath)
	if err != nil {
		handleErr(err, false)
		coverPath = ""
	}
	for trackNum, track := range albTracksMeta {
		trackNum++
		if !track.IsStreamable {
			fmt.Println("Track isn't streamable.")
			continue
		}
		parsedMeta := parseTrackMeta(track, parsedAlbMeta, trackNum, trackTotal)
		selFormat := selectFormat(config.Format, track)
		if selFormat == nil {
			fmt.Println("No format was selected.")
			continue
		}
		streamMeta, err := client.GetStreamMeta(track.ID, selFormat)
		if err != nil {
			handleErr(err, false)
			continue
		}

		trackFname := parseTemplate(config.TrackTemplate, parsedMeta, false)
		sanTrackFname := utils.Sanitise(trackFname, false)
		specs, ext := parseSpecs(selFormat)
		trackPathNoExt := filepath.Join(albumPath, sanTrackFname)
		trackPathIncomp := trackPathNoExt + ".incomplete"
		trackPath := trackPathNoExt+ext

		exists, err := utils.FileExists(trackPath)
		if err != nil {
			handleErr(err, false)
			continue
		}
		if exists {
			fmt.Println("Track already exists locally.")
			continue
		}
		fmt.Printf(
			"Downloading track %d of %d: %s - %s\n", trackNum, trackTotal, parsedMeta["title"],
			specs,
		)
		err = downloadTrack(client, trackPathIncomp, streamMeta.PrimaryURL)
		if err != nil {
			handleErr(err, false)
			continue
		}
		err = writeTags(trackPathIncomp, coverPath, ext, parsedMeta)
		if err != nil {
			handleErr(err, false)
			continue
		}

		err = os.Rename(trackPathIncomp, trackPath)
		if err != nil {
			handleErr(err, false)
		}
	}
	if coverPath != "" && !config.KeepCover {
		err := os.Remove(coverPath)
		if err != nil {
			handleErr(err, false)
		}
	}
	return nil
}

func getPlaylist(client *api.Client, _url string) (m3u8.Playlist, error) {
	resp, err := client.Client.Get(_url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}
	playlist, _, err := m3u8.DecodeFrom(resp.Body, true)
	if err != nil {
		return nil, err
	}
	return playlist, nil
}

func getMasterPlaylist(client *api.Client, _url string) (*m3u8.MasterPlaylist, error) {
	playlist, err := getPlaylist(client, _url)
	if err != nil {
		return nil, err
	}
	return playlist.(*m3u8.MasterPlaylist), nil
}

func getMediaPlaylist(client *api.Client, _url string) (*m3u8.MediaPlaylist, error) {
	playlist, err := getPlaylist(client, _url)
	if err != nil {
		return nil, err
	}
	return playlist.(*m3u8.MediaPlaylist), nil
}

func selectMasterVariant(master *m3u8.MasterPlaylist) *m3u8.Variant {
	sort.Slice(master.Variants, func(x, y int) bool {
		return master.Variants[x].Bandwidth > master.Variants[y].Bandwidth
	})
	return master.Variants[0]
}

func formatResolution(res string) string {
	splitRes := strings.SplitN(res, "x", 2)
	if len(splitRes) < 2 {
		return "?"
	}
	yRes := splitRes[1]
	if yRes == "2160" {
		return "4K"
	} else {
		return yRes + "p" 
	}
}

func getPlistBaseURL(_url string) string {
	lastIdx := strings.LastIndex(_url, "/")
	return _url[:lastIdx+1]
}

func getSegPaths(media *m3u8.MediaPlaylist) []string {
	var segPaths []string
	for _, seg := range media.Segments {
		if seg == nil {
			break
		}
		segPaths = append(segPaths, seg.URI)
	}
	return segPaths
}

func downloadSegs(client *api.Client, segPaths []string, baseURL, outPath string) error {
	f, err := os.OpenFile(outPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0777)
	if err != nil {
		return err
	}
	defer f.Close()

	segTotal := len(segPaths)
	for segNum, segPath := range segPaths {
		segNum++
		fmt.Printf("\rSegment %d of %d.", segNum, segTotal)
		if err != nil {
			return err
		}
		do, err := client.Client.Get(baseURL+segPath)
		if err != nil {
			return err
		}
		if do.StatusCode != http.StatusOK {
			do.Body.Close()
			return errors.New(do.Status)
		}
		_, err = io.Copy(f, do.Body)
		do.Body.Close()
		if err != nil {
			return err
		}
	}
	fmt.Println("")
	return nil
}

func tsToMP4(ffmpegNameStr, inPath, outPath string) error {
	var errBuffer bytes.Buffer
	cmd := exec.Command(
		ffmpegNameStr, "-hide_banner", "-loglevel", "fatal", "-i", inPath,
		"-c", "copy", outPath)
	cmd.Stderr = &errBuffer
	err := cmd.Run()
	if err != nil {
		errString := fmt.Sprintf("%s\n%s", err, errBuffer.String())
		return errors.New(errString)
	}
	return nil
}

func processVideo(client *api.Client, id string, config *Config) error {
	videoMeta, err := client.GetVideoMeta(id)
	if err != nil {
		return err
	}

	outFname := videoMeta.ContributingArtists[0].Name + " - " + videoMeta.Name
	fmt.Println(outFname)
	streamMeta, err := client.GetVideoStreamMeta(id)
	if err != nil {
		return err
	}

	masterPlistURL := streamMeta.PrimaryURL
	master, err := getMasterPlaylist(client, masterPlistURL)
	if err != nil {
		return err
	}

	selVariant := selectMasterVariant(master)
	fmtRes := formatResolution(selVariant.Resolution)
	fmt.Printf("%.3f FPS, ~%d kbps, %s (%s)\n",
		selVariant.FrameRate, selVariant.Bandwidth/1000, fmtRes, selVariant.Resolution)	
	
	sanVideoFname := utils.Sanitise(outFname, false)
	videoPathNoExt := filepath.Join(config.OutPath, sanVideoFname)
	videoPathIncomp := videoPathNoExt + ".incomplete.ts"
	videoPath := videoPathNoExt+".mp4"

	exists, err := utils.FileExists(videoPath)	
	if err != nil {
		return err
	}

	if exists {
		fmt.Println("Video already exists locally.")
		return nil
	}

	plistBaseURL := getPlistBaseURL(masterPlistURL)
	media, err := getMediaPlaylist(client, plistBaseURL + selVariant.URI)
	if err != nil {
		return err
	}
	segPaths := getSegPaths(media)
	err = downloadSegs(client, segPaths, plistBaseURL, videoPathIncomp)
	if err != nil {
		return err
	}
	err = tsToMP4(config.FFmpegNameStr, videoPathIncomp, videoPath)
	if err != nil {
		return err
	}
	err = os.Remove(videoPathIncomp)
	return err
}

func init() {
	fmt.Println(`
 _____             _              ____                _           _         
|   | |___ ___ ___| |_ ___ ___   |    \ ___ _ _ _ ___| |___ ___ _| |___ ___ 
| | | | .'| . |_ -|  _| -_|  _|  |  |  | . | | | |   | | . | .'| . | -_|  _|
|_|___|__,|  _|___|_| |___|_|    |____/|___|_____|_|_|_|___|__,|___|___|_|  
          |_|
	`)
}

func main() {
	err := utils.CDToScriptDir()
	if err != nil {
		panic(err)
	}
	config, err := parseConfig()
	if err != nil {
		handleErr(err, true)
	}
	err = utils.MakeDirs(config.OutPath)
	if err != nil {
		handleErr(err, true)
	}
	client, err := api.NewClient(config.Email, config.Password)
	if err != nil {
		handleErr(err, true)
	}
	subName := client.User.SubName
	if subName == "" {
		panic("Account subscription required.")
	}
	fmt.Println("Signed in successfully - " + subName + "\n")
	urlTotal := len(config.Urls)
	for urlNum, _url := range config.Urls {
		fmt.Printf("URL %d of %d:\n", urlNum+1, urlTotal)
		id, urlType := checkUrl(_url)
		if id == "" {
			fmt.Println("Invalid URL:", _url)
			continue
		}
		switch urlType {
		case 0:
			err = processAlbum(client, id, config)
		case 1:
			err = processVideo(client, id, config)
		}
		if err != nil {
			handleErr(err, false)
		}
	}
}
