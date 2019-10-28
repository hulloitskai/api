package spotify

import (
	"time"

	"github.com/zmb3/spotify"
	"go.stevenxie.me/api/v2/music"
)

func tracksFromSpotify(dst *[]music.Track, src []spotify.SimpleTrack) {
	if src == nil {
		*dst = nil
		return
	}
	*dst = make([]music.Track, len(src))
	for i, st := range src {
		trackFromSpotify(&(*dst)[i], &st)
	}
}

func trackFromSpotify(dst *music.Track, src *spotify.SimpleTrack) {
	if src == nil {
		return
	}
	dst.ID = src.ID.String()
	dst.Name = src.Name
	dst.ExternalURL = spotifyURL(src.ExternalURLs)
	dst.URI = string(src.URI)
	dst.Duration = time.Duration(src.Duration) * time.Millisecond
	artistsFromSpotify(&dst.Artists, src.Artists)
}

func albumsFromSpotify(dst *[]music.Album, src []spotify.SimpleAlbum) {
	if src == nil {
		*dst = nil
		return
	}
	*dst = make([]music.Album, len(src))
	for i, sa := range src {
		albumFromSpotify(&(*dst)[i], &sa)
	}
}

func albumFromSpotify(dst *music.Album, src *spotify.SimpleAlbum) {
	if src == nil {
		return
	}
	dst.ID = src.ID.String()
	dst.Name = src.Name
	dst.ExternalURL = spotifyURL(src.ExternalURLs)
	dst.URI = string(src.URI)
	imagesFromSpotify(&dst.Images, src.Images)
	artistsFromSpotify(&dst.Artists, src.Artists)
}

func artistsFromSpotify(dst *[]music.Artist, src []spotify.SimpleArtist) {
	if src == nil {
		*dst = nil
		return
	}
	*dst = make([]music.Artist, len(src))
	for i, sa := range src {
		artistFromSpotify(&(*dst)[i], &sa)
	}
}

func artistFromSpotify(dst *music.Artist, src *spotify.SimpleArtist) {
	if src == nil {
		return
	}
	dst.ID = src.ID.String()
	dst.Name = src.Name
	dst.ExternalURL = spotifyURL(src.ExternalURLs)
	dst.URI = string(src.URI)
}

func imagesFromSpotify(dst *[]music.Image, src []spotify.Image) {
	if src == nil {
		*dst = nil
		return
	}
	*dst = make([]music.Image, len(src))
	for i, img := range src {
		(*dst)[i] = music.Image(img)
	}
}

func spotifyURL(urls map[string]string) string {
	return urls["spotify"]
}
