package ctrlsubsonic

import (
	"log"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/jinzhu/gorm"

	"senan.xyz/g/gonic/db"
	"senan.xyz/g/gonic/dir"
	"senan.xyz/g/gonic/server/ctrlsubsonic/params"
	"senan.xyz/g/gonic/server/ctrlsubsonic/spec"
	"senan.xyz/g/gonic/server/encode"
)

// "raw" handlers are ones that don't always return a spec response.
// it could be a file, stream, etc. so you must either
//   a) write to response writer
//   b) return a non-nil spec.Response
//  _but not both_

func (c *Controller) ServeGetCoverArt(w http.ResponseWriter, r *http.Request) *spec.Response {
	params := r.Context().Value(CtxParams).(params.Params)
	id, err := params.GetInt("id")
	if err != nil {
		return spec.NewError(10, "please provide an `id` parameter")
	}
	folder := &db.Album{}
	err = c.DB.
		Select("id, left_path, right_path, cover").
		First(folder, id).
		Error
	if gorm.IsRecordNotFoundError(err) {
		return spec.NewError(10, "could not find a cover with that id")
	}
	if folder.Cover == "" {
		return spec.NewError(10, "no cover found for that folder")
	}

	relPath := path.Join(
		folder.LeftPath,
		folder.RightPath,
		folder.Cover,
	)
	lastModified, readerSeeker, err := c.MusicDir.GetFile(relPath)
	if err != nil {
		return spec.NewError(11, "failed to get file: %v", err)
	}
	http.ServeContent(w, r, folder.Cover, lastModified, readerSeeker)
	if err := readerSeeker.Close(); err != nil {
		return spec.NewError(21, "failed to close input file: %v", err)
	}
	return nil
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

type serveTrackOptions struct {
	track      *db.Track
	pref       *db.TranscodePreference
	maxBitrate int
	cachePath  string
	musicDir   dir.Dir
}

func serveTrackRaw(w http.ResponseWriter, r *http.Request, opts serveTrackOptions) *spec.Response{
	log.Printf("serving raw %q\n", opts.track.Filename)
	w.Header().Set("Content-Type", opts.track.MIME())
	lastModified, readerSeeker, err := opts.musicDir.GetFile(opts.track.RelPath())
	if err != nil {
		return spec.NewError(11, "failed to get file: %v", err)
	}
	http.ServeContent(w, r, opts.track.Filename, lastModified, readerSeeker)
	if err := readerSeeker.Close(); err != nil {
		return spec.NewError(21, "Failed to close input data: %v", err)
	}
	return nil
}

func serveTrackEncode(w http.ResponseWriter, r *http.Request, opts serveTrackOptions) *spec.Response {
	profile := encode.Profiles[opts.pref.Profile]
	bitrate := encode.GetBitrate(opts.maxBitrate, profile)

	cacheKey := encode.CacheKey(opts.track.RelPath(), opts.pref.Profile, bitrate)
	cacheFile := path.Join(opts.cachePath, cacheKey)
	if fileExists(cacheFile) {
		log.Printf("serving transcode `%s`: cache [%s/%s] hit!\n", opts.track.Filename, profile.Format, bitrate)
		http.ServeFile(w, r, cacheFile)
		return nil
	}

	log.Printf("serving transcode `%s`: cache [%s/%s] miss!\n", opts.track.Filename, profile.Format, bitrate)
	_, originalFile, err := opts.musicDir.GetFile(opts.track.RelPath())
	if err != nil {
		return spec.NewError(11, "failed to read original file for encode: %v", err)
	}
	if err := encode.Encode(originalFile, w, cacheFile, profile, bitrate); err != nil {
		if err2 := originalFile.Close(); err2 != nil {
			return spec.NewError(121, "error encoding %v: %v\nEncountered error while closing input data: %v", opts.track.RelPath(), err, err2)
		} else {
			return spec.NewError(12, "error encoding %v: %v", opts.track.RelPath(), err)
		}
	}
	log.Printf("serving transcode `%s`: encoded to [%s/%s] successfully\n", opts.track.Filename, profile.Format, bitrate)
	if err := originalFile.Close(); err != nil {
		return spec.NewError(21, "Encountered error while closing input data: %v", err)
	}
	return nil
}

func (c *Controller) ServeStream(w http.ResponseWriter, r *http.Request) *spec.Response {
	params := r.Context().Value(CtxParams).(params.Params)
	id, err := params.GetInt("id")
	if err != nil {
		return spec.NewError(10, "please provide an `id` parameter")
	}
	track := &db.Track{}
	err = c.DB.
		Preload("Album").
		First(track, id).
		Error
	if gorm.IsRecordNotFoundError(err) {
		return spec.NewError(70, "media with id `%d` was not found", id)
	}
	user := r.Context().Value(CtxUser).(*db.User)
	defer func() {
		play := db.Play{
			AlbumID: track.Album.ID,
			UserID:  user.ID,
		}
		c.DB.
			Where(play).
			First(&play)
		play.Time = time.Now() // for getAlbumList?type=recent
		play.Count++           // for getAlbumList?type=frequent
		c.DB.Save(&play)
	}()
	client := params.Get("c")
	servOpts := serveTrackOptions{
		track:     track,
		musicDir:  c.MusicDir,
	}
	pref := &db.TranscodePreference{}
	err = c.DB.
		Where("user_id=?", user.ID).
		Where("client COLLATE NOCASE IN (?)", []string{"*", client}).
		Order("client DESC"). // ensure "*" is last if it's there
		First(pref).
		Error
	if gorm.IsRecordNotFoundError(err) {
		return serveTrackRaw(w, r, servOpts)
	} else if _, ok := encode.Profiles[pref.Profile]; !ok {
		return serveTrackRaw(w, r, servOpts)
	} else {
		servOpts.pref = pref
		servOpts.maxBitrate = params.GetIntOr("maxBitRate", 0)
		servOpts.cachePath = c.cachePath
		return serveTrackEncode(w, r, servOpts)
	}
}

func (c *Controller) ServeDownload(w http.ResponseWriter, r *http.Request) *spec.Response {
	params := r.Context().Value(CtxParams).(params.Params)
	id, err := params.GetInt("id")
	if err != nil {
		return spec.NewError(10, "please provide an `id` parameter")
	}
	track := &db.Track{}
	err = c.DB.
		Preload("Album").
		First(track, id).
		Error
	if gorm.IsRecordNotFoundError(err) {
		return spec.NewError(70, "media with id `%d` was not found", id)
	}
	return serveTrackRaw(w, r, serveTrackOptions{
		track:     track,
		musicDir:  c.MusicDir,
	})
}
