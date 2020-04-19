package ctrlupnp

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"html"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"senan.xyz/g/gonic/db"
	"strconv"
	"strings"
)

func (c *Controller) ServeCMSEvent(r *http.Request) *Response {
	if r.Method != "SUBSCRIBE" {
		return &Response{code: http.StatusNotImplemented, err: "Unsupported method"}
	}

	log.Printf("[UPNP] Event msg received:\nPath: %s\n Callback: %s\nHeader: %s\n---\n",
		r.URL.Path, r.Header.Get("Callback"), r.Header)

	// TODO: Implement subscriptions
	// sid, _ := uuid.NewRandom()
	// w.Header().Set("SID", "uuid:"+sid.String())
	// w.Header().Set("Timeout", "Second-1800")
	// w.Header().Set("Server", "unix/5.1 UPnP/1.0 YodlCastServer/1.0")
	// w.Header().Set("transferMode.dlna.org", "Streaming")
	// w.Header().Set("contentFeatures.dlna.org", "DLNA.ORG_OP=01;DLNA.ORG_CI=0;DLNA.ORG_FLAGS=01700000000000000000000000000000")
	return &Response{code: http.StatusOK}
}

func (c *Controller) ServeCMSControl(r *http.Request) *Response {
	action := strings.ToLower(
		regexp.MustCompile("\".*#(.+)\"").FindStringSubmatch(
			r.Header.Get("SOAPACTION"))[1])
	log.Printf("[CMS] Action received: %s\n", action)

	switch action {
	case "getprotocolinfo":
		return &Response{code: http.StatusOK, responseData: []byte(
	`<s:Envelope xmlns="urn:schemas-upnp-org:service-1-0" xmlns:s="http://schemas.xmlsoap.org/soap/envelope/" s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/">
		<s:Body>
			<u:GetProtocolInfoResponse xmlns:u="urn:schemas-upnp-org:service:ConnectionManager:1">
				<Source>http</Source>
				<Sink>http</Sink>
			</u:GetProtocolInfoResponse>
		</s:Body>
	</s:Envelope>`)}
	default:
		return &Response{code: http.StatusNotImplemented, err: fmt.Sprintf("[CMS] Unknown action '%s'\n", action)}
	}
}

func (c *Controller) ServeCDSEvents(r *http.Request) *Response {
	if r.Method != "SUBSCRIBE" {
		return &Response{code: http.StatusNotImplemented, err: "Unsupported method"}
	}

	log.Printf("Event msg received: %s\n", r.Header.Get("Callback"))
	// Let's just reject subscription requests...
	return &Response{code: http.StatusInternalServerError}
	// TODO: Implement subscriptions
}

