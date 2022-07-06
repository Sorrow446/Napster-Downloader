package main

import "time"

type Transport struct{}

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
}

type Args struct {
	Urls    []string `arg:"positional, required"`
	Format  int      `arg:"-f" default:"-1" help:"Download format.\n\t\t\t 1 = 64 Kbps AAC, 2 = 128 Kbps MP3, 3 = 192 Kbps AAC, 4 = 320 Kbps AAC, 5 = FLAC."`
	OutPath string   `arg:"-o" help:"Where to download to. Path will be made if it doesn't already exist."`
}

type Auth struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	ExpiresIn    int         `json:"expires_in"`
	TokenType    string      `json:"token_type"`
	Username     string      `json:"username"`
	FirstName    string      `json:"first_name"`
	LastName     string      `json:"last_name"`
	GUID         string      `json:"guid"`
	Catalog      string      `json:"catalog"`
	AuthCodeID   interface{} `json:"authCodeId"`
}

type UserInfo struct {
	Type                         string    `json:"type"`
	Href                         string    `json:"href"`
	ID                           string    `json:"id"`
	Created                      time.Time `json:"created"`
	OriginCode                   string    `json:"originCode"`
	Lang                         string    `json:"lang"`
	UserName                     string    `json:"userName"`
	ScreenName                   string    `json:"screenName"`
	DmcaRadioUser                bool      `json:"dmcaRadioUser"`
	ParentalControlEnabled       bool      `json:"parentalControlEnabled"`
	IsCurrentSubscriptionPayable bool      `json:"isCurrentSubscriptionPayable"`
	ScreenNameAutoCreated        bool      `json:"screenNameAutoCreated"`
	Subscription                 struct {
		IsTrial      bool   `json:"isTrial"`
		IsSuspended  bool   `json:"isSuspended"`
		TierCode     string `json:"tierCode"`
		TierName     string `json:"tierName"`
		Entitlements struct {
			CanStreamOnWeb                         bool   `json:"canStreamOnWeb"`
			CanStreamOnMobile                      bool   `json:"canStreamOnMobile"`
			CanStreamOnHomeDevice                  bool   `json:"canStreamOnHomeDevice"`
			CanStreamOnPC                          bool   `json:"canStreamOnPC"`
			CanUpgradeStreams                      bool   `json:"canUpgradeStreams"`
			MaxStreamCount                         int    `json:"maxStreamCount"`
			IsPlayBasedTier                        bool   `json:"isPlayBasedTier"`
			PlayBasedTierPeriod                    string `json:"playBasedTierPeriod"`
			CanPlayPremiumRadio                    bool   `json:"canPlayPremiumRadio"`
			CanStreamOffline                       bool   `json:"canStreamOffline"`
			MaxDeviceCount                         int    `json:"maxDeviceCount"`
			CanStreamRadio                         bool   `json:"canStreamRadio"`
			CanStreamOnDemand                      bool   `json:"canStreamOnDemand"`
			CanStreamHiRes                         bool   `json:"canStreamHiRes"`
			MaxStreamBitrate                       int    `json:"maxStreamBitrate"`
			MaxDownloadBitrate                     int    `json:"maxDownloadBitrate"`
			MaxPcCount                             int    `json:"maxPcCount"`
			HasStreamEntitlements                  bool   `json:"hasStreamEntitlements"`
			MaxBitDepth                            int    `json:"maxBitDepth"`
			MaxSampleRate                          int    `json:"maxSampleRate"`
			CanStream360Video                      bool   `json:"canStream360Video"`
			CanStreamMusicVideo                    bool   `json:"canStreamMusicVideo"`
			CanStreamTvShow                        bool   `json:"canStreamTvShow"`
			CanStreamAcousticVideo                 bool   `json:"canStreamAcousticVideo"`
			CanStreamBehindTheScenesVideo          bool   `json:"canStreamBehindTheScenesVideo"`
			CanStreamConcertVideo                  bool   `json:"canStreamConcertVideo"`
			CanStreamContentProviderOriginalsVideo bool   `json:"canStreamContentProviderOriginalsVideo"`
			CanStreamDocumentaryVideo              bool   `json:"canStreamDocumentaryVideo"`
			CanStreamFeatureVideo                  bool   `json:"canStreamFeatureVideo"`
			CanStreamInterviewVideo                bool   `json:"canStreamInterviewVideo"`
			CanStreamMakingOfVideo                 bool   `json:"canStreamMakingOfVideo"`
			CanStreamMarketingVideo                bool   `json:"canStreamMarketingVideo"`
			CanStreamNapsterExclusiveVideo         bool   `json:"canStreamNapsterExclusiveVideo"`
			CanStreamRawFootageVideo               bool   `json:"canStreamRawFootageVideo"`
			CanStreamSpecialEventVideo             bool   `json:"canStreamSpecialEventVideo"`
			CanStreamTrailerVideo                  bool   `json:"canStreamTrailerVideo"`
			CanStreamVrVideo                       bool   `json:"canStreamVrVideo"`
			CanStreamWebisodeVideo                 bool   `json:"canStreamWebisodeVideo"`
		} `json:"entitlements"`
		BillingPartnerCode string    `json:"billingPartnerCode"`
		Catalog            string    `json:"catalog"`
		CreateDate         time.Time `json:"createDate"`
		ProductCode        string    `json:"productCode"`
		ProductName        string    `json:"productName"`
		ProductServiceTerm string    `json:"productServiceTerm"`
		TrialLengthDays    int       `json:"trialLengthDays"`
		Country            string    `json:"country"`
		PartnerID          string    `json:"partnerId"`
		State              string    `json:"state"`
	} `json:"subscription"`
	Profile struct {
		IsPublic       bool   `json:"isPublic"`
		Bio            string `json:"bio"`
		Gender         string `json:"gender"`
		Role           string `json:"role"`
		FollowerCount  int    `json:"followerCount"`
		FollowingCount int    `json:"followingCount"`
		Avatar         struct {
			URL           string `json:"url"`
			DefaultAvatar bool   `json:"defaultAvatar"`
		} `json:"avatar"`
		Links struct {
			Followers string `json:"followers"`
			Following string `json:"following"`
		} `json:"links"`
		NickName string `json:"nickName"`
	} `json:"profile"`
	MiddleName string `json:"middleName"`
}

type Album struct {
	Type                string        `json:"type"`
	ID                  string        `json:"id"`
	UPC                 string        `json:"upc"`
	Shortcut            string        `json:"shortcut"`
	Href                string        `json:"href"`
	Name                string        `json:"name"`
	Released            time.Time     `json:"released"`
	OriginallyReleased  time.Time     `json:"originallyReleased"`
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
	Albums []Album `json:"albums"`
}

type Format struct {
	Type       string `json:"type"`
	Bitrate    int    `json:"bitrate"`
	Name       string `json:"name"`
	SampleBits int    `json:"sampleBits"`
	BitDepth   int    `json:"bitDepth"`
	SampleRate int    `json:"sampleRate"`
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
	Formats            []Format      `json:"formats"`
	LosslessFormats    []Format      `json:"losslessFormats"`
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
	Tracks []Track `json:"tracks"`
}

type StreamMeta struct {
	Meta struct {
		Query struct {
		} `json:"query"`
		ReturnedCount int `json:"returnedCount"`
		TotalCount    int `json:"totalCount"`
	} `json:"meta"`
	Streams []struct {
		Type       string `json:"type"`
		Format     Format `json:"format"`
		PrimaryURL string `json:"primaryUrl"`
	} `json:"streams"`
}
