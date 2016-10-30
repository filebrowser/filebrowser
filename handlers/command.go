package handlers

import (
	"bytes"
	"fmt"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/hacdias/caddy-filemanager/config"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Command handles the requests for VCS related commands: git, svn and mercurial
func Command(w http.ResponseWriter, r *http.Request, c *config.Config, u *config.User) (int, error) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return 0, nil
	}
	defer conn.Close()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("read:", err)
			break
		}

		command := strings.Split(string(message), " ")

		if len(command) == 0 {
			continue
		}
		// Check if the command is allowed
		mayContinue := false

		for _, cmd := range u.Commands {
			if cmd == command[0] {
				mayContinue = true
			}
		}

		if !mayContinue {
			err = conn.WriteMessage(websocket.BinaryMessage, []byte("FORBIDDEN"))
			if err != nil {
				fmt.Println("write:", err)
				break
			}

			return 0, nil
		}

		// Check if the program is talled is installed on the computer
		if _, err = exec.LookPath(command[0]); err != nil {
			err = conn.WriteMessage(websocket.BinaryMessage, []byte("Command not implemented."))
			if err != nil {
				fmt.Println("write:", err)
				break
			}

			return http.StatusNotImplemented, nil
		}

		path := strings.Replace(r.URL.Path, c.BaseURL, c.Scope, 1)
		path = filepath.Clean(path)

		buff := new(bytes.Buffer)

		cmd := exec.Command(command[0], command[1:len(command)]...)
		cmd.Dir = path
		cmd.Stderr = buff
		cmd.Stdout = buff
		err = cmd.Start()
		if err != nil {
			return http.StatusInternalServerError, err
		}

		done := false
		go func() {
			err = cmd.Wait()
			done = true
		}()

		for !done {
			by := buff.Bytes()
			if len(by) > 0 {
				err = conn.WriteMessage(websocket.TextMessage, by)
				if err != nil {
					fmt.Println("write:", err)
					break
				}
			}

			time.Sleep(100 * time.Millisecond)
		}

		by := buff.Bytes()
		if len(by) > 0 {
			err = conn.WriteMessage(websocket.TextMessage, by)
			if err != nil {
				fmt.Println("write:", err)
				break
			}
		}

		time.Sleep(100 * time.Millisecond)

		break
	}

	/* command := strings.Split(r.Header.Get("command"), " ")

	// Check if the command is allowed
	mayContinue := false

	for _, cmd := range u.Commands {
		if cmd == command[0] {
			mayContinue = true
		}
	}

	if !mayContinue {
		return http.StatusForbidden, nil
	}

	// Check if the program is talled is installed on the computer
	if _, err := exec.LookPath(command[0]); err != nil {
		return http.StatusNotImplemented, nil
	}

	path := strings.Replace(r.URL.Path, c.BaseURL, c.Scope, 1)
	path = filepath.Clean(path)

	cmd := exec.Command(command[0], command[1:len(command)]...)
	cmd.Dir = path
	cmd.Stderr = w
	cmd.Stdout = w
	cmd.Start()

	/*cmd.Stderr = b
	cmd.Stdout = b

	// Starts the comamnd
	err := cmd.Start()
	if err != nil {
		return http.StatusInternalServerError, err
	}

	done := false
	go func() {
		err = cmd.Wait()
		done = true
	}()

	for !done {
		by := b.Bytes()
		if len(by) > 0 {
			fmt.Println(string(by))
		}

		//w.Write(by)

	}*/

	//out, err := cmd.CombinedOutput()
	//fmt.Println(string(out))

	//if err != nil {
	//	return http.StatusInternalServerError, err
	//}

	/* cmd.Wait()

	//p := &page.Page{Info: &page.Info{Data: string(output)}} */
	return 0, nil
}
