package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// google token generation

const hdr = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9"

type JwtPayload struct {
	Iss   string `json:"iss"`
	Scope string `json:"scope"`
	Aud   string `json:"aud"`
	Exp   int64  `json:"exp"`
	Iat   int64  `json:"iat"`
}

type GoogleCred struct {
	Type                    string `json:"type"`
	ProjectID               string `json:"project_id"`
	PrivateKeyID            string `json:"private_key_id"`
	PrivateKey              string `json:"private_key"`
	ClientEmail             string `json:"client_email"`
	ClientID                string `json:"client_id"`
	AuthUri                 string `json:"auth_uri"`
	TokenUri                string `json:"token_uri"`
	AuthProviderX509CertUrl string `json:"auth_provider_x509_cert_url"`
	ClientX509CertUrl       string `json:"client_x509_cert_url"`
}

type GoogleToken struct {
	AccessToken string `json:"access_token"`
	Scope       string `json:"scope"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

func NewGoogleTranslateJwt() (*GoogleToken, error) {
	fc, err := ioutil.ReadFile("google_cred.json")
	if err != nil {
		return nil, err
	}
	var cred GoogleCred
	if err := json.Unmarshal(fc, &cred); err != nil {
		return nil, err
	}
	if cred.Type != "service_account" {
		return nil, fmt.Errorf("Expected service account credentials got: %s\n", cred.Type)
	}

	now := time.Now()

	tokenUri := cred.TokenUri

	turl, err := url.Parse(tokenUri)
	if err != nil {
		return nil, err
	}
	tokenHost := turl.Host

	payload := JwtPayload{
		Iss:   cred.ClientEmail,
		Scope: "https://www.googleapis.com/auth/cloud-translation",
		Aud:   tokenUri,
		Exp:   now.Add(time.Minute * 30).Unix(),
		Iat:   now.Unix(),
	}

	payloadJson, err := json.Marshal(&payload)
	if err != nil {
		return nil, err
	}

	payloadBase64 := base64.URLEncoding.EncodeToString(payloadJson)

	block, _ := pem.Decode([]byte(cred.PrivateKey))
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	sgn := fmt.Sprintf("%s.%s", hdr, payloadBase64)
	hash := sha256.Sum256([]byte(sgn))
	signed, err := rsa.SignPKCS1v15(rand.Reader, key.(*rsa.PrivateKey), crypto.SHA256, hash[:])
	if err != nil {
		return nil, err
	}

	signed64 := base64.URLEncoding.EncodeToString(signed)

	jwt := fmt.Sprintf("%s.%s", sgn, signed64)

	d := url.Values{}
	d.Set("grant_type", "urn:ietf:params:oauth:grant-type:jwt-bearer")
	d.Set("assertion", jwt)

	encData := d.Encode()

	req, err := http.NewRequest("POST", cred.TokenUri, strings.NewReader(encData))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(encData)))
	req.Header.Add("Host", tokenHost)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("%s\n%s", res.Status, string(resBody))
	}

	googleToken := new(GoogleToken)
	if err := json.Unmarshal(resBody, googleToken); err != nil {
		return nil, err
	}

	return googleToken, nil
}
