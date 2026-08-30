// Command snapshot_legacy_families copies immutable registry blobs for the
// protocol 766, 748, and 475 adapters from locked Dragonfly/gophertunnel Git
// objects. Git blob IDs and sizes are checked before any output is accepted.
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

type snapshot struct {
	Source string
	Output string
	Blob   string
	Size   int
}

type family struct {
	Commit   string
	Output   string
	Files    []snapshot
	BiomeGit string
}

var families = map[string]family{
	"v766": {
		Commit: "800ee17878ab055a1410228b6a2dde4732d04192",
		Output: "data/v766",
		Files: []snapshot{
			{Source: "server/world/block_states.nbt", Output: "block_states.nbt", Blob: "556c9edfbf9247f3bfabb11a56a264e0bbbcb4c1", Size: 2059051},
			{Source: "server/world/item_runtime_ids.nbt", Output: "item_runtime_ids.nbt", Blob: "10fa3de3f89659d168e340160bbf1c5a966fc807", Size: 52097},
			{Source: "server/item/creative/creative_items.nbt", Output: "creative_items.nbt", Blob: "f6f96a370823063ba7ced42ad50e764193bf12ea", Size: 117886},
			{Source: "server/item/recipe/crafting_data.nbt", Output: "crafting_data.nbt", Blob: "c3809b5cc604e17caa6fcfab8d50e784c591247b", Size: 795973},
			{Source: "server/item/recipe/furnace_data.nbt", Output: "furnace_data.nbt", Blob: "dbaa9ca744a278b66018bcf2ff947ba0da03653b", Size: 32386},
			{Source: "server/item/recipe/potion_data.nbt", Output: "potion_data.nbt", Blob: "9f419a696dbd21b875265e4d2746ecb6b6c1efa4", Size: 33442},
			{Source: "server/item/recipe/smithing_data.nbt", Output: "smithing_data.nbt", Blob: "362d1e7903f2b84c97501f1b1dbddf475b5d2649", Size: 2351},
		},
		BiomeGit: "ecff04b74c9d14ae88a284bc45fc9bb180b79853",
	},
	"v748": {
		Commit: "206bf97c6c88f29ee7006c0e1339491cb3752aa6",
		Output: "data/v748",
		Files: []snapshot{
			{Source: "server/world/block_states.nbt", Output: "block_states.nbt", Blob: "ff4d75701640921818a7fda13f0449f5e3595a19", Size: 1918591},
			{Source: "server/world/item_runtime_ids.nbt", Output: "item_runtime_ids.nbt", Blob: "e1eb29e8491d10cabdb133424de52201b12f7115", Size: 50902},
			{Source: "server/item/creative/creative_items.nbt", Output: "creative_items.nbt", Blob: "82f3b87f080c43007317b2070193c5dc248fa5ac", Size: 115280},
			{Source: "server/item/recipe/crafting_data.nbt", Output: "crafting_data.nbt", Blob: "72562b025995de0a98acd3667fcfb9ae79ad3ffd", Size: 780535},
			{Source: "server/item/recipe/furnace_data.nbt", Output: "furnace_data.nbt", Blob: "165f16c112c18efbdac7be46b4e17cb2233d0630", Size: 31755},
			{Source: "server/item/recipe/potion_data.nbt", Output: "potion_data.nbt", Blob: "42d7951ae83a0e9badf8e8d01b0ed268b8e3463a", Size: 33442},
			{Source: "server/item/recipe/smithing_data.nbt", Output: "smithing_data.nbt", Blob: "362d1e7903f2b84c97501f1b1dbddf475b5d2649", Size: 2351},
		},
		BiomeGit: "268adeb5710990ebd95edda3b42b0cf13f58bf8b",
	},
	"v475": {
		Commit: "5ac88dcd93d5ae79fca6142d9cd326cb23015d2f",
		Output: "data/v475",
		Files: []snapshot{
			{Source: "server/world/block_states.nbt", Output: "block_states.nbt", Blob: "e92bf912da0e579d399f3d464a56550c36bfac59", Size: 1174943},
			{Source: "server/world/item_runtime_ids.nbt", Output: "item_runtime_ids.nbt", Blob: "4af29ae15f78535d2f2ae745c04c3f60f84eac42", Size: 29729},
			{Source: "server/item/creative/creative_items.nbt", Output: "creative_items.nbt", Blob: "35d41cdec4e5109a6adfd26f47a7eab6e861eead", Size: 133971},
			{Source: "server/world/chunk/legacy_states.nbt", Output: "legacy_states.nbt", Blob: "c38d55d35b6f6709f4d0a12304c08c16cb78e9f3", Size: 399275},
		},
		BiomeGit: "c40bf8288fb93ebd014375e46e4592424153927c",
	},
}

func main() {
	var dragonfly, gophertunnel, root string
	flag.StringVar(&dragonfly, "dragonfly", "", "path to the df-mc/dragonfly Git checkout")
	flag.StringVar(&gophertunnel, "gophertunnel", "", "path to the Sandertv/gophertunnel Git checkout")
	flag.StringVar(&root, "out", ".", "go-multiversion repository root")
	flag.Parse()
	if dragonfly == "" || gophertunnel == "" {
		flag.Usage()
		os.Exit(2)
	}
	for name, target := range families {
		output := filepath.Join(root, target.Output)
		if err := os.MkdirAll(output, 0o755); err != nil {
			fatal(err)
		}
		for _, file := range target.Files {
			data := gitObject(dragonfly, target.Commit+":"+file.Source)
			if len(data) != file.Size || blobID(data) != file.Blob {
				fatal(fmt.Errorf("%s %s: got blob %s/%d, want %s/%d", name, file.Source, blobID(data), len(data), file.Blob, file.Size))
			}
			write(filepath.Join(output, file.Output), data)
		}
		conn := gitObject(gophertunnel, target.BiomeGit+":minecraft/conn.go")
		re := regexp.MustCompile("(?s)client crashes when not sending all biomes.*?const s = `([A-Za-z0-9+/=]+)`")
		match := re.FindSubmatch(conn)
		if len(match) != 2 {
			fatal(fmt.Errorf("%s default biome definitions were not found", name))
		}
		biomes, err := base64.StdEncoding.DecodeString(string(match[1]))
		if err != nil {
			fatal(fmt.Errorf("%s decode biome definitions: %w", name, err))
		}
		write(filepath.Join(output, "biome_definitions.nbt"), biomes)
	}
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
