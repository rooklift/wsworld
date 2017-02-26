package wsworld

import (
    "fmt"
    "io/ioutil"
    "net/http"
    "strings"
    "sync"

    "github.com/gorilla/websocket"
)

var player_count int
var player_count_mutex sync.Mutex

func new_player_id() int {
    player_count_mutex.Lock()
    defer player_count_mutex.Unlock()
    player_count++
    return player_count - 1
}

func ws_handler(writer http.ResponseWriter, request * http.Request) {

    fmt.Printf("Connection opened: %s\n", request.RemoteAddr)

    var upgrader = websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024, CheckOrigin: func(r *http.Request) bool {return true}}

    conn, err := upgrader.Upgrade(writer, request, nil)
    if err != nil {
        return
    }

    pid := new_player_id()

    eng.mutex.Lock()

    if eng.multiplayer == false {
        delete(eng.players, eng.latest_player)
    }

    keyboard := make(map[string]bool)
    eng.players[pid] = &player{pid, keyboard, conn}
    eng.latest_player = pid

    eng.mutex.Unlock()

    // Handle incoming messages until connection fails...

    for {
        _, reader, err := conn.NextReader()

        if err != nil {

            conn.Close()
            fmt.Printf("Connection CLOSED: %s (%v)\n", request.RemoteAddr, err)

            eng.mutex.Lock()
            delete(eng.players, pid)
            eng.mutex.Unlock()

            return
        }

        bytes, err := ioutil.ReadAll(reader)        // FIXME: this may be vulnerable to malicious huge messages

        fields := strings.Fields(string(bytes))

        switch fields[0] {

        case "keyup":

            eng.mutex.Lock()
            if eng.players[pid] != nil {
                eng.players[pid].keyboard[fields[1]] = false
            }
            eng.mutex.Unlock()

        case "keydown":

            eng.mutex.Lock()
            if eng.players[pid] != nil {
                eng.players[pid].keyboard[fields[1]] = true
            }
            eng.mutex.Unlock()
        }
    }
}
