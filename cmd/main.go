package main

import (
	"cmp"
	"fmt"
	"golang.org/x/sys/windows"
	"reflect"
	"slices"
	"strings"
)

func main() {
	err, p := getPidAndPath()
	if err != nil {
		panic(err)
	}

	slices.SortStableFunc(p, func(a, b Proc) int {
		return cmp.Compare(a.PID, b.PID)
	})

	for _, proc := range p {
		if proc.Err != nil {
			fmt.Printf("PID: %d Error: %s\n", proc.PID, proc.Err)
		} else {
			fmt.Printf("PID: %d Name: %s\n", proc.PID, proc.Path)
			if strings.Contains(proc.Path, "Taskmgr.exe") {
				err := proc.Terminate()
				if err != nil {
					fmt.Printf("ERROR CLOSING TASKMANAGER: %v\n", err)
				}
			}
		}
	}
}

type Proc struct {
	PID  uint32
	Path string
	Err  error
}

func (p *Proc) Terminate() error {
	handle, err := windows.OpenProcess(windows.PROCESS_TERMINATE, false, p.PID)
	if err != nil {
		return err
	}
	return windows.TerminateProcess(handle, 1337)
}

func getPidAndPath() (error, []Proc) {
	ids := make([]uint32, 500)
	var r uint32 = 0
	err := windows.EnumProcesses(ids, &r)
	if err != nil {
		return err, nil
	}

	result := make([]Proc, r/4)

	for i, pId := range ids[:r/4] {
		handle, err := windows.OpenProcess(windows.PROCESS_QUERY_INFORMATION|windows.PROCESS_VM_READ, false, pId)
		if err != nil {
			result[i] = Proc{PID: pId, Err: err}
			continue
		}
		var hMod windows.Handle
		var cbNeeded uint32
		err = windows.EnumProcessModules(handle, &hMod, uint32(reflect.TypeOf(hMod).Size()), &cbNeeded)
		if err != nil {
			result[i] = Proc{PID: pId, Err: err}
			err := windows.CloseHandle(handle)
			if err != nil {
				return err, nil
			}
			continue
		}
		var length uint32 = 200
		text := make([]uint16, length)
		err = windows.QueryFullProcessImageName(handle, 0, &text[0], &length)
		if err != nil {
			result[i] = Proc{PID: pId, Err: err}
			err := windows.CloseHandle(handle)
			if err != nil {
				return err, nil
			}
			continue
		}
		path := windows.UTF16ToString(text[0:length])
		result[i] = Proc{PID: pId, Err: err, Path: path}
		err = windows.CloseHandle(handle)
		if err != nil {
			return err, nil
		}

	}
	return nil, result
}
