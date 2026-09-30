package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
)

// dropPrivileges runs when the container starts as root with RUN_AS_UID set.
// It gives the data directory to that user and then permanently switches to
// it, before any request is served. Volumes written by older root-run
// versions are therefore fixed automatically, with no manual chown.
func dropPrivileges(dbPath string) error {
	uidS := os.Getenv("RUN_AS_UID")
	if uidS == "" || os.Getuid() != 0 {
		return nil
	}
	uid, err := strconv.Atoi(uidS)
	if err != nil || uid <= 0 {
		return fmt.Errorf("invalid RUN_AS_UID %q", uidS)
	}
	gid := uid

	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("create %s: %w", dir, err)
	}
	fixed := 0
	err = filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if st, ok := info.Sys().(*syscall.Stat_t); ok && int(st.Uid) == uid && int(st.Gid) == gid {
			return nil
		}
		fixed++
		return os.Lchown(p, uid, gid)
	})
	if err != nil {
		return fmt.Errorf("chown %s: %w", dir, err)
	}
	if fixed > 0 {
		log.Printf("gave %d file(s) in %s to uid %d", fixed, dir, uid)
	}

	// Order matters: supplementary groups and gid must change while still root.
	if err := syscall.Setgroups([]int{}); err != nil {
		return fmt.Errorf("setgroups: %w", err)
	}
	if err := syscall.Setgid(gid); err != nil {
		return fmt.Errorf("setgid: %w", err)
	}
	if err := syscall.Setuid(uid); err != nil {
		return fmt.Errorf("setuid: %w", err)
	}
	return nil
}
