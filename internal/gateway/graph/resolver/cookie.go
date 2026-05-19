package resolver

import "net/http"

func (r *Resolver) setRefreshTokenCookie(w http.ResponseWriter, token string) {
	//TODO: Fix the issue of Secure field. It must be set to true in production. Remove the secure check when in production and switch to HTTPS!!!!
	secure := r.cfg.Application.Environment != "development"
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    token,
		Path:     "/",
		Domain:   "",
		MaxAge:   int(r.cfg.JWT.RefreshTokenExpires.Seconds()),
		Secure:   secure,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
}

func (r *Resolver) clearRefreshTokenCookie(w http.ResponseWriter) {
	//TODO: Fix the issue of Secure field. It must be set to true in production. Remove the secure check when in production and switch to HTTPS!!!!
	secure := r.cfg.Application.Environment != "development"
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Secure:   secure,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
}