func (c *Controller) ServeCDSControl(r *http.Request) *Response {
	action := strings.ToLower(
		regexp.MustCompile("\".*#(.+)\"").FindStringSubmatch(
			r.Header.Get("SOAPACTION"))[1])
	log.Printf("CDS Control Action received: %s\n", action)

	switch action {
	case "browse":
		body, _ := ioutil.ReadAll(r.Body)
		log.Printf("Request body: %s\n", body)

		objectIDStr := regexp.MustCompile("<ObjectID>(.+)</ObjectID>").FindStringSubmatch(string(body))[1]
		numItems := 0
		itemsResponse := ""
		if objectIDStr == "0" {
			// identify our root folders
			var dbFolders []*db.Album
			c.DB.
				Select("*").
				Where("parent_id is NULL").
				Find(&dbFolders)

			// render result
			for _, dbFolder := range dbFolders {
				if dbFolder.RightPath != "." {
					itemsResponse = fmt.Sprintf(`%s<container id="%d" parentID="0" restricted="1">
					<dc:title>%s</dc:title>
					<upnp:class>object.container.storageFolder</upnp:class>
				</container>`, itemsResponse, dbFolder.ID, dbFolder.RightPath)
					numItems = numItems + 1
				}
			}
		} else if strings.HasPrefix(objectIDStr, "track:") {
			// not a folder browse, but one regarding a track
			trackId, _ := strconv.Atoi(objectIDStr[6:])
			track := &db.Track{}
			err := c.DB.
				Preload("Album").
				Preload("Artist").
				First(track, trackId).
				Error
			if err != nil {
				return &Response{code: http.StatusInternalServerError, err: fmt.Sprintf("%v", err)}
			}
			itemsResponse = fmt.Sprintf(`<item id="track:%d" parentID="%d" restricted="1">
					<dc:title>%s</dc:title>
					<dc:creator>%s</dc:creator>
					<upnp:class>object.item.audioItem</upnp:class>
					<upnp:artist>%s</upnp:artist>
					<upnp:album>%s</upnp:album>
				</item>`, track.ID, track.AlbumID, track.TagTitle, track.Artist.Name, track.Artist.Name, track.Album.RightPath)
			numItems = 1
		} else {
			var childFolders []*db.Album
			c.DB.
				Where("parent_id=?", objectIDStr).
				Find(&childFolders)
			for _, c := range childFolders {
				itemsResponse = fmt.Sprintf(`%s<container id="%d" parentID="%s" restricted="1">
					<dc:title>%s</dc:title>
					<upnp:class>object.container.storageFolder</upnp:class>
				</container>`, itemsResponse, c.ID, objectIDStr, c.RightPath)
				numItems = numItems + 1
			}

			var childTracks []*db.Track
			c.DB.
				Where("album_id=?", objectIDStr).
				Preload("Album").
				Order("filename").
				Find(&childTracks)
			for _, child := range childTracks {
				itemsResponse = fmt.Sprintf(`%s<item id="track:%d" parentID="%s" restricted="1">
					<dc:title>%s</dc:title>
					<upnp:class>object.item.audioItem</upnp:class>
					<res protocolInfo="http-get:*:%s:DLNA.ORG_PS=1;DLNA_ORG_OP=00">http://%s/upnp/streaming?id=%d</res>
				</item>`, itemsResponse, child.ID, objectIDStr, child.Filename, child.MIME(), c.localEndpoint, child.ID)
				numItems = numItems + 1
			}
		}

		result := html.EscapeString(fmt.Sprintf(`<DIDL-Lite
					xmlns="urn:schemas-upnp-org:metadata-1-0/DIDL-Lite/"
					xmlns:dc="http://purl.org/dc/elements/1.1/"
					xmlns:upnp="urn:schemas-upnp-org:metadata-1-0/upnp/"
					xmlns:dlna="urn:schemas-dlna-org:metadata-1-0/">%s;</DIDL-Lite>`, itemsResponse))
		result = strings.ReplaceAll(result, "&#34;", "&quot;")
		result = strings.ReplaceAll(result, "&#39;", "&apos;")

		response := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
		<s:Envelope xmlns="urn:schemas-upnp-org:service-1-0" xmlns:s="http://schemas.xmlsoap.org/soap/envelope/" s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/">
			<s:Body>
				<u:BrowseResponse xmlns:u="urn:schemas-upnp-org:service:ContentDirectory:1">
					<Result>%s</Result>
					<NumberReturned>%d</NumberReturned>
					<TotalMatches>%d</TotalMatches>
					<UpdateID>1</UpdateID>
				</u:BrowseResponse>
			</s:Body>
		</s:Envelope>`, result, numItems, numItems)
		return &Response{code: http.StatusOK, responseData: []byte(response)}
	case "getsortcapabilities":
		// <?xml version="1.0" encoding="utf-8"?>
		// <s:Envelope s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/" xmlns:s="http://schemas.xmlsoap.org/soap/envelope/">
		//   <s:Body>
		//	   <u:GetSortCapabilities xmlns:u="urn:schemas-upnp-org:service:ContentDirectory:1">
		//	   </u:GetSortCapabilities>
		//	 </s:Body>
		// </s:Envelope>

		body, _ := ioutil.ReadAll(r.Body)
		log.Printf("Request body: %s\n", body)

		response := `<?xml version="1.0" encoding="UTF-8"?>
		<s:Envelope xmlns="urn:schemas-upnp-org:service-1-0" xmlns:s="http://schemas.xmlsoap.org/soap/envelope/" s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/">
			<s:Body>
				<u:GetSortCapabilitiesResponse xmlns:u="urn-schemas-upnp-org:service:ContentDirectory:1">
					<SortCaps>dc:title,dc:date,res@size</SortCaps>
				</u:GetSortCapabilitiesResponse>
			</s:Body>
		</s:Envelope>`
		return &Response{code: http.StatusOK, responseData: []byte(response)}
	default:
		return &Response{code: http.StatusNotImplemented, err: fmt.Sprintf("Unknown action '%s'\n", action)}
	}
}

func (c *Controller) ServeStream(w http.ResponseWriter, r *http.Request) *Response {
	idParam := r.URL.Query().Get("id")
	if idParam == "" {
		return &Response{code: http.StatusBadRequest, err: "Missing id parameter"}
	}

	track := &db.Track{}
	err := c.DB.
		Preload("Album").
		First(track, idParam).
		Error
	if gorm.IsRecordNotFoundError(err) {
		return &Response{code: http.StatusBadRequest, err: "media with id `%d` was not found"}
	}

	if r.Method == "HEAD" {
		log.Printf("Serving HEAD for %s", track.Filename)

		w.WriteHeader(200)
		w.Header().Set("Content-Length", string(track.Size))
		w.Header().Set("Content-Type", track.MIME())
		w.Header().Set("COntentFEatures.DLNA.ORG", "DNLNA.ORG_OP=00")
		w.Header().Set("TransferMode.DLNA.ORG", "Streaming")
		return nil
	} else {
		log.Printf("serving raw %q\n", track.Filename)

		lastModified, readerSeeker, err := c.MusicDir.GetFile(track.RelPath())
		if err != nil {
			return &Response{code: http.StatusInternalServerError, err: fmt.Sprintf("failed to get file: %v", err)}
		}
		w.Header().Set("Content-Type", track.MIME())
		http.ServeContent(w, r, track.Filename, lastModified, readerSeeker)
		if err := readerSeeker.Close(); err != nil {
			log.Printf("Failed to close input data: %v", err)
		}
		return nil
	}
}