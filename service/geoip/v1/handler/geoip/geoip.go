package geoip

import (
	"context"
	"net"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"go-micro.dev/v4/errors"
	"wz2100.net/microlobby/service/geoip/v1/internal/maxmind"
	"wz2100.net/microlobby/service/geoip/v1/proto/geoippb"
	"wz2100.net/microlobby/shared/component"

	"github.com/oschwald/geoip2-golang"
	"github.com/sirupsen/logrus"
)

type Config struct {
	RefreshSeconds int
	AccountID      string
	LicenseKey     string
	DataDirectory  string
}

type Handler struct {
	sync.RWMutex

	cRegistry *component.Registry
	logrus    *logrus.Logger

	config Config

	refreshCancel context.CancelFunc

	dbCity    *geoip2.Reader
	dbCountry *geoip2.Reader
	dbASN     *geoip2.Reader
}

func NewHandler(cregistry *component.Registry) (*Handler, error) {
	h := &Handler{
		cRegistry: cregistry,
	}

	return h, nil
}

func (h *Handler) Start(config Config) error {
	h.config = config

	logrus, err := component.Logrus(h.cRegistry)
	if err != nil {
		return err
	}
	h.logrus = logrus.Logger()

	refreshContext, cancel := context.WithCancel(context.Background())
	go func(ctx context.Context) {
		ticker := time.NewTicker(time.Duration(h.config.RefreshSeconds) * time.Second)
		client, err := maxmind.New(maxmind.Config{
			Logrus:            h.logrus,
			DownloadDirectory: h.config.DataDirectory,
			LicenseKey:        h.config.LicenseKey,
		})
		if err != nil {
			h.logrus.Error(err)
			return
		}
		for {
			downloader, err := client.NewDownloader(maxmind.EIDGeoLite2Country)
			if err != nil {
				h.logrus.Error(err)
				return
			}
			_, err = downloader.Download()
			if err != nil {
				h.logrus.Error(err)
			} else {
				db, err := geoip2.Open(filepath.Join(h.config.DataDirectory, "GeoLite2-Country.mmdb"))
				if err != nil {
					h.logrus.Error(err)
				} else {
					h.Lock()
					h.dbCountry = db
					h.Unlock()
				}
			}

			downloader, err = client.NewDownloader(maxmind.EIDGeoLite2City)
			if err != nil {
				h.logrus.Error(err)
				return
			}
			_, err = downloader.Download()
			if err != nil {
				h.logrus.Error(err)
				continue
			} else {
				db, err := geoip2.Open(filepath.Join(h.config.DataDirectory, "GeoLite2-City.mmdb"))
				if err != nil {
					h.logrus.Error(err)
				} else {
					h.Lock()
					h.dbCity = db
					h.Unlock()
				}
			}

			downloader, err = client.NewDownloader(maxmind.EIDGeoLite2ASN)
			if err != nil {
				h.logrus.Error(err)
				return
			}
			_, err = downloader.Download()
			if err != nil {
				h.logrus.Error(err)
				continue
			} else {
				db, err := geoip2.Open(filepath.Join(h.config.DataDirectory, "GeoLite2-ASN.mmdb"))
				if err != nil {
					h.logrus.Error(err)
				} else {
					h.Lock()
					h.dbASN = db
					h.Unlock()
				}
			}

			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				continue
			}
		}
	}(refreshContext)

	h.refreshCancel = cancel

	return nil
}
func (h *Handler) Stop() error {
	h.refreshCancel()

	h.Lock()
	if h.dbCountry != nil {
		if err := h.dbCountry.Close(); err != nil {
			logrus.Error(err)
		}
	}
	if h.dbCity != nil {
		if err := h.dbCity.Close(); err != nil {
			logrus.Error(err)
		}
	}
	if h.dbASN != nil {
		if err := h.dbASN.Close(); err != nil {
			logrus.Error(err)
		}
	}
	h.Unlock()

	return nil
}

func (h *Handler) Country(ctx context.Context, in *geoippb.IpRequest, out *geoippb.CountryResponse) error {
	ip := net.ParseIP(in.Ip)
	if ip == nil {
		return errors.BadRequest("INVALID_IP", "invalid ip '%s' given", in.Ip)
	}

	h.RLock()
	if h.dbCountry == nil {
		h.RUnlock()
		return errors.InternalServerError("DB_NOT_READY_YET", "maxmind country database is not ready yet")
	}
	record, err := h.dbCountry.Country(ip)
	h.RUnlock()

	if err != nil {
		logrus.Error(err)
	}

	out.IsoCode = record.Country.IsoCode

	out.CountryName = make(map[string]string)
	for _, lang := range in.Languages {
		if n, ok := record.Country.Names[lang]; ok {
			out.CountryName[lang] = n
		}
	}

	if in.CommaLanguages != "" {
		for _, lang := range strings.Split(in.CommaLanguages, ",") {
			if n, ok := record.Country.Names[lang]; ok {
				out.CountryName[lang] = n
			}
		}
	}

	return nil
}

func (h *Handler) City(ctx context.Context, in *geoippb.IpRequest, out *geoippb.CityResponse) error {
	ip := net.ParseIP(in.Ip)
	if ip == nil {
		return errors.BadRequest("INVALID_IP", "invalid ip '%s' given", in.Ip)
	}

	h.RLock()
	if h.dbCity == nil {
		h.RUnlock()
		return errors.InternalServerError("DB_NOT_READY_YET", "maxmind country database is not ready yet")
	}
	record, err := h.dbCity.City(ip)
	h.RUnlock()

	if err != nil {
		logrus.Error(err)
	}

	out.IsoCode = record.Country.IsoCode
	out.TimeZone = record.Location.TimeZone
	out.Latitude = record.Location.Latitude
	out.Longitude = record.Location.Longitude

	out.CountryName = make(map[string]string)
	for _, lang := range in.Languages {
		if n, ok := record.Country.Names[lang]; ok {
			out.CountryName[lang] = n
		}
	}

	if in.CommaLanguages != "" {
		for _, lang := range strings.Split(in.CommaLanguages, ",") {
			if n, ok := record.Country.Names[lang]; ok {
				out.CountryName[lang] = n
			}
		}
	}

	out.CityName = make(map[string]string)
	for _, lang := range in.Languages {
		if n, ok := record.City.Names[lang]; ok {
			out.CityName[lang] = n
		}
	}

	if in.CommaLanguages != "" {
		for _, lang := range strings.Split(in.CommaLanguages, ",") {
			if n, ok := record.City.Names[lang]; ok {
				out.CityName[lang] = n
			}
		}
	}

	return nil
}
