package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

type output struct {
	Server  map[uint32]string `json:"server"`
	Client  map[uint32]string `json:"client"`
	Skipped map[string]string `json:"skipped,omitempty"`
}

func main() {
	path := flag.String("out", "", "output JSON path")
	flag.Parse()
	if *path == "" && flag.NArg() == 3 {
		roundTrip(flag.Arg(0), flag.Arg(1), flag.Arg(2))
		return
	}
	if *path == "" {
		flag.Usage()
		os.Exit(2)
	}
	result := output{Server: map[uint32]string{}, Client: map[uint32]string{}, Skipped: map[string]string{}}
	encodePool("server", packet.NewServerPool(), result.Server, result.Skipped)
	encodePool("client", packet.NewClientPool(), result.Client, result.Skipped)
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		panic(err)
	}
	data = append(data, '\n')
	if err := os.MkdirAll(filepath.Dir(*path), 0o755); err != nil {
		panic(err)
	}
	if err := os.WriteFile(*path, data, 0o644); err != nil {
		panic(err)
	}
}

func roundTrip(direction, rawID, encoded string) {
	id, err := strconv.ParseUint(rawID, 10, 32)
	if err != nil {
		panic(err)
	}
	data, err := hex.DecodeString(encoded)
	if err != nil {
		panic(err)
	}
	pool := packet.NewServerPool()
	if direction == "client" {
		pool = packet.NewClientPool()
	}
	constructor, ok := pool[uint32(id)]
	if !ok {
		panic(fmt.Sprintf("packet %d is missing", id))
	}
	pk := constructor()
	pk.Marshal(protocol.NewReader(zeroSafeReader{Reader: bytes.NewReader(data)}, -1, true))
	var buffer bytes.Buffer
	pk.Marshal(protocol.NewWriter(&buffer, -1))
	_, _ = fmt.Fprint(os.Stdout, hex.EncodeToString(buffer.Bytes()))
}

type zeroSafeReader struct{ *bytes.Reader }

func (r zeroSafeReader) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	return r.Reader.Read(p)
}

func encodePool(direction string, pool packet.Pool, encoded map[uint32]string, skipped map[string]string) {
	ids := make([]int, 0, len(pool))
	for id := range pool {
		ids = append(ids, int(id))
	}
	sort.Ints(ids)
	for _, rawID := range ids {
		id := uint32(rawID)
		func() {
			defer func() {
				if recovered := recover(); recovered != nil {
					skipped[fmt.Sprintf("%s/%d", direction, id)] = fmt.Sprint(recovered)
				}
			}()
			var buffer bytes.Buffer
			pool[id]().Marshal(protocol.NewWriter(&buffer, -1))
			encoded[id] = hex.EncodeToString(buffer.Bytes())
		}()
	}
}
