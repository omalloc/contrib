package runtime

import (
	"fmt"
	"os/user"
	"strconv"
	"syscall"
)

func SetCurrentUser(username string) error {
	if username != "" {
		current, err := user.Current()
		if err != nil {
			return err
		}

		if current.Username == username {
			return nil
		}

		wantedUser, err := user.Lookup(username)
		if err != nil {
			return err
		}

		uid, err := strconv.Atoi(wantedUser.Uid)
		if err != nil {
			return fmt.Errorf("error converting UID [%s] to int: %s", wantedUser.Uid, err)
		}

		gid, err := strconv.Atoi(wantedUser.Gid)
		if err != nil {
			return fmt.Errorf("error converting GID [%s] to int: %s", wantedUser.Gid, err)
		}

		if err = syscall.Setgid(gid); err != nil {
			return fmt.Errorf("setting group id: %s", err)
		}

		if err = syscall.Setuid(uid); err != nil {
			return fmt.Errorf("setting user id: %s", err)
		}
	}

	return nil
}
