package maxmind

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/docker/go-units"
	"github.com/mholt/archiver/v3"
	"github.com/pkg/errors"
)

// Downloader represents an active downloader object
type Downloader struct {
	*Client
	eid   EditionID
	dlDir string
}

// Download downloads a database
func (d *Downloader) Download() ([]os.FileInfo, error) {
	// Retrieve expected hash
	expHash, err := d.expectedHash()
	if err != nil {
		return nil, errors.Wrap(err, "Cannot get archive checksum")
	}

	// Download DB archive
	archive := path.Join(d.workDir, d.eid.Filename())
	if err := d.downloadArchive(expHash, archive); err != nil {
		return nil, err
	}

	// Create checksum file
	checksumFile := path.Join(d.workDir, fmt.Sprintf(".%s.%s", d.eid.Filename(), "sha256"))
	if err := createFile(checksumFile, expHash); err != nil {
		return nil, errors.Errorf("Cannot create checksum file %s", checksumFile)
	}

	// Extract DB from archive
	dbs, err := d.extractArchive(archive)
	if err != nil {
		return nil, err
	}

	return dbs, nil
}

func (d *Downloader) expectedHash() (string, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/app/geoip_download", d.baseURL), nil)
	if err != nil {
		return "", errors.Wrap(err, "Request failed")
	}

	q := req.URL.Query()
	q.Add("license_key", d.licenseKey)
	q.Add("edition_id", d.eid.String())
	q.Add("suffix", fmt.Sprintf("%s.sha256", d.eid.Suffix().String()))
	req.URL.RawQuery = q.Encode()

	res, err := d.http.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", errors.Errorf("Received invalid status code %d", res.StatusCode)
	}

	checksum, err := io.ReadAll(res.Body)
	if err != nil {
		return "", errors.Wrap(err, "Cannot download checksum file")
	}

	checksumAr := strings.SplitN(strings.TrimSpace(string(checksum)), " ", 2)
	if len(checksumAr[0]) != 64 {
		return "", errors.Errorf("Invalid checksum: %s", checksum)
	}

	return checksumAr[0], nil
}

func (d *Downloader) downloadArchive(expHash string, archive string) error {
	if _, err := os.Stat(archive); err == nil {
		curHash, err := checksumFromFile(archive)
		if err != nil {
			return errors.Wrap(err, "Cannot get archive checksum")
		}
		if expHash == curHash {
			d.logrus.
				WithField("edition_id", d.eid.String()).
				WithField("hash", expHash).
				Debug("Archive already downloaded and valid. Skipping download")
			return nil
		}
	}

	d.logrus.
		WithField("edition_id", d.eid.String()).
		WithField("archive", filepath.Base(archive)).
		Info("Downloading archive...")

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/app/geoip_download", d.baseURL), nil)
	if err != nil {
		return errors.Wrap(err, "Request failed")
	}

	q := req.URL.Query()
	q.Add("license_key", d.licenseKey)
	q.Add("edition_id", d.eid.String())
	q.Add("suffix", d.eid.Suffix().String())
	req.URL.RawQuery = q.Encode()

	res, err := d.http.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return errors.Errorf("Received invalid status code %d", res.StatusCode)
	}

	out, err := os.Create(archive)
	if err != nil {
		return errors.Wrap(err, "Cannot create archive file")
	}
	defer out.Close()

	_, err = io.Copy(out, res.Body)
	if err != nil {
		return errors.Wrap(err, "Cannot download archive")
	}

	curHash, err := checksumFromFile(archive)
	if err != nil {
		return errors.Wrap(err, "Cannot get archive checksum")
	}

	if expHash != curHash {
		return errors.Errorf("Checksum of downloaded archive (%s) does not match the expected one (%s)", curHash, expHash)
	}

	return nil
}

func (d *Downloader) extractArchive(archive string) ([]os.FileInfo, error) {
	var dbs []os.FileInfo
	err := archiver.Walk(archive, func(f archiver.File) error {
		if f.IsDir() {
			return nil
		}
		if filepath.Ext(f.Name()) != ".csv" && filepath.Ext(f.Name()) != ".mmdb" {
			return nil
		}

		expHash, reader, err := checksumFromReader(f)
		if err != nil {
			return err
		}

		sublog := d.logrus.
			WithField("edition_id", d.eid.String()).
			WithField("db_name", f.Name()).
			WithField("db_size", units.HumanSize(float64(f.Size()))).
			WithField("db_modtime", f.ModTime()).
			WithField("db_hash", expHash)

		dbpath := path.Join(d.dlDir, f.Name())
		if fileExists(dbpath) {
			curHash, err := checksumFromFile(dbpath)
			if err != nil {
				return err
			}
			if expHash == curHash {
				sublog.Debug("Database is already up to date")
				return nil
			}
		}

		sublog.Debug("Extracting database")
		dbfile, err := os.Create(dbpath)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("Cannot create database file %s", f.Name()))
		}
		defer dbfile.Close()

		_, err = io.Copy(dbfile, reader)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("Cannot extract database file %s", f.Name()))
		}

		if err = os.Chtimes(dbpath, f.ModTime(), f.ModTime()); err != nil {
			sublog.WithField("error", err).Warn("Cannot preserve modtime of database file")
		}

		dbs = append(dbs, f)
		return nil
	})

	return dbs, err
}
