package numcap

import (
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/stretchr/testify/assert"
	"github.com/traestan/whophone/internal/config"
	"github.com/traestan/whophone/internal/db"
	"go.uber.org/zap"
)

func Suites() (config.TomlConfig, *zap.Logger) {
	var (
		configFile = flag.String("config.toml", "../../config_test.toml", "Service port")
		config     config.TomlConfig
	)
	flag.Parse()

	if _, err := toml.DecodeFile(*configFile, &config); err != nil {
		panic(err)
	}
	// logger
	logger := zap.NewExample()
	defer logger.Sync()
	logger.Info("Who phone test start")

	return config, logger
}
func TestDownload(t *testing.T) {
	var wg sync.WaitGroup
	var filename = []string{
		"ABC-3xx.csv",
		"ABC-4xx.csv",
		"ABC-8xx.csv",
		"DEF-9xx.csv",
	}
	files := make(chan string, 4)
	for _, s := range filename {
		wg.Add(1)
		go func(s string) {
			fullURLFile := fmt.Sprintf("https://opendata.digital.gov.ru/downloads/%s", s)
			fileURL, err := url.Parse(fullURLFile)
			if err != nil {
				t.Log(err)
			}
			path := fileURL.Path
			segments := strings.Split(path, "/")
			fileName := segments[len(segments)-1]
			file, err := os.Create(fileName)
			if err != nil {
				t.Log(err)
			}
			roundTripper := &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					Timeout:   0 * time.Second,
					KeepAlive: 0 * time.Second,
				}).DialContext,
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			}

			httpClient := &http.Client{
				Transport: roundTripper,
				Timeout:   0 * time.Second,
			}
			resp, err := httpClient.Get(fullURLFile)
			if err != nil {
				t.Log(err)
			}
			defer resp.Body.Close()

			size, err := io.Copy(file, resp.Body)

			defer func() {
				file.Close()
				wg.Done()
			}()
			t.Log("Downloaded a file  with size  \n", fileName, size)
			files <- fileName

		}(s)
	}
	wg.Wait()
	select {
	case newState := <-files:
		//t.Log("info: work stopped")
		t.Log(newState)
	}
}

func TestCsv2Storage(t *testing.T) {
	cfg, logger := Suites()
	testStorage, err := db.NewClient(&cfg.Db, logger)
	if err != nil {
		t.Log(err)
	}
	nc, err := NewNumCap(testStorage, logger, cfg)
	if err != nil {
		t.Log(err)
	}

	state, err := nc.Csv2Storage()
	if err != nil {
		t.Log(err)
	}
	assert.Equal(t, state, true, "they should be equal")
	testStorage.StatsInfo()
}
