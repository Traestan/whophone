package numcap

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/traestan/whophone/internal/config"
	"github.com/traestan/whophone/internal/db"
	"github.com/traestan/whophone/internal/models"
	"go.uber.org/zap"
)

type Numcap interface {
	Download() (bool, error)
	Csv2Storage() (bool, error)
}
type numcap struct {
	db     *db.Client
	logger *zap.Logger
	cfg    config.TomlConfig
}

func NewNumCap(db *db.Client, logger *zap.Logger, cfg config.TomlConfig) (Numcap, error) {
	nc := &numcap{
		logger: logger,
		db:     db,
		cfg:    cfg,
	}
	return nc, nil
}

// Download files from opendata
func (nc *numcap) Download() (bool, error) {
	var wg sync.WaitGroup
	files := make(chan string, 4)

	for _, s := range nc.cfg.Numcap.Filename {
		wg.Add(1)
		go func(s string) {
			fullURLFile := fmt.Sprintf("https://opendata.digital.gov.ru/downloads/%s", s)
			fileURL, err := url.Parse(fullURLFile)
			if err != nil {
				nc.logger.Info("Download ", zap.String("update base error ", err.Error()))
			}
			path := fileURL.Path
			segments := strings.Split(path, "/")
			fileName := segments[len(segments)-1]
			file, err := os.Create(fileName)
			if err != nil {
				nc.logger.Info("Download ", zap.String("update base error", err.Error()))
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
				nc.logger.Info("Download ", zap.String("update base error", err.Error()))
			}
			defer resp.Body.Close()

			size, err := io.Copy(file, resp.Body)

			defer func() {
				file.Close()
				wg.Done()
			}()
			nc.logger.Info(fmt.Sprintf("Downloaded a file %s with size %d \n", fileName, size))
			files <- fileName

		}(s)
	}
	wg.Wait()

	select {
	case newState := <-files:
		nc.logger.Info("Download ", zap.String("update base normal", newState))
	}

	return true, nil
}

// Csv2Storage parses the csv file and adds it to boltdb
func (nc *numcap) Csv2Storage() (bool, error) {
	var wg sync.WaitGroup
	for _, s := range nc.cfg.Numcap.Filename {
		wg.Add(1)
		go func(s string) {
			fullPathFile := fmt.Sprintf("./%s", s)
			readFile, err := os.Open(fullPathFile)
			if err != nil {
				nc.logger.Error("Csv2Storage", zap.String("Error", err.Error()))
			}
			fileScanner := bufio.NewScanner(readFile)

			fileScanner.Split(bufio.ScanLines)
			for fileScanner.Scan() {
				line := strings.Split(fileScanner.Text(), ";")
				number := models.Numbers{
					Code:     line[0],
					Begin:    line[1],
					End:      line[2],
					Capacity: line[3],
					Operator: line[4],
					Region:   line[5],
					Inn:      line[6],
				}

				err := nc.db.PutHash(number)
				if err != nil {
					nc.logger.Error("Csv2Storage", zap.String("Error", err.Error()))
				}
				nc.logger.Info("Phone added ", zap.String("Code", line[0]), zap.String("Begin", line[0]))
			}
			defer func() {
				readFile.Close()
				nc.logger.Info("Deleting the file ", zap.String("", fullPathFile))
				if err := os.Remove(fullPathFile); err != nil {
					nc.logger.Fatal("File not deleted ", zap.String("error", err.Error()))
				}
				wg.Done()
			}()
		}(s)
	}
	wg.Wait()
	return true, nil
}
