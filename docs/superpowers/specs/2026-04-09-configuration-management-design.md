# Design: Configuration Management

**Date:** 2026-04-09
**Status:** Draft

## Purpose

This document describes the configuration management design used by the application: how settings are defined, how defaults are established, how a configuration file overrides those defaults, and how command-line flags override both. The design is generic and intended to be portable to other CLI tools.

## Problem

A CLI application needs a layered configuration model that balances three concerns:

1. **Zero-config usability** — the tool should produce sensible output with no flags and no config file.
2. **Persistent customisation** — users who repeatedly invoke the tool with the same styling or behaviour preferences should be able to capture those in a file rather than repeating flags.
3. **Ad-hoc overrides** — individual invocations may need to override file-based settings without editing the file (e.g., highlighting specific items for a one-off diagram).

A flat "flags only" approach forces repetition. A "config file only" approach makes ad-hoc use clumsy. The design below layers the three sources so that each concern is addressed naturally.

## Design Decisions

### Three-layer configuration model

Configuration is resolved by applying three layers in order. Each layer can override values set by the previous one:

```
Layer 1: Hard-coded defaults (always present)
Layer 2: Configuration file    (optional, loaded when --config is specified)
Layer 3: Command-line flags    (optional, set per invocation)
```

A later layer only overrides a specific field when it supplies a non-zero value for that field. Fields left unset (zero-valued) in a layer are transparent — the value from the previous layer passes through unchanged.

### Single configuration struct

All configuration is represented by a single root struct (e.g., `Config`). This struct is the source of truth throughout the application. CLI flags, config file fields, and defaults all map into the same structure.

Benefits:

- One type to document, serialize, and pass around.
- The "effective configuration" at any point is just the current state of this struct.
- Easy to export (see below) for debugging or bootstrapping a config file.

### Defaults via constructor

A factory function (e.g., `New() *Config`) creates the struct populated with default values. This is called unconditionally at startup, regardless of whether a config file or flags are provided. Defaults are therefore always present and discoverable by reading the constructor.

Rules for defaults:

- Every field that has a meaningful default should be set here.
- Nested sub-structs are allocated and populated with their own defaults.
- Zero values of the language's type system (empty string, 0, false) are intentional when they mean "not set" — they act as sentinel values for the overlay logic.

Example (pseudocode):

```
func New() *Config {
    return &Config{
        GraphType: "",          // empty = resolved later with a fallback
        Rendering: &Rendering{
            Font:     "Verdana",
            FontSize: 16,
            Edges: &EdgeStyle{
                Color: "black",
                Width: 1,
                Style: "solid",
            },
        },
    }
}
```

### Configuration file loading

When the user specifies a config file path via a CLI flag (e.g., `--config path/to/file`), the file is loaded and unmarshalled **on top of** the already-initialised default struct. This ensures that any field not present in the file retains its default value.

Key rules:

1. **Format detection by extension**: `.yaml` / `.yml` → YAML parser; `.json` → JSON parser. Unknown extensions attempt YAML first, then fall back to JSON.
2. **Unmarshal into the existing struct**: The file content is unmarshalled into the struct returned by `New()`, not into a fresh zero-valued struct. This is what makes partial config files work — you only need to specify the fields you want to change.
3. **File read errors are fatal**: If the user explicitly provides `--config` and the file cannot be read or parsed, the application exits with a clear error message. Silent fallback to defaults would hide misconfiguration.
4. **No implicit config file discovery**: The application does not search well-known paths (home directory, XDG config, etc.) for a config file. The file must be explicitly specified. This keeps behaviour predictable and avoids "works on my machine" surprises.

Example partial YAML config file:

```yaml
rendering:
  font: "Fira Code"
  fontSize: 12
  edges:
    color: "red"
```

This overrides only `font`, `fontSize`, and `edges.color`. All other fields retain their defaults.

### Command-line flag overrides

After the config file (if any) is loaded, CLI flags are applied as the final layer. Each flag maps to a specific field in the `Config` struct.

Override semantics by type:

