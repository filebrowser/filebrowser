package file

import (
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/mholt/archiver"
)

func (i *Info) DownloadAs(w http.ResponseWriter, query string) (int, error) {
	var (
		extension string
		temp      string
		err       error
	)

	temp, err = ioutil.TempDir("", "")
	if err != nil {
		return http.StatusInternalServerError, err
	}

	switch query {
	case "zip":
		extension, err = ".zip", archiver.Zip.Make(temp+"/temp", []string{i.Path})
	case "tar":
		extension, err = ".tar", archiver.Tar.Make(temp+"/temp", []string{i.Path})
	case "targz":
		extension, err = ".tar.gz", archiver.TarGz.Make(temp+"/temp", []string{i.Path})
	case "tarbz2":
		extension, err = ".tar.bz2", archiver.TarBz2.Make(temp+"/temp", []string{i.Path})
	default:
		return http.StatusNotImplemented, nil
	}

	if err != nil {
		return http.StatusInternalServerError, err
	}

	file, err := os.Open(temp + "/temp")
	if err != nil {
		return http.StatusInternalServerError, err
	}

	w.Header().Set("Content-Disposition", "attachment; filename="+i.Name()+extension)
	io.Copy(w, file)
	return http.StatusOK, nil
}
