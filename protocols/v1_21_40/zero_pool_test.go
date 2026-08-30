package v1_21_40

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func TestProtocol748ZeroValuePacketPools(t *testing.T) {
	data, err := os.ReadFile("testdata/zero_pool.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Server map[uint32]string `json:"server"`
		Client map[uint32]string `json:"client"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	for _, direction := range []struct {
		name     string
		listener bool
		want     map[uint32]string
	}{{"server", false, fixture.Server}, {"client", true, fixture.Client}} {
		t.Run(direction.name, func(t *testing.T) {
			pool := (Protocol{}).Packets(direction.listener)
			for id, encoded := range direction.want {
				// These zero values are not semantic equivalents across the
				// current model. Populated fixtures cover their normalised forms.
				switch id {
				case packet.IDResourcePackClientResponse, packet.IDStartGame, packet.IDPlayerSkin,
					packet.IDBiomeDefinitionList, packet.IDLevelSoundEvent, packet.IDClientBoundDebugRenderer:
					continue
				}
				constructor, ok := pool[id]
				if !ok {
					t.Errorf("historical packet %d is missing", id)
					continue
				}
				want, err := hex.DecodeString(encoded)
				if err != nil {
					t.Fatal(err)
				}
				var buffer bytes.Buffer
				func() {
					defer func() {
						if recovered := recover(); recovered != nil {
							t.Errorf("packet %d panicked: %v", id, recovered)
						}
					}()
					constructor().Marshal((Protocol{}).NewWriter(&buffer, -1))
				}()
				if got := buffer.Bytes(); !bytes.Equal(got, want) {
					t.Errorf("packet %d (%T): got %x, want %x", id, constructor(), got, want)
				}
			}
		})
	}
}
