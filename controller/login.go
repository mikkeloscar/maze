package controller

import (
	"encoding/base32"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/securecookie"
	"github.com/mikkeloscar/maze/model"
	"github.com/mikkeloscar/maze/pkg/token"
	"github.com/mikkeloscar/maze/remote"
	"github.com/mikkeloscar/maze/store"
	log "github.com/sirupsen/logrus"
)

func GetLogin(c *gin.Context) {
	remote := remote.FromContext(c)

	tmpUser, err := remote.Login(c.Writer, c.Request)
	if err != nil {
		log.Errorf("failed to authenticate user. %s", err)
		c.Redirect(http.StatusSeeOther, "/login?error=oauth_error")
		return
	}

	if tmpUser == nil {
		return
	}

	u, err := store.GetUserLogin(c, tmpUser.Login)
	if err != nil {
		count, err := store.CountUsers(c)
		if err != nil {
			log.Errorf("cannot register %s. %s", tmpUser.Login, err)
			c.Redirect(http.StatusSeeOther, "/login?error=internal_error")
			return
		}

		// if self-registration is disabled we should
		// return a notAuthorized error. the only exception
		// is if no users exist yet in the system we'll proceed.
		if count != 0 {
			log.Errorf("failed to register %s.", tmpUser.Login)
			c.Redirect(http.StatusSeeOther, "/login?error=access_denied")
			return
		}

		// create the user account
		u = &model.User{}
		u.Login = tmpUser.Login
		u.Hash = base32.StdEncoding.EncodeToString(
			securecookie.GenerateRandomKey(32),
		)

		// insert the user into the database
		if err := store.CreateUser(c, u); err != nil {
			log.Errorf("failed to insert %s. %s", u.Login, err)
			c.Redirect(http.StatusSeeOther, "/login?error=internal_error")
			return
		}

		// if this is the first user, they
		// should be an admin.
		if count == 0 {
			u.Admin = true
		}
	}

	// update the user meta data and authorization
	// data and cache in the datastore.
	u.Token = tmpUser.Token

	if err := store.UpdateUser(c, u); err != nil {
		log.Errorf("failed to update %s. %s", u.Login, err)
		c.Redirect(http.StatusSeeOther, "/login?error=internal_error")
		return
	}

	exp := time.Now().Add(time.Hour * 72).Unix()
	token := token.New(token.SessToken, u.Login)
	tokenstr, err := token.SignExpires(u.Hash, exp)
	if err != nil {
		log.Errorf("failed to create token for %s. %s", u.Login, err)
		c.Redirect(http.StatusSeeOther, "/login?error=internal_error")
		return
	}

	cookie := http.Cookie{
		Name:     "user_sess",
		Value:    tokenstr,
		Path:     "/",
		Domain:   c.Request.URL.Host,
		HttpOnly: true,
		Secure:   isHttps(c.Request),
		MaxAge:   2147483647, // the cooke value (token) is responsible for expiration
	}
	http.SetCookie(c.Writer, &cookie)

	redirectCookie, err := c.Request.Cookie("user_last")
	if err != nil {
		return
	}
	redirect := redirectCookie.Value
	if len(redirect) == 0 {
		redirect = "/"
	}
	c.Redirect(http.StatusSeeOther, redirect)
}

func GetLogout(c *gin.Context) {
	deleteCookie(c.Writer, c.Request, "user_sess")
	deleteCookie(c.Writer, c.Request, "user_last")
	c.Redirect(http.StatusSeeOther, "/login")
}

// deleteCookie deletes a cookie.
func deleteCookie(w http.ResponseWriter, r *http.Request, name string) {
	cookie := http.Cookie{
		Name:   name,
		Value:  "deleted",
		Path:   "/",
		Domain: r.URL.Host,
		MaxAge: -1,
	}

	http.SetCookie(w, &cookie)
}

// isHttps is a helper function that evaluates the http.Request
// and returns True if the Request uses HTTPS. It is able to detect,
// using the X-Forwarded-Proto, if the original request was HTTPS and
// routed through a reverse proxy with SSL termination.
func isHttps(r *http.Request) bool {
	switch {
	case r.URL.Scheme == "https":
		return true
	case r.TLS != nil:
		return true
	case strings.HasPrefix(r.Proto, "HTTPS"):
		return true
	case r.Header.Get("X-Forwarded-Proto") == "https":
		return true
	default:
		return false
	}
}
