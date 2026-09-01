// Package multiversion provides Minecraft Bedrock protocol adapters for
// gophertunnel-based servers.
package multiversion

import (
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/shawtymarco/go-multiversion/mapping"
	"github.com/shawtymarco/go-multiversion/protocols/v1_16_100"
	"github.com/shawtymarco/go-multiversion/protocols/v1_18_0"
	"github.com/shawtymarco/go-multiversion/protocols/v1_18_10"
	"github.com/shawtymarco/go-multiversion/protocols/v1_21_100"
	"github.com/shawtymarco/go-multiversion/protocols/v1_21_110"
	"github.com/shawtymarco/go-multiversion/protocols/v1_21_130"
	"github.com/shawtymarco/go-multiversion/protocols/v1_21_40"
	"github.com/shawtymarco/go-multiversion/protocols/v1_21_50"
	"github.com/shawtymarco/go-multiversion/protocols/v1_26_0"
	"github.com/shawtymarco/go-multiversion/protocols/v1_26_10"
	"github.com/shawtymarco/go-multiversion/protocols/v1_26_20"
	"github.com/shawtymarco/go-multiversion/protocols/v1_26_30"
	"github.com/shawtymarco/go-multiversion/protocols/v1_26_44"
)

// Protocols returns all non-native protocols supported by this module.
func Protocols() []minecraft.Protocol {
	return []minecraft.Protocol{v1_26_44.New()}
}

// V1_26_30 returns the wire-only Minecraft protocol 1001 family adapter for
// Minecraft 1.26.30 through 1.26.36. The adapter is not included in Protocols
// until its registry and chunk conversions are complete.
func V1_26_30() minecraft.Protocol {
	return v1_26_30.New()
}

// V1_26_20 returns the wire-only Minecraft protocol 975 family adapter for
// stable Minecraft 1.26.20, 1.26.21, and 1.26.23. Registry-aware consumers
// should use ProtocolsWithRegistries before advertising it.
func V1_26_20() minecraft.Protocol {
	return v1_26_20.New()
}

// V1_26_10 returns the wire-only Minecraft protocol 944 family adapter for
// stable Minecraft 1.26.10 through 1.26.14. Registry-aware consumers should
// use ProtocolsWithRegistries before advertising it.
func V1_26_10() minecraft.Protocol {
	return v1_26_10.New()
}

// V1_26_0 returns the wire-only Minecraft protocol 924 family adapter for
// stable Minecraft 1.26.0 through 1.26.3. Registry-aware consumers should use
// ProtocolsWithRegistries before advertising it.
func V1_26_0() minecraft.Protocol {
	return v1_26_0.New()
}

// V1_21_130 returns the wire-only Minecraft protocol 898 family adapter for
// stable Minecraft 1.21.130 through 1.21.132. Registry-aware consumers should
// use ProtocolsWithRegistries before advertising it.
func V1_21_130() minecraft.Protocol {
	return v1_21_130.New()
}

// V1_21_110 returns the wire-only Minecraft protocol 844 family adapter for
// stable Minecraft 1.21.110 through 1.21.114. Registry-aware consumers should
// use ProtocolsWithRegistries before advertising it.
func V1_21_110() minecraft.Protocol {
	return v1_21_110.New()
}

// V1_21_100 returns the wire-only Minecraft protocol 827 family adapter for
// stable Minecraft 1.21.100 through 1.21.102. Registry-aware consumers should
// use ProtocolsWithRegistries before advertising it.
func V1_21_100() minecraft.Protocol {
	return v1_21_100.New()
}

// V1_21_50 returns the wire-only Minecraft protocol 766 family adapter for
// stable Minecraft 1.21.50 and 1.21.51.
func V1_21_50() minecraft.Protocol { return v1_21_50.New() }

// V1_21_40 returns the wire-only Minecraft protocol 748 family adapter for
// stable Minecraft 1.21.40, 1.21.41, 1.21.43, and 1.21.44.
func V1_21_40() minecraft.Protocol { return v1_21_40.New() }

