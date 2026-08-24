// Command snapshot_v486 copies the immutable Minecraft 1.18.1x registry blobs
// from the locked Dragonfly and gophertunnel Git objects into data/v486.
package main

import (
	"bytes"
	"crypto/sha1" // #nosec G505 -- Git object IDs intentionally use SHA-1.
	"encoding/base64"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
)

const (
	dragonflyCommit    = "677c8fa1753d662e115e9c610dfe334daed2d265"
	gophertunnelCommit = "2cb1e399e53928529916dd2bda1efd11a0ac374d"
)

type snapshot struct {
	Source string
	Output string
	Blob   string
	Size   int
}

var snapshots = []snapshot{
	{Source: "server/world/block_states.nbt", Output: "block_states.nbt", Blob: "df8c6f77b8070acaa3514e6f5749ad6e38f3c09e", Size: 1174608},
	{Source: "server/world/item_runtime_ids.nbt", Output: "item_runtime_ids.nbt", Blob: "e58e154fd2060c590b9aa3ca5688a337431ffb0c", Size: 30027},
	{Source: "server/item/creative/creative_items.nbt", Output: "creative_items.nbt", Blob: "4d5be4ed21f01c61174cf41d577794c15674aca8", Size: 134053},
	{Source: "server/world/chunk/legacy_states.nbt", Output: "legacy_states.nbt", Blob: "c38d55d35b6f6709f4d0a12304c08c16cb78e9f3", Size: 399275},
}

func main() {
	var dragonfly, gophertunnel, output string
	flag.StringVar(&dragonfly, "dragonfly", "", "path to the df-mc/dragonfly Git checkout")
	flag.StringVar(&gophertunnel, "gophertunnel", "", "path to the Sandertv/gophertunnel Git checkout")
	flag.StringVar(&output, "out", "data/v486", "output directory")
	flag.Parse()
	if dragonfly == "" || gophertunnel == "" {
		flag.Usage()
		os.Exit(2)
	}
	if err := os.MkdirAll(output, 0o755); err != nil {
		fatal(err)
	}
	for _, file := range snapshots {
		data := gitObject(dragonfly, dragonflyCommit+":"+file.Source)
		if len(data) != file.Size {
			fatal(fmt.Errorf("%s size: got %d, want %d", file.Source, len(data), file.Size))
		}
		if blobID(data) != file.Blob {
			fatal(fmt.Errorf("%s Git blob: got %s, want %s", file.Source, blobID(data), file.Blob))
		}
		write(filepath.Join(output, file.Output), data)
	}

	conn := gitObject(gophertunnel, gophertunnelCommit+":minecraft/conn.go")
	re := regexp.MustCompile(`(?s)client crashes when not sending all biomes.*?const s = \x60([A-Za-z0-9+/=]+)\x60`)
	match := re.FindSubmatch(conn)
	if len(match) != 2 {
		fatal(fmt.Errorf("default protocol-486 biome definitions were not found"))
	}
	biomes, err := base64.StdEncoding.DecodeString(string(match[1]))
	if err != nil {
		fatal(fmt.Errorf("decode biome definitions: %w", err))
	}
	write(filepath.Join(output, "biome_definitions.nbt"), biomes)
}

func gitObject(repository, object string) []byte {
	cmd := exec.Command("git", "-C", repository, "show", object)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	data, err := cmd.Output()
	if err != nil {
		fatal(fmt.Errorf("git show %s: %w: %s", object, err, stderr.String()))
	}
	return data
}

func blobID(data []byte) string {
	h := sha1.New() // #nosec G401 -- required to verify Git's immutable blob ID.
	_, _ = fmt.Fprintf(h, "blob %d%c", len(data), byte(0))
	_, _ = h.Write(data)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func write(path string, data []byte) {
	if err := os.WriteFile(path, data, 0o644); err != nil {
		fatal(err)
	}
}

func fatal(err error) {
	_, _ = fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
