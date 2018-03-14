// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package utils

import (
	"crypto"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strings"

	"github.com/SoulDemon/mattermostp/model"
)

func CheckOrigin(r *http.Request, allowedOrigins string) bool {
	origin := r.Header.Get("Origin")
	if allowedOrigins == "*" {
		return true
	}
	for _, allowed := range strings.Split(allowedOrigins, " ") {
		if allowed == origin {
			return true
		}
	}
	return false
}

func OriginChecker(allowedOrigins string) func(*http.Request) bool {
	return func(r *http.Request) bool {
		return CheckOrigin(r, allowedOrigins)
	}
}

func RenderWebAppError(w http.ResponseWriter, r *http.Request, err *model.AppError, s crypto.Signer) {
	RenderWebError(w, r, err.StatusCode, url.Values{
		"message": []string{err.Message},
	}, s)
}

func RenderWebError(w http.ResponseWriter, r *http.Request, status int, params url.Values, s crypto.Signer) {
	queryString := params.Encode()

	h := crypto.SHA256
	sum := h.New()
	sum.Write([]byte("/error?" + queryString))
	signature, err := s.Sign(rand.Reader, sum.Sum(nil), h)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	destination := "/error?" + queryString + "&s=" + base64.URLEncoding.EncodeToString(signature)

	if status >= 300 && status < 400 {
		http.Redirect(w, r, destination, status)
		return
	}

	w.WriteHeader(status)
	fmt.Fprintln(w, `<!DOCTYPE html><html><head></head>`)
	fmt.Fprintln(w, `<body onload="window.location = '`+template.HTMLEscapeString(template.JSEscapeString(destination))+`'">`)
	fmt.Fprintln(w, `<noscript><meta http-equiv="refresh" content="0; url=`+template.HTMLEscapeString(destination)+`"></noscript>`)
	fmt.Fprintln(w, `<a href="`+template.HTMLEscapeString(destination)+`" style="color: #c0c0c0;">...</a>`)
	fmt.Fprintln(w, `</body></html>`)
}
