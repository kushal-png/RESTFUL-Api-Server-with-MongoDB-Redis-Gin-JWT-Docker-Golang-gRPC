package utils

import (
	"net/http"
	"project/initializers"
)

func SetCookies(rw http.ResponseWriter, accessToken, refreshToken string, config *initializers.Config) {
	http.SetCookie(rw, &http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		MaxAge:   config.AccessTokenMaxAge,
		Path:     "/",
		Domain:   "localhost",
		Secure:   false,
		HttpOnly: true,
		SameSite: http.SameSiteNoneMode,
	})
	http.SetCookie(rw, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		MaxAge:   config.RefreshTokenMaxAge,
		Path:     "/",
		Domain:   "localhost",
		Secure:   false,
		HttpOnly: true,
		SameSite: http.SameSiteNoneMode,
	})
	http.SetCookie(rw, &http.Cookie{
		Name:     "logged_in",
		Value:    "true",
		MaxAge:   config.AccessTokenMaxAge,
		Path:     "/",
		Domain:   "localhost",
		Secure:   false,
		HttpOnly: false,
		SameSite: http.SameSiteNoneMode,
	})

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	// Send JSON response
	_, _ = rw.Write([]byte(`{"status": "success", "access_token": "` + accessToken + `"}`))
}
