package cookie

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"

	"github.com/golang-jwt/jwt/v5"
)

type CookieStruct struct {
	Email      string `json:"email"`
	GivenName  string `json:"givenName"`
	FamilyName string `json:"familyName"`
	Picture    string `json:"picture"`
	UserID     string `json:"userID"`
	jwt.MapClaims
}

func getCookie(r *http.Request, name string) (*http.Cookie, error) {
	cookie, err := r.Cookie(name)
	if err != nil {
		log.Println("Cookie not found:", name)
		return nil, nil // return nil, nil if cookie not found
	}

	return cookie, nil
}

func parseCookie(cookie *http.Cookie) (CookieStruct, error) {
	var cookieValue CookieStruct
	cookieValueUrl, err := url.QueryUnescape(cookie.Value)
	if err != nil {
		log.Println("Error unescaping token cookie:", err)
		return CookieStruct{}, err
	}

	if cookieValueUrl == "" {
		return CookieStruct{}, nil
	}

	err = json.Unmarshal([]byte(cookieValueUrl), &cookieValue)
	if err != nil {
		log.Println("Error parsing token cookie:", err)
		return CookieStruct{}, err
	}

	return cookieValue, nil
}

func HandleCookie(r *http.Request, name string) (CookieStruct, error) {
	cookie, err := getCookie(r, name)
	if err != nil {
		return CookieStruct{}, err
	}
	if cookie == nil {
		// If cookie is not found, return default value
		return CookieStruct{}, nil
	}

	cookieValue, err := parseCookie(cookie)
	if err != nil {
		return CookieStruct{}, err
	}

	return cookieValue, nil
}
