package utils

import (
	"io"
	"os"
	"regexp"
)

var re = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

func IsAlphanumeric(s string) bool {
	return re.MatchString(s)
}

func FileCopy(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()
	return
}
