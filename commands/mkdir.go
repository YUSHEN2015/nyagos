package commands

import (
	"context"
	"fmt"
	"os"
	"syscall"
)

func cmdMkdir(ctx context.Context, cmd Param) (int, error) {
	if len(cmd.Args()) <= 1 {
		fmt.Fprintln(cmd.Err(), "Usage: mkdir [/p] DIRECTORIES...")
		return 0, nil
	}
	errorcount := 0
	mkdir := os.Mkdir
	for _, arg1 := range cmd.Args()[1:] {
		if arg1 == "/p" {
			mkdir = os.MkdirAll
			continue
		}
		err := mkdir(arg1, 0777)
		if err != nil {
			fmt.Fprintf(cmd.Err(), "%s: %s\n", arg1, err)
			errorcount++
		}
	}
	return errorcount, nil
}

func cmdRmdir(ctx context.Context, cmd Param) (int, error) {
	if len(cmd.Args()) <= 1 {
		fmt.Fprintln(cmd.Err(), "Usage: rmdir [/s] [/q] DIRECTORIES...")
		return 0, nil
	}
	sOption := false
	quiet := false
	message := "%s: Rmdir Are you sure? [Yes/No/Quit] "
	errorcount := 0
	for _, arg1 := range cmd.Args()[1:] {
		switch arg1 {
		case "/s":
			sOption = true
			message = "%s : Delete Tree. Are you sure? [Yes/No/Quit] "
			continue
		case "/q":
			quiet = true
			continue
		}
		stat, err := os.Lstat(arg1)
		if err != nil {
			fmt.Fprintf(cmd.Err(), "%s: %s\n", arg1, err)
			errorcount++
			continue
		}
		if !quiet {
			fmt.Fprintf(cmd.Err(), message, arg1)
			ch, err := getkey()
			if err != nil {
				return 1, err
			}
			fmt.Fprintf(cmd.Err(), "%c ", ch)
			switch ch {
			case 'y', 'Y':

			case 'q', 'Q':
				fmt.Fprintln(cmd.Err(), "-> canceled all")
				return errorcount, nil
			default:
				fmt.Fprintln(cmd.Err(), "-> canceled")
				continue
			}
		}
		if sOption {
			if !stat.IsDir() {
				fmt.Fprintf(cmd.Err(), "%s: not directory\n", arg1)
				errorcount++
				continue
			}
			if !quiet {
				fmt.Fprintln(cmd.Out())
			}
			err = truncate(arg1, func(path string, err error) bool {
				fmt.Fprintf(cmd.Err(), "%s -> %s\n", path, err)
				return true
			}, cmd.Out())
		} else {
			err = syscall.Rmdir(arg1)
		}
		if err != nil {
			fmt.Fprintf(cmd.Err(), "-> %s\n", err)
			errorcount++
		} else {
			fmt.Fprintln(cmd.Err(), "-> done.")
		}
	}
	return errorcount, nil
}
