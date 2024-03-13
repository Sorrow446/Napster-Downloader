package api

import "net/http"

type Client struct {
	Client  	*http.Client
	User         struct {
		Lang    	string
		Catalog 	string
		AccessToken string
		SubName     string	
		Country     string	
	}
}

type Transport struct{}

type Auth struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
}

type UserInfo struct {
	Lang                         string    `json:"lang"`
	Subscription                 struct {
		Catalog            string    `json:"catalog"`
		ProductName        string    `json:"productName"`
		Country            string    `json:"country"`
	} `json:"subscription"`
}

type Album struct {
	Type                string        `json:"type"`
	ID                  string        `json:"id"`
	UPC                 string        `json:"upc"`
	Shortcut            string        `json:"shortcut"`
	Href                string        `json:"href"`
	Name                string        `json:"name"`
	Released            string     `json:"released"`
	OriginallyReleased  string     `json:"originallyReleased"`
	Label               string        `json:"label"`
	Copyright           string        `json:"copyright"`
	Tags                []interface{} `json:"tags"`
	DiscCount           int           `json:"discCount"`
	TrackCount          int           `json:"trackCount"`
	IsExplicit          bool          `json:"isExplicit"`
	IsSingle            bool          `json:"isSingle"`
	AccountPartner      string        `json:"accountPartner"`
	ArtistName          string        `json:"artistName"`
	IsAvailableInHiRes  bool          `json:"isAvailableInHiRes"`
	IsAvailableInAtmos  bool          `json:"isAvailableInAtmos"`
	ContributingArtists struct {
		PrimaryArtist string `json:"primaryArtist"`
	} `json:"contributingArtists"`
	Discographies struct {
		Main []string `json:"main"`
	} `json:"discographies"`
	Links struct {
		Images struct {
			Href string `json:"href"`
		} `json:"images"`
		Tracks struct {
			Href string `json:"href"`
		} `json:"tracks"`
		Posts struct {
			Href string `json:"href"`
		} `json:"posts"`
		Artists struct {
			Ids  []string `json:"ids"`
			Href string   `json:"href"`
		} `json:"artists"`
		Genres struct {
			Ids  []string `json:"ids"`
			Href string   `json:"href"`
		} `json:"genres"`
		Reviews struct {
			Ids  []string `json:"ids"`
			Href string   `json:"href"`
		} `json:"reviews"`
	} `json:"links"`
	IsStreamable bool `json:"isStreamable"`
}

type AlbumMeta struct {
	Meta struct {
		ReturnedCount int         `json:"returnedCount"`
		TotalCount    interface{} `json:"totalCount"`
	} `json:"meta"`
	Albums []*Album `json:"albums"`
}

type Track struct {
	Type               string        `json:"type"`
	ID                 string        `json:"id"`
	Index              int           `json:"index"`
	Disc               int           `json:"disc"`
	Href               string        `json:"href"`
	PlaybackSeconds    int           `json:"playbackSeconds"`
	IsExplicit         bool          `json:"isExplicit"`
	IsStreamable       bool          `json:"isStreamable"`
	IsAvailableInHiRes bool          `json:"isAvailableInHiRes"`
	Name               string        `json:"name"`
	ISRC               string        `json:"isrc"`
	Shortcut           string        `json:"shortcut"`
	Blurbs             []interface{} `json:"blurbs"`
	ArtistID           string        `json:"artistId"`
	ArtistName         string        `json:"artistName"`
	AlbumName          string        `json:"albumName"`
	Formats            []*Format      `json:"formats"`
	LosslessFormats    []*Format      `json:"losslessFormats"`
	AlbumID            string        `json:"albumId"`
	IsAvailableInAtmos bool          `json:"isAvailableInAtmos"`
	Contributors       struct {
		PrimaryArtist string `json:"primaryArtist"`
	} `json:"contributors"`
	Links struct {
		Artists struct {
			Ids  []string `json:"ids"`
			Href string   `json:"href"`
		} `json:"artists"`
		Albums struct {
			Ids  []string `json:"ids"`
			Href string   `json:"href"`
		} `json:"albums"`
		Genres struct {
			Ids  []string `json:"ids"`
			Href string   `json:"href"`
		} `json:"genres"`
		Tags struct {
			Ids  []string `json:"ids"`
			Href string   `json:"href"`
		} `json:"tags"`
	} `json:"links"`
	PreviewURL string `json:"previewURL"`
}

type AlbumTracksMeta struct {
	Meta struct {
		ReturnedCount int         `json:"returnedCount"`
		TotalCount    interface{} `json:"totalCount"`
	} `json:"meta"`
	Tracks []*Track `json:"tracks"`
}

type Stream struct {
	Type       string `json:"type"`
	Format     *Format `json:"format"`
	PrimaryURL string `json:"primaryUrl"`	
}

type StreamMeta struct {
	Streams []*Stream `json:"streams"`
}

type Format struct {
	Bitrate    int    `json:"bitrate"`
	Name       string `json:"name"`
	SampleBits int    `json:"sampleBits"`
	SampleRate int    `json:"sampleRate"`
}

type ContributingArtist struct {
	Name string `json:"name"`
}

type Video struct {
	Name string `json:"name"`
	ContributingArtists []*ContributingArtist `json:"contributingArtists"`
}

type VideoMeta struct {
	Videos []*Video `json:"videos"`
}