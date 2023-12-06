package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/Sorrow446/go-mp4tag"
	"github.com/alexflint/go-arg"
	"github.com/bogem/id3v2"
	"github.com/dustin/go-humanize"
	"github.com/go-flac/flacpicture"
	"github.com/go-flac/flacvorbis"
	"github.com/go-flac/go-flac"
)

const (
	apiBase   = "https://api.napster.com/"
	apiKey    = "ZTJlOWNhZGUtNzlmZS00ZGU2LTkwYjMtZDk1ODRlMDkwODM5"
	authToken = "Basic WlRKbE9XTmhaR1V0TnpsbVpTMDBaR1UyTFRrd1lqTXRaRGsxT0RS" +
		"bE1Ea3dPRE01Ok1UUmpaVFZqTTJFdE9HVmxaaTAwT1RVM0xXRm1Oamt0TlRsbE9ERmhObVl5TnpJNQ=="
	userAgent     = "android/8.0.1.1012/NapsterGlobal"
	regexString   = `^http(?:s|)://(?:play|web).napster.com/album/([aA]lb.\d+)(?:/|)$`
	trackTemplate = "{{.trackPad}}. {{.title}}"
	albumTemplate = "{{.albumArtist}} - {{.album}}"
	coverUrl      = "http://direct-ns.rhapsody.com/imageserver/v2/albums/%s/images/600x600.jpg"
)

var (
	jar, _ = cookiejar.New(nil)
	client = &http.Client{Transport: &Transport{}, Jar: jar}
)

func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add(
		"User-Agent", userAgent,
	)
	return http.DefaultTransport.RoundTrip(req)
}

func (wc *WriteCounter) Write(p []byte) (int, error) {
	var speed int64 = 0
	n := len(p)
	wc.Downloaded += int64(n)
	percentage := float64(wc.Downloaded) / float64(wc.Total) * float64(100)
	wc.Percentage = int(percentage)
	toDivideBy := time.Now().UnixMilli() - wc.StartTime
	if toDivideBy != 0 {
		speed = int64(wc.Downloaded) / toDivideBy * 1000
	}
	fmt.Printf("\r%d%% @ %s/s, %s/%s ", wc.Percentage, humanize.Bytes(uint64(speed)),
		humanize.Bytes(uint64(wc.Downloaded)), wc.TotalStr)
	return n, nil
}

func handleErr(errText string, err error, _panic bool) {
	errString := errText + "\n" + err.Error()
	if _panic {
		panic(errString)
	}
	fmt.Println(errString)
}

func wasRunFromSrc() bool {
	buildPath := filepath.Join(os.TempDir(), "go-build")
	return strings.HasPrefix(os.Args[0], buildPath)
}

func getScriptDir() (string, error) {
	var (
		ok    bool
		err   error
		fname string
	)
	runFromSrc := wasRunFromSrc()
	if runFromSrc {
		_, fname, _, ok = runtime.Caller(0)
		if !ok {
			return "", errors.New("Failed to get script filename.")
		}
	} else {
		fname, err = os.Executable()
		if err != nil {
			return "", err
		}
	}
	return filepath.Dir(fname), nil
}

func readTxtFile(path string) ([]string, error) {
	var lines []string
	f, err := os.OpenFile(path, os.O_RDONLY, 0755)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			lines = append(lines, line)
		}
	}
	if scanner.Err() != nil {
		return nil, scanner.Err()
	}
	return lines, nil
}

func contains(lines []string, value string) bool {
	for _, line := range lines {
		if strings.EqualFold(line, value) {
			return true
		}
	}
	return false
}

func processUrls(urls []string) ([]string, error) {
	var (
		processed []string
		txtPaths  []string
	)
	for _, _url := range urls {
		if strings.HasSuffix(_url, ".txt") && !contains(txtPaths, _url) {
			txtLines, err := readTxtFile(_url)
			if err != nil {
				return nil, err
			}
			for _, txtLine := range txtLines {
				if !contains(processed, txtLine) {
					processed = append(processed, txtLine)
				}
			}
			txtPaths = append(txtPaths, _url)
		} else {
			if !contains(processed, _url) {
				processed = append(processed, _url)
			}
		}
	}
	return processed, nil
}

func parseCfg() (*Config, error) {
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
	if cfg.OutPath == "" {
		cfg.OutPath = "Napster downloads"
	}
	cfg.Urls, err = processUrls(args.Urls)
	if err != nil {
		errString := "Failed to process URLs.\n" + err.Error()
		return nil, errors.New(errString)
	}
	return cfg, nil
}

