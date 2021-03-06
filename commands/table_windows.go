package commands

import (
	"context"
	"io"

	"golang.org/x/sys/windows"

	"github.com/zetamatta/go-texts/mbcs"

	"github.com/zetamatta/nyagos/dos"
)

func init() {
	buildInCommand = map[string]func(context.Context, Param) (int, error){
		".":        cmdSource,
		"alias":    cmdAlias,
		"attrib":   cmdAttrib,
		"bindkey":  cmdBindkey,
		"box":      cmdBox,
		"cd":       cmdCd,
		"clip":     cmdClip,
		"clone":    cmdClone,
		"cls":      cmdCls,
		"chmod":    cmdChmod,
		"copy":     cmdCopy,
		"del":      cmdDel,
		"dirs":     cmdDirs,
		"diskfree": cmdDiskFree,
		"diskused": cmdDiskUsed,
		"echo":     cmdEcho,
		"env":      cmdEnv,
		"erase":    cmdDel,
		"exit":     cmdExit,
		"foreach":  cmdForeach,
		"history":  cmdHistory,
		"if":       cmdIf,
		"ln":       cmdLn,
		"lnk":      cmdLnk,
		"kill":     cmdKill,
		"killall":  cmdKillAll,
		"ls":       cmdLs,
		"md":       cmdMkdir,
		"mkdir":    cmdMkdir,
		"more":     cmdMore,
		"move":     cmdMove,
		"open":     cmdOpen,
		"popd":     cmdPopd,
		"ps":       cmdPs,
		"pushd":    cmdPushd,
		"pwd":      cmdPwd,
		"rd":       cmdRmdir,
		"rem":      cmdRem,
		"rmdir":    cmdRmdir,
		"set":      cmdSet,
		"source":   cmdSource,
		"su":       cmdSu,
		"touch":    cmdTouch,
		"type":     cmdType,
		"which":    cmdWhich,
	}
}

func newMbcsReader(r io.Reader) io.Reader {
	return mbcs.NewAutoDetectReader(r, mbcs.ConsoleCP())
}

func readShortCut(dir string) (string, error) {
	newdir, _, err := dos.ReadShortcut(dir)
	return newdir, err
}

func setWritable(path string) error {
	perm, err := dos.GetFileAttributes(path)
	if err != nil {
		return err
	}
	return dos.SetFileAttributes(path, perm&^windows.FILE_ATTRIBUTE_READONLY)
}

func truncate(path string, f func(path string, err error) bool, out io.Writer) error {
	return dos.Truncate(path, f, out)
}
