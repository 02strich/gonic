package ctrlsubsonic

import (
	"log"
	"net/http"
	"strconv"
	"time"
	"unicode"

	"github.com/jinzhu/gorm"

	"senan.xyz/g/gonic/model"
	"senan.xyz/g/gonic/scanner"
	"senan.xyz/g/gonic/server/ctrlsubsonic/params"
	"senan.xyz/g/gonic/server/ctrlsubsonic/spec"
	"senan.xyz/g/gonic/server/funk"
	"senan.xyz/g/gonic/server/key"
	"senan.xyz/g/gonic/server/lastfm"
)

func lowerUDecOrHash(in string) string {
	lower := unicode.ToLower(rune(in[0]))
	if !unicode.IsLetter(lower) {
		return "#"
	}
	return string(lower)
}

type playlistOpValues struct {
	c    *Controller
	r    *http.Request
	user *model.User
	id   int
}

func playlistDelete(opts playlistOpValues) {
	indexes, ok := opts.r.URL.Query()["songIndexToRemove"]
	if !ok {
		return
	}
	trackIDs := []int{}
	opts.c.DB.
		Order("created_at").
		Model(&model.PlaylistItem{}).
		Where("playlist_id = ?", opts.id).
		Pluck("track_id", &trackIDs)
	for _, indexStr := range indexes {
		i, err := strconv.Atoi(indexStr)
		if err != nil {
			continue
		}
		opts.c.DB.Delete(&model.PlaylistItem{},
			"track_id = ?", trackIDs[i])
	}
}

func playlistAdd(opts playlistOpValues) {
	var toAdd []string
	for _, val := range []string{"songId", "songIdToAdd"} {
		var ok bool
		toAdd, ok = opts.r.URL.Query()[val]
		if ok {
			break
		}
	}
	for _, trackIDStr := range toAdd {
		trackID, err := strconv.Atoi(trackIDStr)
		if err != nil {
			continue
		}
		opts.c.DB.Save(&model.PlaylistItem{
			PlaylistID: opts.id,
			TrackID:    trackID,
		})
	}
}

func (c *Controller) ServeGetLicence(r *http.Request) *spec.Response {
	sub := spec.NewResponse()
	sub.Licence = &spec.Licence{
		Valid: true,
	}
	return sub
}

func (c *Controller) ServePing(r *http.Request) *spec.Response {
	return spec.NewResponse()
}

func (c *Controller) ServeScrobble(r *http.Request) *spec.Response {
	params := r.Context().Value(CtxParams).(params.Params)
	id, err := params.GetInt("id")
	if err != nil {
		return spec.NewError(10, "please provide an `id` parameter")
	}
	track := &model.Track{}
	c.DB.
		Preload("Album").
		Preload("Artist").
		First(track, id)
	user := r.Context().Value(key.User).(*model.User)
	// scrobble with above info
	err = lastfm.Scrobble(lastfm.ScrobbleOptions{
		BaseAuthOptions: lastfm.BaseAuthOptions{
			APIKey: c.DB.GetSetting("lastfm_api_key"),
			Secret: c.DB.GetSetting("lastfm_secret"),
		},
		Session: user.LastFMSession,
		Track:   track,
		// clients will provide time in miliseconds, so use that or
		// instead convert UnixNano to miliseconds
		StampMili:  params.GetIntOr("time", int(time.Now().UnixNano()/1e6)),
		Submission: params.GetOr("submission", "true") != "false",
	})
	if err != nil {
		log.Printf("error while submitting to lastfm: %v\n", err)
	}
	err = funk.Funk(funk.FunkOptions{
		BaseURL:  c.DB.GetSetting("funk_node"),
		Username: user.FunkPassword,
		Password: user.FunkPassword,
		Track:    track,
	})
	if err != nil {
		log.Printf("error while submitting to funk: %v\n", err)
	}
	return spec.NewResponse()
}

func (c *Controller) ServeGetMusicFolders(r *http.Request) *spec.Response {
	folders := &spec.MusicFolders{}
	folders.List = []*spec.MusicFolder{
		{ID: 1, Name: "music"},
	}
	sub := spec.NewResponse()
	sub.MusicFolders = folders
	return sub
}

