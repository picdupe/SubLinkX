package middlewares

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"sublink/models"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/text/encoding/simplifiedchinese"
	"gorm.io/gorm"
)

var ipInfoHTTPClient = &http.Client{Timeout: time.Second}
var ipInfoLookupSlots = make(chan struct{}, 16)

func GetIp(c *gin.Context) {
	c.Next()
	func() {
		subname, ok := c.Get("subname")
		if !ok {
			return
		}
		subnameStr, ok := subname.(string)
		if !ok || subnameStr == "" {
			return
		}

		ip := c.ClientIP()
		var sub models.Subcription
		if err := models.DB.Select("id").Where("name = ?", subnameStr).First(&sub).Error; err != nil {
			log.Printf("subscription access log lookup failed: %v", err)
			return
		}

		iplog := models.SubLogs{IP: ip}
		err := iplog.Find(sub.ID)
		now := time.Now().Format("2006-01-02 15:04:05")
		if err == nil {
			iplog.Count++
			iplog.Date = now
			if err := iplog.Update(); err != nil {
				log.Printf("subscription access log update failed: %v", err)
			}
			return
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("subscription access log lookup failed: %v", err)
			return
		}

		newIplog := models.SubLogs{
			IP:            ip,
			SubcriptionID: sub.ID,
			Date:          now,
			Count:         1,
		}
		if err := newIplog.Add(); err != nil {
			log.Printf("subscription access log insert failed: %v", err)
			return
		}
		enrichIPLogAddress(newIplog.ID, ip)
	}()
}

func enrichIPLogAddress(logID int, ip string) {
	select {
	case ipInfoLookupSlots <- struct{}{}:
		go func() {
			defer func() { <-ipInfoLookupSlots }()
			address := lookupIPAddress(ip)
			if address == "" {
				return
			}
			if err := models.DB.Model(&models.SubLogs{}).Where("id = ?", logID).Update("addr", address).Error; err != nil {
				log.Printf("subscription access location update failed: %v", err)
			}
		}()
	default:
		// Access logging must not wait for optional IP geolocation.
	}
}

func lookupIPAddress(ip string) string {
	query := url.Values{"ip": []string{ip}, "json": []string{"true"}}
	resp, err := ipInfoHTTPClient.Get("https://whois.pconline.com.cn/ipJson.jsp?" + query.Encode())
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return ""
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 64<<10))
	if err != nil {
		return ""
	}
	utf8Body, err := simplifiedchinese.GBK.NewDecoder().Bytes(body)
	if err != nil {
		return ""
	}
	var ipinfo struct {
		Addr string `json:"addr"`
	}
	if err := json.Unmarshal(utf8Body, &ipinfo); err != nil {
		return ""
	}
	return ipinfo.Addr
}