| Type          | Zero value    | Override rule                                                                                          |
| ------------- | ------------- | ------------------------------------------------------------------------------------------------------ |
| `bool`        | `false`       | Flag set → override to `true`. Flags cannot set a value back to `false`; use the config file for that. |
| `string`      | `""`          | Non-empty flag value → override. Empty string means "not set".                                         |
| `int`         | `0`           | Non-zero flag value → override. Zero means "not set".                                                  |
| `[]T` (slice) | `nil` / empty | Non-empty → append to or replace the existing value, depending on the field's semantics.               |

The override function is a single method (e.g., `applyConfigOverrides`) that checks each flag and writes to the config struct only if the flag carries a non-zero value. This makes the logic explicit and auditable.

Example (pseudocode):

```
func applyOverrides(cfg *Config, flags Flags) {
    if flags.GroupByNamespace {
        cfg.GroupByNamespace = true
    }
    if flags.GraphType != "" {
        cfg.GraphType = flags.GraphType
    }
}
```

### Derived and computed settings

Some settings are not stored directly but are computed from a combination of flag values and config values at resolution time. These follow a specific resolution order:

1. If the CLI flag has a non-empty value, use it.
2. Else if the config file set the field, use that.
3. Else fall back to a hard-coded default.

This three-step resolution is used for fields where the "default" is contextual and should not be baked into the constructor. For example, a graph type might default to `"dot"` but the default is applied at resolution time rather than in `New()`, so that the config file can distinguish between "not set" and "explicitly set to dot."

Example (pseudocode):

```
func resolveGraphType(flagValue string, cfg *Config) string {
    if flagValue != "" {
        return flagValue
    }
    if cfg.GraphType != "" {
        return cfg.GraphType
    }
    return "dot"  // hard-coded fallback
}
```

### Complex flag-to-config mappings

Some CLI flags map to config structures in a non-trivial way. For example, a `--highlight` flag that accepts a comma-separated list of patterns might generate an array of style rule structs in the config:

1. Parse the flag value (split on delimiters, trim whitespace, skip empties).
2. For each parsed token, construct a config sub-struct with the token and any relevant defaults (e.g., default highlight colour from the config file).
3. Append the generated sub-structs to the appropriate config slice.

This allows a simple flag syntax to drive rich configuration, while still letting the config file provide the detailed defaults (e.g., what colour to use for highlighting).

### Config export

The application supports exporting the effective (fully resolved) configuration to a file via a CLI flag (e.g., `--export-config path/to/file`). The format is determined by the output file extension (YAML or JSON).

This serves two purposes:

1. **Bootstrapping**: Users can run the tool with their preferred flags, export the resulting config, and use that file going forward instead of repeating the flags.
2. **Debugging**: When output looks wrong, exporting the effective config reveals exactly what settings are in play.

The export happens after config file loading and flag overrides, but before the main operation. This ensures the exported file reflects the exact configuration that would be used for the run.

### Feature flags and runtime-computed configuration

Some configuration is not purely declarative but triggers runtime computation. For example, an `--auto-color` flag might cause the application to analyse the input data and generate additional style rules dynamically.

Design rules for runtime-computed configuration:

1. **Activation is declarative**: The flag or config field is a simple boolean. It does not encode the computed result — it simply enables the computation.
2. **Computed rules integrate with the existing config model**: Generated rules are inserted into the same config structures as user-defined rules (e.g., prepended to a style rules list).
3. **User-defined rules take precedence**: When generated rules conflict with user-defined rules, the user's intent wins. This is achieved by ordering: generated rules are placed before user-defined rules so that last-match-wins semantics give priority to user rules.

## Architecture

### Startup sequence

The application entry point orchestrates configuration in a fixed sequence:

```
1. Parse CLI flags (via CLI parsing library)
2. Create logger (configured from --verbose flag)
3. Create config:
   a. Call New() to get defaults
   b. If --config is set, load and unmarshal file on top of defaults
   c. Apply CLI flag overrides
4. If --export-config is set, write effective config to file
5. Bundle config + logger into a flags/context struct
6. Execute the main command, passing the context struct
7. Within the command, apply any runtime-computed config (e.g., auto-color)
8. Perform the main operation using the final config
```

### Passing configuration through the application

Configuration is bundled into a context struct (e.g., `Flags`) alongside other cross-cutting concerns like the logger. This struct is passed as a parameter to the main command's `Run()` method and from there into sub-operations.

