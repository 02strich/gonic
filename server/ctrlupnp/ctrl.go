package ctrlupnp

import (
	"log"
	"net/http"
	"senan.xyz/g/gonic/server/ctrlbase"
	"strings"
)

type Controller struct {
	*ctrlbase.Controller
	localEndpoint string
}

// DeviceUUID is the UPnP UUID we are broadcasting and announcing
const DeviceUUID = "c000ffee-cafe-c0c0-dead-c000ffffeeee"

func New(base *ctrlbase.Controller, frontendEndpoint string) *Controller {
	return &Controller{
		Controller: base,
		localEndpoint: frontendEndpoint,
	}
}

type Response struct {
	code int
	// code is 200
	responseData []byte
	// code is 303
	redirect string
	// code is >= 400
	err  string
}

type handlerUPnP func(r *http.Request) *Response

func (c *Controller) H(h handlerUPnP) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := h(r)
		if resp.redirect != "" {
			to := resp.redirect
			if strings.HasPrefix(to, "/") {
				to = c.Path(to)
			}
			http.Redirect(w, r, to, http.StatusSeeOther)
			return
		}
		if resp.err != "" {
			http.Error(w, resp.err, resp.code)
			return
		}
		w.WriteHeader(resp.code)
		_, err := w.Write(resp.responseData)
		if err != nil {
			log.Fatal("Failed to write response")
		}
	})
}

type handlerUPnPRaw func(w http.ResponseWriter, r *http.Request) *Response

func (c *Controller) HR(h handlerUPnPRaw) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := h(w, r)
		if resp == nil {
			return
		}
		if resp.err != "" {
			http.Error(w, resp.err, resp.code)
			return
		}
		w.WriteHeader(resp.code)
		_, err := w.Write(resp.responseData)
		if err != nil {
			log.Fatal("Failed to write response")
		}
	})
}
