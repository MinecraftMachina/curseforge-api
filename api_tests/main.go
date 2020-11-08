package main

import (
	"curseforge-api/util"
	"fmt"
	"github.com/ViRb3/sling/v2"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"
)

var globalClient *sling.Sling
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
var customTransport = &CustomTransport{}
var parsedApiEndpoint *url.URL

const (
	apiEndpoint  = "https://addons-ecs.forgesvc.net/api/v2/"
	DebugPrint   = false // dump requests and responses
	OpticPort    = "8888"
	BypassOptic  = false
	RequestDelay = 3 * time.Second // prevent rate limit ban
)

var fixedTransport = &http.Transport{}

type CustomTransport struct{}

func (CustomTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	log.Println("Requesting:", r.URL.String())
	r.Header.Del("X-Forwarded-For") // inserted by httputil.NewSingleHostReverseProxy
	if DebugPrint {
		b, err := httputil.DumpRequestOut(r, false)
		if err != nil {
			return nil, err
		}
		fmt.Println(string(b))
	}
	return fixedTransport.RoundTrip(r)
}

func init() {
	parsed, err := url.Parse(apiEndpoint)
	if err != nil {
		log.Fatal(err)
		return
	}
	parsedApiEndpoint = parsed
	doer, _ := util.NewBaseDoer()
	globalClient = sling.New().
		Doer(doer).
		SetMany(defaultHeaders).
		Path("http://localhost:" + OpticPort).
		Path(parsed.Path)
}

func serve() error {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		proxy := httputil.NewSingleHostReverseProxy(parsedApiEndpoint)
		proxy.Director = func(r *http.Request) {
			r.Host = parsedApiEndpoint.Host
			r.URL = parsedApiEndpoint.ResolveReference(r.URL)
		}
		proxy.Transport = customTransport
		proxy.ServeHTTP(w, r)
		log.Println("Returned:", r.URL.String())
	})
	if BypassOptic {
		return http.ListenAndServe(":"+OpticPort, nil)
	} else {
		return http.ListenAndServe(":"+os.Getenv("OPTIC_API_PORT"), nil)
	}
}

func main() {
	waitInternetAccess()
	go func() { log.Fatal(serve()) }()
	if err := testAPI(); err != nil {
		log.Println(err)
	}
	// wait for last request to be returned to Optic
	time.Sleep(1 * time.Second)
}

// e.g. get firewall permission
func waitInternetAccess() {
	for {
		_, err := http.Get("https://google.com/")
		if err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}
}

type TestConfig struct {
	name       string
	data       interface{}
	requestUrl string
	method     string
}

func runTest(config *TestConfig) error {
	log.Println("Running test:", config.requestUrl)
	request := globalClient.New().Method(config.method).Path(config.requestUrl)
	if config.data != nil {
		request.BodyJSON(config.data)
	}
	resp, err := request.ReceiveBody()
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if DebugPrint {
		b, err := httputil.DumpResponse(resp, true)
		if err != nil {
			return err
		}
		log.Println(string(b))
	}
	return nil
}

func testAPI() error {
	var tests = []TestConfig{
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

	log.Printf("Defined %d tests\n", len(tests))
	for _, test := range tests {
		if err := runTest(&test); err != nil {
			return err
		}
		time.Sleep(RequestDelay)
	}
	return nil
}