// V1_18_10 returns the wire-only Minecraft protocol 486 family adapter for
// stable Minecraft 1.18.10 through 1.18.12. Registry-aware consumers should
// use ProtocolsWithRegistries before advertising it.
func V1_18_10() minecraft.Protocol {
	return v1_18_10.New()
}

// V1_18_0 returns the wire-only Minecraft protocol 475 family adapter for
// stable Minecraft 1.18.0 through 1.18.2.
func V1_18_0() minecraft.Protocol { return v1_18_0.New() }

// V1_16_100 returns the wire-only Minecraft protocol 419 adapter. Registry-aware consumers should use
// ProtocolsWithRegistries before advertising it.
func V1_16_100() minecraft.Protocol { return v1_16_100.New() }

// V1_21_110WithBlockRegistry returns a configured protocol-844 adapter using
// a current native block registry. Item mapping is initialised from ItemRegistry.
func V1_21_110WithBlockRegistry(native mapping.BlockRegistry) (minecraft.Protocol, error) {
	return v1_21_110.NewWithBlockRegistry(native)
}

// V1_21_100WithBlockRegistry returns a configured protocol-827 adapter using
// a current native block registry. Item mapping is initialised from ItemRegistry.
func V1_21_100WithBlockRegistry(native mapping.BlockRegistry) (minecraft.Protocol, error) {
	return v1_21_100.NewWithBlockRegistry(native)
}

// V1_18_10WithBlockRegistry returns a configured protocol-486 adapter using
// a current native block registry. Item mapping is initialised from ItemRegistry.
func V1_18_10WithBlockRegistry(native mapping.BlockRegistry) (minecraft.Protocol, error) {
	return v1_18_10.NewWithBlockRegistry(native)
}

func V1_21_50WithBlockRegistry(native mapping.BlockRegistry) (minecraft.Protocol, error) {
	return v1_21_50.NewWithBlockRegistry(native)
}

func V1_21_40WithBlockRegistry(native mapping.BlockRegistry) (minecraft.Protocol, error) {
	return v1_21_40.NewWithBlockRegistry(native)
}

func V1_18_0WithBlockRegistry(native mapping.BlockRegistry) (minecraft.Protocol, error) {
	return v1_18_0.NewWithBlockRegistry(native)
}

func V1_16_100WithBlockRegistry(native mapping.BlockRegistry) (minecraft.Protocol, error) {
	return v1_16_100.NewWithBlockRegistry(native)
}

// V1_26_30WithBlockRegistry returns a fully configured protocol-1001 adapter
// using a current native block registry.
func V1_26_30WithBlockRegistry(native mapping.BlockRegistry) (minecraft.Protocol, error) {
	return v1_26_30.NewWithBlockRegistry(native)
}

// V1_26_20WithBlockRegistry returns a configured protocol-975 adapter using a
// current native block registry. Item mapping is initialised from ItemRegistry.
func V1_26_20WithBlockRegistry(native mapping.BlockRegistry) (minecraft.Protocol, error) {
	return v1_26_20.NewWithBlockRegistry(native)
}

// V1_26_10WithBlockRegistry returns a configured protocol-944 adapter using a
// current native block registry. Item mapping is initialised from ItemRegistry.
func V1_26_10WithBlockRegistry(native mapping.BlockRegistry) (minecraft.Protocol, error) {
	return v1_26_10.NewWithBlockRegistry(native)
}

// V1_26_0WithBlockRegistry returns a configured protocol-924 adapter using a
// current native block registry. Item mapping is initialised from ItemRegistry.
func V1_26_0WithBlockRegistry(native mapping.BlockRegistry) (minecraft.Protocol, error) {
	return v1_26_0.NewWithBlockRegistry(native)
}

