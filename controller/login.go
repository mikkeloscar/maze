package controller

import (
	"encoding/base32"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/drone/drone/shared/httputil"
	"github.com/drone/drone/shared/token"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/securecookie"
	"github.com/mikkeloscar/maze/model"
	"github.com/mikkeloscar/maze/remote"
	"github.com/mikkeloscar/maze/store"
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

	httputil.SetCookie(c.Writer, c.Request, "user_sess", tokenstr)
	redirect := httputil.GetCookie(c.Request, "user_last")
	if len(redirect) == 0 {
		redirect = "/"
	}
	c.Redirect(http.StatusSeeOther, redirect)
}

func GetLogout(c *gin.Context) {
	httputil.DelCookie(c.Writer, c.Request, "user_sess")
	httputil.DelCookie(c.Writer, c.Request, "user_last")
	c.Redirect(http.StatusSeeOther, "/login")
}
