package main

import (
	"fmt"
	"github.com/ViRb3/optic-go"
	"log"
	"net/http"
	"time"
)

var defaultHeaders = map[string]string{
	"Accept-Language": "en-US",
	"Connection":      "keep-alive",
	"Origin":          "https://www.twitch.tv",
	"Referer":         "https://www.twitch.tv",
	"Sec-Fetch-Mode":  "cors",
	"Sec-Fetch-Site":  "cross-site",
	"User-Agent": "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) " +
		"twitch-desktop-electron-platform/1.0.0 Chrome/78.0.3904.130 Electron/7.3.3 Safari/537.36 desklight/8.57.0",
}

type CustomTripper struct{}

func (t CustomTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	for key, val := range defaultHeaders {
		req.Header.Set(key, val)
	}
	return http.DefaultTransport.RoundTrip(req)
}

func main() {
	tester, err := opticgo.NewTester(opticgo.Config{
		ApiUrl:               opticgo.MustUrl("https://addons-ecs.forgesvc.net/api/v2/"),
		ProxyListenAddr:      "",
		OpticUrl:             opticgo.MustUrl("http://localhost:8889"),
		DebugPrint:           false,
		RoundTripper:         CustomTripper{},
		InternetCheckTimeout: 10 * time.Second,
	})
	if err != nil {
		log.Fatalln(err)
	}

	errChan, _ := tester.Start(getTests())
	errText := ""
	for err := range errChan {
		errText += err.Error() + "\n"
	}
	if errText != "" {
		log.Fatalln(errText)
	}
}

func getTests() []opticgo.TestDefinition {
	return []opticgo.TestDefinition{
		// ---------------------------------------------------------------------------------------------
		// Addon
		// ---------------------------------------------------------------------------------------------
		{
			"Get Addons Database Timestamp",
			nil,
			fmt.Sprintf("addon/timestamp"),
			"GET",
		},
		{
			"Addon Search",
			nil,
			fmt.Sprintf("addon/search"+
				"?categoryId=%d"+
				"&gameId=%d"+
				"&gameVersion=%s"+
				"&index=%d"+
				"&pageSize=%d"+
				"&searchFilter=%s"+
				"&sectionId=%d"+
				"&sort=%d",
				0, 432, "1.12.2", 0, 25, "ultimate", 4471, 0),
			"GET",
		},
		{
			"Get Addon Info",
			nil,
			fmt.Sprintf("addon/%d", 310806),
			"GET",
		},
		{
			"Get Multiple Addons",
			[]int64{310806, 304026},
			fmt.Sprintf("addon"),
			"POST",
		},
		{
			"Get Featured Addons",
			struct {
				GameId        int     `json:"GameId"`
				AddonIds      []int64 `json:"addonIds"`
				FeaturedCount int     `json:"featuredCount"`
				PopularCount  int     `json:"popularCount"`
				UpdatedCount  int     `json:"updatedCount"`
			}{
				432,
				[]int64{},
				6,
				14,
				14,
			},
			fmt.Sprintf("addon/featured"),
			"POST",
		},
		{
			"Get Addon by Fingerprint",
			[]int64{3028671922},
			"fingerprint",
			"POST",
		},
		{
			"Get Addon Description",
			nil,
			fmt.Sprintf("addon/%d/description", 310806),
			"GET",
		},
		{
			"Get Addon File Changelog",
			nil,
			fmt.Sprintf("addon/%d/file/%d/changelog", 310806, 2657461),
			"GET",
		},
		{
			"Get Addon File Download URL",
			nil,
			fmt.Sprintf("addon/%d/file/%d/download-url", 296062, 2724357),
			"GET",
		},
		{
			"Get Addon File Information",
			nil,
			fmt.Sprintf("addon/%d/file/%d", 310806, 2657461),
			"GET",
		},
		{
			"Get Addon Files",
			nil,
			fmt.Sprintf("addon/%d/files", 304026),
			"GET",
		},
		// ---------------------------------------------------------------------------------------------
		// Category
		// ---------------------------------------------------------------------------------------------
		{
			"Get Category Info",
			nil,
			fmt.Sprintf("category/%d", 423),
			"GET",
		},
		{
			"Get Category List",
			nil,
			fmt.Sprintf("category"),
			"GET",
		},
		{
			"Get Category Section Info",
			nil,
			fmt.Sprintf("category/section/%d", 6),
			"GET",
		},
		{
			"Get Category Timestamp",
			nil,
			fmt.Sprintf("category/timestamp"),
			"GET",
		},
		// ---------------------------------------------------------------------------------------------
		// Game
		// ---------------------------------------------------------------------------------------------
		{
			"Get Game Info",
			nil,
			fmt.Sprintf("game/%d", 432),
			"GET",
		},
		{
			"Get Games List",
			nil,
			fmt.Sprintf("game"),
			"GET",
		},
		{
			"Get Addon-Supported Games List",
			nil,
			fmt.Sprintf("game?supportsAddons"),
			"GET",
		},
		// ---------------------------------------------------------------------------------------------
		// Minecraft
		// ---------------------------------------------------------------------------------------------
		{
			"Get Minecraft Version Info",
			nil,
			fmt.Sprintf("minecraft/version/%s", "1.12.2"),
			"GET",
		},
		{
			"Get Minecraft Version List",
			nil,
			fmt.Sprintf("minecraft/version"),
			"GET",
		},
		{
			"Get Minecraft Version Timestamp",
			nil,
			fmt.Sprintf("minecraft/version/timestamp"),
			"GET",
		},
		{
			"Get Modloaders for Version",
			nil,
			fmt.Sprintf("minecraft/modloader?version=1.12.2"),
			"GET",
		},
		{
			"Get Modloader Info",
			nil,
			fmt.Sprintf("minecraft/modloader/%s", "forge-12.17.0.1980"),
			"GET",
		},
		{
			"Get Modloader List",
			nil,
			fmt.Sprintf("minecraft/modloader"),
			"GET",
		},
		{
			"Get Modloader Timestamp",
			nil,
			fmt.Sprintf("minecraft/modloader/timestamp"),
			"GET",
		},
	}
}
