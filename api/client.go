package api

import (
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"errors"
	"encoding/json"
	"strconv"
)

const (
	apiBase   = "https://api.napster.com/"
	apiKey    = "ZTJlOWNhZGUtNzlmZS00ZGU2LTkwYjMtZDk1ODRlMDkwODM5"
	authToken = "WlRKbE9XTmhaR1V0TnpsbVpTMDBaR1UyTFRrd1lqTXRaRGsxT0RSbE1Ea3dPRE01Ok1UUmpaVFZqTTJFdE9HVmxaaTAwT1RVM0xXRm1Oamt0TlRsbE9ERmhObVl5TnpJNQ"
	userAgent     = "android/8.3.4.1091/NapsterGlobal"
)



func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add(
		"User-Agent", userAgent,
	)
	return http.DefaultTransport.RoundTrip(req)
}

func (c *Client) Auth(email, password string) error {
	data := url.Values{}
	data.Set("username", email)
	data.Set("password", password)
	data.Set("grant_type", "password")
	req, err := http.NewRequest(
		http.MethodPost, apiBase+"oauth/token", strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", authToken)
	do, err := c.Client.Do(req)
	if err != nil {
		return err
	}
	defer do.Body.Close()
	if do.StatusCode != http.StatusOK {
		return errors.New(do.Status)
	}
	var obj Auth
	err = json.NewDecoder(do.Body).Decode(&obj)
	if err != nil {
		return err
	}
	c.User.AccessToken = obj.AccessToken
	return nil
}

func (c *Client) GetUserInfo() error {
	req, err := http.NewRequest(http.MethodGet, apiBase+"v3/me", nil)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", "Bearer " + c.User.AccessToken)
	do, err := c.Client.Do(req)
	if err != nil {
		return err
	}
	defer do.Body.Close()
	if do.StatusCode != http.StatusOK {
		return errors.New(do.Status)
	}
	var obj UserInfo
	err = json.NewDecoder(do.Body).Decode(&obj)
	if err != nil {
		return err
	}
	c.User.SubName = obj.Subscription.ProductName
	c.User.Catalog = obj.Subscription.Catalog
	c.User.Country = obj.Subscription.Country
	c.User.Lang = obj.Lang
	return nil
}

func (c *Client) GetAlbumMeta(albumID string) (*Album, error) {
	req, err := http.NewRequest(http.MethodGet, apiBase+"v2.2/albums/"+albumID, nil)
	if err != nil {
		return nil, err
	}
	query := url.Values{}
	query.Set("catalog", c.User.Catalog)
	query.Set("lang", c.User.Lang)
	query.Set("rights", "2")
	req.URL.RawQuery = query.Encode()
	req.Header.Add("apikey", apiKey)
	do, err := c.Client.Do(req)
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
	if len(obj.Albums) < 1 {
		return nil, errors.New("The API didn't return any album metadata.")
	}
	return obj.Albums[0], nil
}

func (c *Client) GetAlbTracksMeta(albumID string) ([]*Track, error) {
	req, err := http.NewRequest(http.MethodGet, apiBase+"v2.2/albums/"+albumID+"/tracks", nil)
	if err != nil {
		return nil, err
	}
	query := url.Values{}
	query.Set("catalog", c.User.Catalog)
	query.Set("lang", c.User.Lang)
	query.Set("rights", "2")
	req.URL.RawQuery = query.Encode()
	req.Header.Add("apikey", apiKey)
	do, err := c.Client.Do(req)
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
	if len(obj.Tracks) < 1 {
		return nil, errors.New("The API didn't return any album tracks metadata.")
	}
	return obj.Tracks, nil
}

func (c *Client) GetStreamMeta(trackID string, format *Format) (*Stream, error) {
	req, err := http.NewRequest(http.MethodGet, apiBase+"v3/streams/tracks", nil)
	if err != nil {
		return nil, err
	}
	query := url.Values{}
	query.Set("bitDepth", strconv.Itoa(format.SampleBits))
	query.Set("bitrate", strconv.Itoa(format.Bitrate))
	query.Set("format", format.Name)
	query.Set("id", trackID)
	query.Set("sampleRate", strconv.Itoa(format.SampleRate))
	req.URL.RawQuery = query.Encode()
	req.Header.Add("Authorization", "Bearer " + c.User.AccessToken)
	do, err := c.Client.Do(req)
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
	if len(obj.Streams) < 1 {
		return nil, errors.New("The API didn't return any stream metadata.")
	}
	return obj.Streams[0], nil
}

func (c *Client) GetVideoMeta(videoID string) (*Video, error) {
	req, err := http.NewRequest(http.MethodGet, apiBase+"v3/videos", nil)
	if err != nil {
		return nil, err
	}
	query := url.Values{}
	query.Set("country", c.User.Country)
	query.Set("isStreamableOnly", "true")
	query.Set("partnerId", "614884be5a375ab1e17e8e03")
	query.Set("id", videoID)
	req.URL.RawQuery = query.Encode()
	req.Header.Add("Authorization", "Bearer " + c.User.AccessToken)
	do, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer do.Body.Close()
	if do.StatusCode != http.StatusOK {
		return nil, errors.New(do.Status)
	}
	var obj VideoMeta
	err = json.NewDecoder(do.Body).Decode(&obj)
	if err != nil {
		return nil, err
	}
	if len(obj.Videos) < 1 {
		return nil, errors.New("The API didn't return any video metadata.")
	}
	return obj.Videos[0], nil	
}

func (c *Client) GetVideoStreamMeta(videoID string) (*Stream, error) {
	req, err := http.NewRequest(http.MethodGet, apiBase+"v3/streams/videos", nil)
	if err != nil {
		return nil, err
	}
	query := url.Values{}
	query.Set("id", videoID)
	req.URL.RawQuery = query.Encode()
	req.Header.Add("Authorization", "Bearer " + c.User.AccessToken)
	do, err := c.Client.Do(req)
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
	if len(obj.Streams) < 1 {
		return nil, errors.New("The API didn't return any video stream metadata.")
	}
	return obj.Streams[0], nil
}
 
func NewClient(email, password string) (*Client, error) {
	jar, _ := cookiejar.New(nil)
	client := &Client{
		Client: &http.Client{Jar: jar, Transport: &Transport{}},
	}
	err := client.Auth(email, password)
	if err != nil {
		return nil, err
	}
	err = client.GetUserInfo()
	if err != nil {
		return nil, err
	}
	return client, nil
}
