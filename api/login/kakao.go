package login

import (
	"encoding/json"
	"fmt"
	"github.com/4538cgy/backend-second/api"
	"github.com/labstack/echo/v4"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const (
	kakaoTokenAddress      = "https://kauth.kakao.com/oauth/token"
	kakaoAuthCodeParamName = "code"
	kakaoRestApiKey        = "85e6e9b31bb205301e2db6572041d673"
	kakaoRedirectUri       = "http://localhost/oauth/kakao"
	kakaoAuthUri           = "/oauth/kakao"
	kakaoMyInfoAddress     = "https://kapi.kakao.com/v2/user/me"
)

type KakaoTokenInfo struct {
	AccessToken           string `json:"access_token"`
	TokenType             string `json:"token_type"`
	RefreshToken          string `json:"refresh_token"`
	ExpiresIn             uint32 `json:"expires_in"`
	RefreshTokenExpiresIn uint32 `json:"refresh_token_expires_in"`
}

func init() {
	e := api.Echo()
	e.GET(kakaoAuthUri, login)
}

func login(ctx echo.Context) error {
	auth := ctx.QueryParam(kakaoAuthCodeParamName)
	form := url.Values{}
	form.Add("grant_type", "authorization_code")
	form.Add("client_id", kakaoRestApiKey)
	form.Add("redirect_uri", kakaoRedirectUri)
	form.Add("code", auth)
	client := http.Client{}
	req, err := http.NewRequest("POST", kakaoTokenAddress, strings.NewReader(form.Encode()))
	if err != nil {
		return ctx.String(http.StatusInternalServerError, "fail")
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	res, err := client.Do(req)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, fmt.Sprintf("failed. %s", err.Error()))
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, fmt.Sprintf("failed. %s", err.Error()))
	}
	go func() {
		ktf := KakaoTokenInfo{}
		err = json.Unmarshal(body, &ktf)
		if err != nil {
			return
		}
		client2 := http.Client{}
		form2 := url.Values{}
		form2.Add("property_keys", "[\"kakao_account.email\"]")
		req2, err := http.NewRequest("GET", kakaoMyInfoAddress, nil)
		if err != nil {
			return
		}
		req2.Header.Add("Authorization", fmt.Sprintf("Bearer %s", ktf.AccessToken))
		resp2, err := client2.Do(req2)
		if err != nil {
			return
		}
		defer resp2.Body.Close()
		body2, err := ioutil.ReadAll(resp2.Body)
		if err != nil {
			return
		}
		fmt.Println("yes => ", string(body2))
	}()
	return ctx.String(http.StatusOK, string(body))
}
