package main

import (
	"fmt"

	"golang.org/x/sys/unix"
)

func main() {
	kprocs, err := unix.SysctlKinfoProcSlice("kern.proc.all")
	if err != nil {
		fmt.Println(err)
	}

	for _, proc := range kprocs {
		pid := proc.Proc.P_pid
		fmt.Println(pid)
	}
}