// V1_21_130WithBlockRegistry returns a configured protocol-898 adapter using
// a current native block registry. Item mapping is initialised from ItemRegistry.
func V1_21_130WithBlockRegistry(native mapping.BlockRegistry) (minecraft.Protocol, error) {
	return v1_21_130.NewWithBlockRegistry(native)
}

// ProtocolsWithBlockRegistry returns all verified non-native protocols,
// including protocol 1001 configured against the current block registry.
func ProtocolsWithBlockRegistry(native mapping.BlockRegistry) ([]minecraft.Protocol, error) {
	legacy, err := V1_26_30WithBlockRegistry(native)
	if err != nil {
		return nil, err
	}
	u2, err := V1_26_20WithBlockRegistry(native)
	if err != nil {
		return nil, err
	}
	u1, err := V1_26_10WithBlockRegistry(native)
	if err != nil {
		return nil, err
	}
	u0, err := V1_26_0WithBlockRegistry(native)
	if err != nil {
		return nil, err
	}
	r21u13, err := V1_21_130WithBlockRegistry(native)
	if err != nil {
		return nil, err
	}
	older, err := V1_21_110WithBlockRegistry(native)
	if err != nil {
		return nil, err
	}
	oldest, err := V1_21_100WithBlockRegistry(native)
	if err != nil {
		return nil, err
	}
	r21u5, err := V1_21_50WithBlockRegistry(native)
	if err != nil {
		return nil, err
	}
	r21u4, err := V1_21_40WithBlockRegistry(native)
	if err != nil {
		return nil, err
	}
	v486, err := V1_18_10WithBlockRegistry(native)
	if err != nil {
		return nil, err
	}
	v475, err := V1_18_0WithBlockRegistry(native)
	if err != nil {
		return nil, err
	}
	return []minecraft.Protocol{v1_26_44.New(), legacy, u2, u1, u0, r21u13, older, oldest, r21u5, r21u4, v486, v475}, nil
}

// ProtocolsWithRegistries returns all verified adapters after eagerly
// validating the current block and item registries.
func ProtocolsWithRegistries(native mapping.BlockRegistry, nativeItems []protocol.ItemEntry) ([]minecraft.Protocol, error) {
	legacy, err := v1_26_30.NewWithRegistries(native, nativeItems)
	if err != nil {
		return nil, err
	}
	u2, err := v1_26_20.NewWithRegistries(native, nativeItems)
	if err != nil {
		return nil, err
	}
	u1, err := v1_26_10.NewWithRegistries(native, nativeItems)
	if err != nil {
		return nil, err
	}
	u0, err := v1_26_0.NewWithRegistries(native, nativeItems)
	if err != nil {
		return nil, err
	}
	r21u13, err := v1_21_130.NewWithRegistries(native, nativeItems)
	if err != nil {
		return nil, err
	}
	older, err := v1_21_110.NewWithRegistries(native, nativeItems)
	if err != nil {
		return nil, err
	}
	oldest, err := v1_21_100.NewWithRegistries(native, nativeItems)
	if err != nil {
		return nil, err
	}
	r21u5, err := v1_21_50.NewWithRegistries(native, nativeItems)
	if err != nil {
		return nil, err
	}
	r21u4, err := v1_21_40.NewWithRegistries(native, nativeItems)
	if err != nil {
		return nil, err
	}
	v486, err := v1_18_10.NewWithRegistries(native, nativeItems)
	if err != nil {
		return nil, err
	}
	v475, err := v1_18_0.NewWithRegistries(native, nativeItems)
	if err != nil {
		return nil, err
	}
	return []minecraft.Protocol{v1_26_44.New(), legacy, u2, u1, u0, r21u13, older, oldest, r21u5, r21u4, v486, v475}, nil
}

// V1_26_44 returns the Minecraft protocol 2168 family adapter. The adapter
// supports Minecraft 1.26.40 through 1.26.44 and keeps this function name for
// API compatibility.
func V1_26_44() minecraft.Protocol {
	return v1_26_44.New()
}