```
type Flags struct {
    Verbose bool
    Log     *Logger
    Config  *Config
}
```

Benefits:

- No global state; configuration is explicit.
- Easy to construct in tests with specific overrides.
- Adding a new cross-cutting concern means adding a field, not refactoring function signatures.

### Config struct design principles

1. **Use `omitempty` on all serialization tags**: This keeps exported config files clean—only non-default values appear. It also makes partial config files natural to write.
2. **Use pointer types for nested structs**: A `nil` pointer means "section not configured" and is omitted during serialization. When present, the sub-struct's fields overlay onto the defaults.
3. **Keep config structs flat where possible**: Deep nesting makes partial overrides harder to reason about. Prefer a two-level hierarchy (root → section) over deeper trees unless the domain demands it.
4. **Config structs are data only**: No methods beyond constructors. Validation, resolution, and defaults live in the calling code, not on the struct itself.

### Supported config file formats

The application supports YAML and JSON. Both formats map 1:1 to the same `Config` struct via struct tags.

| Format | Extension(s)    | Tag                          | Notes                                          |
| ------ | --------------- | ---------------------------- | ---------------------------------------------- |
| YAML   | `.yaml`, `.yml` | `yaml:"fieldName,omitempty"` | Primary format; human-friendly                 |
| JSON   | `.json`         | `json:"fieldName,omitempty"` | Machine-friendly; useful for generated configs |

Both are supported for reading (config file) and writing (export). When the file extension is unrecognised, the loader tries YAML first, then JSON.

### Pattern matching in style rules

Style rules use glob-style patterns (e.g., `*`, `?`) to match item names. All matching rules are applied in order; when multiple rules set the same property, the last match wins. This gives users fine-grained control:

```yaml
nodeStyleRules:
  - match: "build*"
    fillColor: powderblue
  - match: "*test*"
    fillColor: aquamarine
```

## Summary of CLI flags and their config equivalents

The table below shows the relationship between CLI flags and config file fields. This mapping is the contract between the three layers.

| CLI Flag               | Config Field       | Type   | Default    | Override Semantics                             |
| ---------------------- | ------------------ | ------ | ---------- | ---------------------------------------------- |
| (positional)           | —                  | string | (required) | Input file; not stored in config               |
| `--output`, `-o`       | —                  | string | (required) | Output file; not stored in config              |
| `--config`, `-c`       | —                  | string | (none)     | Path to config file; meta-flag                 |
| `--graph-type`         | `graphType`        | string | `"dot"`    | Non-empty → override                           |
| `--group-by-namespace` | `groupByNamespace` | bool   | `false`    | `true` → override                              |
| `--auto-color`         | `autoColor`        | bool   | `false`    | `true` → override                              |
| `--highlight`          | `nodeStyleRules[]` | string | (none)     | Parsed into style rules; appended              |
| `--render-image`       | —                  | string | (none)     | Triggers post-processing; not stored in config |
| `--export-config`      | —                  | string | (none)     | Export path; meta-flag                         |
| `--verbose`            | —                  | bool   | `false`    | Logger level; not stored in config             |

Flags marked "not stored in config" are operational parameters that affect the invocation but do not influence the styling or structural configuration. They are excluded from config file serialization and export.

## Guidelines for extending configuration

When adding a new configurable behaviour:

1. **Add the field to the `Config` struct** with both `json` and `yaml` tags and `omitempty`.
2. **Set a default** in `New()` if the field has a meaningful default.
3. **If it needs a CLI flag**, add the flag to the CLI struct and add the override logic to `applyConfigOverrides`.
4. **If the flag has complex parsing** (e.g., comma-separated values), create a dedicated `apply*Overrides` helper.
5. **Update the export** — no action needed if the field uses `omitempty` tags; it will be included automatically.
6. **Document the new flag** in help text (via struct tags on the CLI struct) and in the config equivalents table.

## Out of Scope

- Config file schema validation beyond what the YAML/JSON unmarshaller provides.
- Implicit config file discovery (XDG, home directory, etc.).
- Environment variable-based configuration.
- Config file inheritance or includes.
- Remote configuration sources.