func (c *Controller) ServeStartScan(r *http.Request) *spec.Response {
	go func() {
		if err := c.Scanner.Start(); err != nil {
			log.Printf("error while scanning: %v\n", err)
		}
	}()
	return c.ServeGetScanStatus(r)
}

func (c *Controller) ServeGetScanStatus(r *http.Request) *spec.Response {
	var trackCount int
	c.DB.
		Model(model.Track{}).
		Count(&trackCount)
	sub := spec.NewResponse()
	sub.ScanStatus = &spec.ScanStatus{
		Scanning: scanner.IsScanning(),
		Count:    trackCount,
	}
	return sub
}

func (c *Controller) ServeGetUser(r *http.Request) *spec.Response {
	user := r.Context().Value(CtxUser).(*model.User)
	sub := spec.NewResponse()
	sub.User = &spec.User{
		Username:          user.Name,
		AdminRole:         user.IsAdmin,
		ScrobblingEnabled: user.LastFMSession != "",
		Folder:            []int{1},
	}
	return sub
}

func (c *Controller) ServeNotFound(r *http.Request) *spec.Response {
	return spec.NewError(70, "view not found")
}

func (c *Controller) ServeGetPlaylists(r *http.Request) *spec.Response {
	user := r.Context().Value(CtxUser).(*model.User)
	var playlists []*model.Playlist
	c.DB.
		Where("user_id = ?", user.ID).
		Find(&playlists)
	sub := spec.NewResponse()
	sub.Playlists = &spec.Playlists{
		List: make([]*spec.Playlist, len(playlists)),
	}
	for i, playlist := range playlists {
		sub.Playlists.List[i] = spec.NewPlaylist(playlist)
		sub.Playlists.List[i].Owner = user.Name
	}
	return sub
}

func (c *Controller) ServeGetPlaylist(r *http.Request) *spec.Response {
	params := r.Context().Value(CtxParams).(params.Params)
	playlistID, err := params.GetInt("id")
	if err != nil {
		return spec.NewError(10, "please provide an `id` parameter")
	}
	playlist := model.Playlist{}
	err = c.DB.
		Where("id = ?", playlistID).
		Find(&playlist).
		Error
	if gorm.IsRecordNotFoundError(err) {
		return spec.NewError(70, "playlist with id `%d` not found", playlistID)
	}
	var tracks []*model.Track
	c.DB.
		Joins(`
            JOIN playlist_items
		    ON playlist_items.track_id = tracks.id
		`).
		Where("playlist_items.playlist_id = ?", playlistID).
		Group("tracks.id").
		Order("playlist_items.created_at").
		Preload("Album").
		Find(&tracks)
	user := r.Context().Value(CtxUser).(*model.User)
	sub := spec.NewResponse()
	sub.Playlist = spec.NewPlaylist(&playlist)
	sub.Playlist.Owner = user.Name
	sub.Playlist.List = make([]*spec.TrackChild, len(tracks))
	for i, track := range tracks {
		sub.Playlist.List[i] = spec.NewTCTrackByFolder(track, track.Album)
	}
	return sub
}

func (c *Controller) ServeUpdatePlaylist(r *http.Request) *spec.Response {
	params := r.Context().Value(CtxParams).(params.Params)
	user := r.Context().Value(CtxUser).(*model.User)
	var playlistID int
	for _, key := range []string{"id", "playlistId"} {
		if val, err := params.GetInt(key); err != nil {
			playlistID = val
		}
	}
	playlist := model.Playlist{ID: playlistID}
	c.DB.Where(playlist).First(&playlist)
	playlist.UserID = user.ID
	if val := r.URL.Query().Get("name"); val != "" {
		playlist.Name = val
	}
	if val := r.URL.Query().Get("comment"); val != "" {
		playlist.Comment = val
	}
	c.DB.Save(&playlist)
	opts := playlistOpValues{c, r, user, playlist.ID}
	playlistDelete(opts)
	playlistAdd(opts)
	return spec.NewResponse()
}

func (c *Controller) ServeDeletePlaylist(r *http.Request) *spec.Response {
	params := r.Context().Value(CtxParams).(params.Params)
	c.DB.
		Where("id = ?", params.GetIntOr("id", 0)).
		Delete(&model.Playlist{})
	return spec.NewResponse()
}
