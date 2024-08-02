package handler

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	rtctokenbuilder "github.com/AgoraIO-Community/go-tokenbuilder/rtctokenbuilder"
	rtmtokenbuilder "github.com/AgoraIO-Community/go-tokenbuilder/rtmtokenbuilder"
)

var appID string
var appCertificate string

func init() {
	appID = os.Getenv("9fb93470853b43df9007aa095a998f14")
	appCertificate = os.Getenv("0090c9c3ea294f87bd83a82629ec00cd")
	if appID == "" || appCertificate == "" {
		log.Fatal("FATAL ERROR: ENV not properly configured, check APP_ID and APP_CERTIFICATE")
	}
}

func Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	switch r.URL.Path {
	case "/api/ping":
		handlePing(w, r)
	case "/api/rtc":
		handleRtc(w, r)
	case "/api/rtm":
		handleRtm(w, r)
	case "/api/rte":
		handleRte(w, r)
	default:
		http.Error(w, "Not found", http.StatusNotFound)
	}
}

func handlePing(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("pong"))
}

func handleRtc(w http.ResponseWriter, r *http.Request) {
	channelName := r.URL.Query().Get("channelName")
	role := r.URL.Query().Get("role")
	tokentype := r.URL.Query().Get("tokentype")
	uidStr := r.URL.Query().Get("uid")
	expireTime := r.URL.Query().Get("expiry")

	var rtcRole rtctokenbuilder.Role
	if role == "publisher" {
		rtcRole = rtctokenbuilder.RolePublisher
	} else {
		rtcRole = rtctokenbuilder.RoleSubscriber
	}

	expireTime64, parseErr := strconv.ParseUint(expireTime, 10, 64)
	if parseErr != nil {
		http.Error(w, "Error parsing expiry", http.StatusBadRequest)
		return
	}

	expireTimeInSeconds := uint32(expireTime64)
	currentTimestamp := uint32(time.Now().UTC().Unix())
	expireTimestamp := currentTimestamp + expireTimeInSeconds

	rtcToken, tokenErr := generateRtcToken(channelName, uidStr, tokentype, rtcRole, expireTimestamp)
	if tokenErr != nil {
		http.Error(w, "Error generating RTC token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"rtcToken": "%s"}`, rtcToken)
}

func handleRtm(w http.ResponseWriter, r *http.Request) {
	uidStr := r.URL.Query().Get("uid")
	expireTime := r.URL.Query().Get("expiry")

	expireTime64, parseErr := strconv.ParseUint(expireTime, 10, 64)
	if parseErr != nil {
		http.Error(w, "Error parsing expiry", http.StatusBadRequest)
		return
	}

	expireTimeInSeconds := uint32(expireTime64)
	currentTimestamp := uint32(time.Now().UTC().Unix())
	expireTimestamp := currentTimestamp + expireTimeInSeconds

	rtmToken, tokenErr := rtmtokenbuilder.BuildToken(appID, appCertificate, uidStr, expireTimestamp, "")
	if tokenErr != nil {
		http.Error(w, "Error generating RTM token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"rtmToken": "%s"}`, rtmToken)
}

func handleRte(w http.ResponseWriter, r *http.Request) {
	channelName := r.URL.Query().Get("channelName")
	role := r.URL.Query().Get("role")
	tokentype := r.URL.Query().Get("tokentype")
	uidStr := r.URL.Query().Get("uid")
	expireTime := r.URL.Query().Get("expiry")

	var rtcRole rtctokenbuilder.Role
	if role == "publisher" {
		rtcRole = rtctokenbuilder.RolePublisher
	} else {
		rtcRole = rtctokenbuilder.RoleSubscriber
	}

	expireTime64, parseErr := strconv.ParseUint(expireTime, 10, 64)
	if parseErr != nil {
		http.Error(w, "Error parsing expiry", http.StatusBadRequest)
		return
	}

	expireTimeInSeconds := uint32(expireTime64)
	currentTimestamp := uint32(time.Now().UTC().Unix())
	expireTimestamp := currentTimestamp + expireTimeInSeconds

	rtcToken, rtcTokenErr := generateRtcToken(channelName, uidStr, tokentype, rtcRole, expireTimestamp)
	rtmToken, rtmTokenErr := rtmtokenbuilder.BuildToken(appID, appCertificate, uidStr, expireTimestamp, "")
	if rtcTokenErr != nil || rtmTokenErr != nil {
		http.Error(w, "Error generating tokens", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"rtcToken": "%s", "rtmToken": "%s"}`, rtcToken, rtmToken)
}

func generateRtcToken(channelName, uidStr, tokentype string, role rtctokenbuilder.Role, expireTimestamp uint32) (rtcToken string, err error) {
	if tokentype == "userAccount" {
		rtcToken, err = rtctokenbuilder.BuildTokenWithAccount(appID, appCertificate, channelName, uidStr, role, expireTimestamp)
	} else if tokentype == "uid" {
		uid64, parseErr := strconv.ParseUint(uidStr, 10, 64)
		if parseErr != nil {
			return "", fmt.Errorf("failed to parse uidStr: %s, to uint causing error: %s", uidStr, parseErr)
		}
		uid := uint32(uid64)
		rtcToken, err = rtctokenbuilder.BuildTokenWithUid(appID, appCertificate, channelName, uid, role, expireTimestamp)
	} else {
		return "", fmt.Errorf("failed to generate RTC token for Unknown Tokentype: %s", tokentype)
	}
	return rtcToken, err
}
