package completion

import (
	"context"
	"os"
	"strings"

	"github.com/zetamatta/go-findfile"
	"github.com/zetamatta/nyagos/nodos"
)

const (
	STD_SLASH = string(os.PathSeparator)
	OPT_SLASH = "/"
)

var IncludeHidden = false

func listUpFiles(ctx context.Context, str string) ([]Element, error) {
	return listUpWithFilter(ctx, str, func(*findfile.FileInfo) bool { return true })
}
func listUpDirs(ctx context.Context, str string) ([]Element, error) {
	return listUpWithFilter(ctx, str, func(fd *findfile.FileInfo) bool {
		return fd.IsDir() || strings.HasSuffix(strings.ToLower(fd.Name()), ".lnk")
	})
}

func listUpWithFilter(ctx context.Context, str string, filter func(*findfile.FileInfo) bool) ([]Element, error) {
	orgSlash := STD_SLASH[0]
	if UseSlash {
		orgSlash = OPT_SLASH[0]
	}
	if pos := strings.IndexAny(str, STD_SLASH+OPT_SLASH); pos >= 0 {
		orgSlash = str[pos]
	}
	str = strings.Replace(strings.Replace(str, OPT_SLASH, STD_SLASH, -1), `"`, "", -1)
	directory := DirName(str)
	wildcard := join(findfile.ExpandEnv(directory), "*")

	// Drive letter
	cutprefix := 0
	if strings.HasPrefix(directory, STD_SLASH) {
		wd, _ := os.Getwd()
		directory = wd[0:2] + directory
		cutprefix = 2
	}
	commons := make([]Element, 0)
	STR := strings.ToUpper(str)
	var canceled error = nil
	fdErr := findfile.Walk(wildcard, func(fd *findfile.FileInfo) bool {
		if ctx != nil {
			select {
			case <-ctx.Done():
				canceled = ctx.Err()
				return false
			default:
			}
		}
		if fd.Name() == "." || fd.Name() == ".." {
			return true
		}
		if !IncludeHidden && fd.IsHidden() {
			return true
		}
		if !filter(fd) {
			return true
		}
		listname := fd.Name()
		name := join(directory, fd.Name())
		if fd.IsDir() {
			name += STD_SLASH
			listname += OPT_SLASH
		}
		if cutprefix > 0 {
			name = name[2:]
		}
		nameUpr := strings.ToUpper(name)
		if strings.HasPrefix(nameUpr, STR) {
			if orgSlash != STD_SLASH[0] {
				name = strings.Replace(name, STD_SLASH, OPT_SLASH, -1)
			}
			element := Element2{name, listname}
			commons = append(commons, element)
		}
		return true
	})
	if canceled != nil {
		return commons, canceled
	}
	return commons, fdErr
}

func join(dir, name string) string {
	return nodos.Join(dir, name)
}
