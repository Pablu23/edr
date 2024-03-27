package main

import (
	"bytes"
	"crypto/sha256"
	"edr/pkg/common"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"golang.org/x/sys/windows"
	"io"
	"net/http"
	"os"
	"reflect"
	"time"
)

func main() {
	//fExists := true
	//_, err = os.Stat("running.txt")
	//if err != nil && errors.Is(err, os.ErrNotExist) {
	//	fExists = false
	//} else if err != nil {
	//	panic(err)
	//}
	//var file *os.File
	//if !fExists {
	//	file, _ = os.Create("running.txt")
	//	defer file.Close()
	//}

	cPid := windows.GetCurrentProcessId()

	clientId := uuid.New()
	for {
		err, p := getPidAndPath()
		if err != nil {
			panic(err)
		}

		r := make([]common.Proc, 0)
		for _, proc := range p {
			if proc.Err != nil {
				fmt.Printf("PID: %d Error: %s\n", proc.PID, proc.Err)
			} else {
				hash, err := getHash(proc)
				if err != nil {
					panic(err)
				}

				//if !fExists {
				//	_, err = file.WriteString(fmt.Sprintf("%s;%s\n", proc.Path, hex.EncodeToString(hash)))
				//	if err != nil {
				//		panic(err)
				//	}
				//}

				r = append(r, common.Proc{
					ExePath: proc.Path,
					HashHex: hex.EncodeToString(hash),
					PID:     proc.PID,
				})
			}
		}

		info := common.ClientInfo{
			ID:      clientId,
			Running: r,
		}
		j, err := json.Marshal(info)
		if err != nil {
			panic(err)
		}
		post, err := http.Post("http://localhost:8080/client/", "application/json", bytes.NewBuffer(j))
		if err != nil {
			panic(err)
		}
		var k []uint32
		err = json.NewDecoder(post.Body).Decode(&k)
		if err != nil {
			panic(err)
		}
		//fmt.Printf("Status: %s\n", post.Status)
		for _, u := range k {
			fmt.Printf("Terminating PID: %d\n", u)
			if u == cPid {
				fmt.Printf("Skippping self")
				continue
			}

			err := Terminate(u)
			if err != nil {
				fmt.Printf("Could not terminate: PID %d, because error: %s\n", u, err)
			}
		}

		time.Sleep(5 * time.Second)
	}
}

type Proc struct {
	PID  uint32
	Path string
	Err  error
}

func Terminate(pid uint32) error {
	handle, err := windows.OpenProcess(windows.PROCESS_TERMINATE, false, pid)
	if err != nil {
		return err
	}
	return windows.TerminateProcess(handle, 1337)
}

func getHash(p Proc) ([]byte, error) {
	file, err := os.Open(p.Path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	h := sha256.New()
	if _, err := io.Copy(h, file); err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
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