func readConfig() (*Config, error) {
	data, err := ioutil.ReadFile("config.json")
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

func makeDirs(path string) error {
	return os.MkdirAll(path, 0755)
}

func fileExists(path string) (bool, error) {
	f, err := os.Stat(path)
	if err == nil {
		return !f.IsDir(), nil
	} else if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func checkUrl(_url string) string {
	regex := regexp.MustCompile(regexString)
	match := regex.FindStringSubmatch(_url)
	if match != nil {
		return match[1]
	}
	return ""
}

func auth(email, password string) (string, error) {
	data := url.Values{}
	data.Set("username", email)
	data.Set("password", password)
	data.Set("grant_type", "password")
	req, err := http.NewRequest(
		http.MethodPost, apiBase+"oauth/token", strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", authToken)
	do, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer do.Body.Close()
	if do.StatusCode != http.StatusOK {
		return "", errors.New(do.Status)
	}
	var obj Auth
	err = json.NewDecoder(do.Body).Decode(&obj)
	if err != nil {
		return "", err
	}
	return "Bearer " + obj.AccessToken, nil
}

func getUserInfo(token string) (*UserInfo, error) {
	req, err := http.NewRequest(http.MethodGet, apiBase+"v3/me", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", token)
	do, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer do.Body.Close()
	if do.StatusCode != http.StatusOK {
		return nil, errors.New(do.Status)
	}
	var obj UserInfo
	err = json.NewDecoder(do.Body).Decode(&obj)
	if err != nil {
		return nil, err
	}
	return &obj, nil
}

func getAlbumMeta(albumId, catalog, lang string) (*AlbumMeta, error) {
	req, err := http.NewRequest(http.MethodGet, apiBase+"v2.2/albums/"+albumId, nil)
	if err != nil {
		return nil, err
	}
	query := url.Values{}
	query.Set("catalog", catalog)
	query.Set("lang", lang)
	query.Set("rights", "2")
	req.URL.RawQuery = query.Encode()
	req.Header.Add("apikey", apiKey)
	do, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer do.Body.Close()
	if do.StatusCode != http.StatusOK {
		return nil, errors.New(do.Status)
	}
	var obj AlbumMeta
	err = json.NewDecoder(do.Body).Decode(&obj)
	if err != nil {
		return nil, err
	}
	return &obj, nil
}

func parseAlbumMeta(meta *Album) map[string]string {
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
		defTemplate = albumTemplate
	} else {
		defTemplate = trackTemplate
	}
	var buffer bytes.Buffer
	for {
		err := template.Must(template.New("").Parse(templateText)).Execute(&buffer, tags)
		if err == nil {
			break
		}
		fmt.Println("Failed to parse template. Default one will be used instead.")
		templateText = defTemplate
		buffer.Reset()
	}
	return html.UnescapeString(buffer.String())
}

func parseTrackMeta(meta *Track, albMeta map[string]string, trackNum, trackTotal int) map[string]string {
	albMeta["artist"] = meta.ArtistName
	albMeta["isrc"] = meta.ISRC
	albMeta["title"] = meta.Name
	albMeta["track"] = strconv.Itoa(trackNum)
	albMeta["trackPad"] = fmt.Sprintf("%02d", trackNum)
	albMeta["trackTotal"] = strconv.Itoa(trackTotal)
	return albMeta
}

func getAlbTrackskMeta(albumId, catalog, lang string) (*AlbumTracksMeta, error) {
	req, err := http.NewRequest(http.MethodGet, apiBase+"v2.2/albums/"+albumId+"/tracks", nil)
	if err != nil {
		return nil, err
	}
	query := url.Values{}
	query.Set("catalog", catalog)
	query.Set("lang", lang)
	query.Set("rights", "2")
	req.URL.RawQuery = query.Encode()
	req.Header.Add("apikey", apiKey)
	do, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer do.Body.Close()
	if do.StatusCode != http.StatusOK {
		return nil, errors.New(do.Status)
	}
	var obj AlbumTracksMeta
	err = json.NewDecoder(do.Body).Decode(&obj)
	if err != nil {
		return nil, err
	}
	return &obj, nil
}

func sanitise(filename string, isFolder bool) string {
	var regexStr string
	if isFolder {
		regexStr = `[:*?"><|]`
	} else {
		regexStr = `[\/:*?"><|]`
	}
	return regexp.MustCompile(regexStr).ReplaceAllString(filename, "_")
}

func getNextBestFmt(wantedFmt int, formats map[int]Format) Format {
	for i := wantedFmt - 1; i > 0; i-- {
		format, ok := formats[i]
		if ok {
			return format
		}
	}
	return Format{}
}

func selectFormat(wantedFmt int, track *Track) *Format {
	parsedFmts := map[int]Format{}
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
		fmt.Println("Unavailable in your chosen format. Will use next best.")
	}
	return &selectedFmt
}

func getStreamMeta(trackId, token string, format *Format) (*StreamMeta, error) {
	req, err := http.NewRequest(http.MethodGet, apiBase+"v3/streams/tracks", nil)
	if err != nil {
		return nil, err
	}
	query := url.Values{}
	query.Set("bitDepth", strconv.Itoa(format.SampleBits))
	query.Set("bitrate", strconv.Itoa(format.Bitrate))
	query.Set("format", format.Name)
	query.Set("id", trackId)
	query.Set("sampleRate", strconv.Itoa(format.SampleRate))
	req.URL.RawQuery = query.Encode()
	req.Header.Add("Authorization", token)
	do, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer do.Body.Close()
	if do.StatusCode != http.StatusOK {
		return nil, errors.New(do.Status)
	}
	var obj StreamMeta
	err = json.NewDecoder(do.Body).Decode(&obj)
	if err != nil {
		return nil, err
	}
	return &obj, nil
}

func formatFreq(freq int) string {
	freqStr := strconv.Itoa(freq / 100)
	if strings.HasSuffix(freqStr, "0") {
		return strconv.Itoa(freq / 1000)
	} else {
		freqStrLen := len(freqStr)
		return freqStr[:freqStrLen-1] + "." + freqStr[freqStrLen-1:]
	}
}

func parseSpecs(format *Format) (string, string) {
	var (
		specs string
		ext   string
	)
	switch formatName := format.Name; {
	case formatName == "AAC PLUS":
		specs = strconv.Itoa(format.Bitrate) + " Kbps AAC"
		ext = ".m4a"
	case formatName == "MP3":
		specs = strconv.Itoa(format.Bitrate) + " Kbps MP3"
		ext = ".mp3"
	case formatName == "AAC":
		specs = strconv.Itoa(format.Bitrate) + " Kbps AAC"
		ext = ".m4a"
	case formatName == "FLAC":
		specs = fmt.Sprintf("%d-bit / %s kHz FLAC", format.BitDepth, formatFreq(format.Bitrate))
		ext = ".flac"
	}
	return specs, ext
}

func downloadCover(albumId, path string) error {
	_url := fmt.Sprintf(coverUrl, albumId)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}
	defer f.Close()
	req, err := client.Get(_url)
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

func downloadTrack(trackPath, url string) error {
	f, err := os.OpenFile(trackPath, os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}
	defer f.Close()
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Range", "bytes=0-")
	do, err := client.Do(req)
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

func writeAacTags(trackPath string, tags map[string]string, coverBytes []byte) error {
	trackNum, err := strconv.Atoi(tags["track"])
	if err != nil {
		return err
	}
	trackTotal, err := strconv.Atoi(tags["trackTotal"])
	if err != nil {
		return err
	}
	_tags := &mp4tag.Tags{
		Album:       tags["album"],
		AlbumArtist: tags["albumArtist"],
		Copyright:   tags["copyright"],
		Custom: map[string]string{
			"ISRC": tags["isrc"],
			"UPC":  tags["upc"],
		},
		Label:       tags["label"],
		Title:       tags["title"],
		TrackNumber: trackNum,
		TrackTotal:  trackTotal,
		Year:        tags["year"],
	}
	if coverBytes != nil {
		_tags.Cover = coverBytes
	}
	return mp4tag.Write(trackPath, _tags)
}

func writeMp3Tags(trackPath string, tags map[string]string, coverBytes []byte) error {
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
	if coverBytes != nil {
		imgFrame := id3v2.PictureFrame{
			Encoding:    id3v2.EncodingUTF8,
			MimeType:    "image/jpeg",
			PictureType: id3v2.PTFrontCover,
			Picture:     coverBytes,
		}
		tag.AddAttachedPicture(imgFrame)
	}
	return tag.Save()
}

func writeFlacTags(trackPath string, tags map[string]string, coverBytes []byte) error {
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
	if coverBytes != nil {
		picture, err := flacpicture.NewFromImageData(
			flacpicture.PictureTypeFrontCover, "", coverBytes, "image/jpeg",
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
		coverBytes []byte
	)
	if coverPath != "" {
		coverBytes, err = ioutil.ReadFile(coverPath)
		if err != nil {
			return err
		}
	}
	delete(tags, "trackPad")
	switch {
	case ext == ".m4a":
		err = writeAacTags(trackPath, tags, coverBytes)
	case ext == ".mp3":
		err = writeMp3Tags(trackPath, tags, coverBytes)
	case ext == ".flac":
		err = writeFlacTags(trackPath, tags, coverBytes)
	}
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
	scriptDir, err := getScriptDir()
	if err != nil {
		panic(err)
	}
	err = os.Chdir(scriptDir)
	if err != nil {
		panic(err)
	}
	cfg, err := parseCfg()
	if err != nil {
		handleErr("Failed to parse config/args.", err, true)
	}
	err = makeDirs(cfg.OutPath)
	if err != nil {
		handleErr("Failed to make output folder.", err, true)
	}
	token, err := auth(cfg.Email, cfg.Password)
	if err != nil {
		handleErr("Failed to auth.", err, true)
	}
	userInfo, err := getUserInfo(token)
	if err != nil {
		handleErr("Failed to get subcription info.", err, true)
	}
	subName := userInfo.Subscription.ProductName
	if subName == "" {
		panic("Account subscription required.")
	}
	fmt.Println("Signed in successfully - " + subName + "\n")
	catalog := userInfo.Subscription.Catalog
	lang := userInfo.Lang
	albumTotal := len(cfg.Urls)
	for albumNum, _url := range cfg.Urls {
		fmt.Printf("Album %d of %d:\n", albumNum+1, albumTotal)
		albumId := checkUrl(_url)
		if albumId == "" {
			fmt.Println("Invalid URL:", _url)
			continue
		}
		_albumMeta, err := getAlbumMeta(albumId, catalog, lang)
		if err != nil {
			handleErr("Failed to get album metadata.", err, false)
			continue
		}
		if !(len(_albumMeta.Albums) > 0) {
			fmt.Println("The API didn't return album metadata.")
			continue
		}
		albumMeta := _albumMeta.Albums[0]
		parsedAlbMeta := parseAlbumMeta(&albumMeta)
		albumFolder := parseTemplate(cfg.AlbumTemplate, parsedAlbMeta, true)
		fmt.Println(parsedAlbMeta["albumArtist"] + " - " + parsedAlbMeta["album"])
		if len(albumFolder) > 120 {
			fmt.Println("Album folder was chopped as it exceeds 120 characters.")
			albumFolder = albumFolder[:120]
		}
		sanAlbumFolder := sanitise(albumFolder, true)
		albumPath := filepath.Join(cfg.OutPath, strings.TrimSuffix(sanAlbumFolder, "."))
		err = makeDirs(albumPath)
		if err != nil {
			handleErr("Failed to make album folder.", err, false)
			continue
		}
		albTracksMeta, err := getAlbTrackskMeta(albumId, catalog, lang)
		if err != nil {
			handleErr("Failed to get album tracks metadata.", err, false)
			continue
		}
		trackTotal := len(albTracksMeta.Tracks)
		if !(trackTotal > 0) {
			fmt.Println("The API didn't return album tracks metadata.")
			continue
		}
		coverPath := filepath.Join(albumPath, "cover.jpg")
		err = downloadCover(albumId, coverPath)
		if err != nil {
			handleErr("Failed to get cover.", err, false)
			coverPath = ""
		}
		for trackNum, track := range albTracksMeta.Tracks {
			trackNum++
			if !track.IsStreamable {
				fmt.Println("Track isn't streamable.")
				continue
			}
			parsedMeta := parseTrackMeta(&track, parsedAlbMeta, trackNum, trackTotal)
			selFormat := selectFormat(cfg.Format, &track)
			if selFormat == nil {
				fmt.Println("No format was selected.")
				continue
			}
			streamMeta, err := getStreamMeta(track.ID, token, selFormat)
			if err != nil {
				handleErr("Failed to get track stream metadata.", err, false)
				continue
			}
			if streamMeta.Meta.ReturnedCount == 0 {
				fmt.Println("The API didn't return the stream metadata.")
				continue
			}
			stream := streamMeta.Streams[0]
			trackFname := parseTemplate(cfg.TrackTemplate, parsedMeta, false)
			sanTrackFname := sanitise(trackFname, false)
			specs, ext := parseSpecs(&stream.Format)
			trackPath := filepath.Join(albumPath, sanTrackFname+ext)
			exists, err := fileExists(trackPath)
			if err != nil {
				handleErr("Failed to check if track already exists locally.", err, false)
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
			err = downloadTrack(trackPath, stream.PrimaryURL)
			if err != nil {
				handleErr("Failed to download track.", err, false)
				continue
			}
			err = writeTags(trackPath, coverPath, ext, parsedMeta)
			if err != nil {
				handleErr("Failed to write tags.", err, false)
			}
		}
		if coverPath != "" && !cfg.KeepCover {
			err := os.Remove(coverPath)
			if err != nil {
				handleErr("Failed to delete cover.", err, false)
			}
		}
	}
}
