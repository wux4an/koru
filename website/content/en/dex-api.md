# DEX 035 API

This API is experimental (`0.2.0-dev`) and will stabilize before v1. Every error propagates explicitly; missing references are not silently resolved.

```go
import "koru.dev/koru/pkg/dex"

code := dex.NewCode(1).
    Const(dex.V(0), 42).
    Emit(dex.Instruction{Op: "return", Regs: []dex.Register{dex.V(0)}}).
    Build()

program := dex.Program{Classes: []dex.Class{{
    Descriptor: "Lexample/Answer;",
    Super:      "Ljava/lang/Object;",
    Access:     dex.Public,
    Methods: []dex.Method{{
        Name: "answer", Descriptor: "()I",
        Access: dex.Public | dex.Static, Code: code,
    }},
}}}

bytes, err := dex.Build(program)
// Handle err before using bytes.
```

`dex.Build` validates definitions, generates tables and sections, then passes the file through the independent parser. `dex.Inspect` returns the decoded representation (`dex.Inspection`) only when structure and bytecode pass validation for the supported subset. These checks do not guarantee ART class resolution and verification; that requires Android execution. The included runtime has been executed on ART/API 30; arbitrary IR needs its own validation.

## Registers and parameters

`Code.Locals` reserves contiguous local slots. `V(n)` refers to a local slot; `P(n)` refers to a parameter word. Automatic register allocation for high-level language variables is not implemented.

`onCreate(Bundle)` with four locals has six registers: `v0..v3`, `this` in `v4=P(0)`, and `Bundle` in `v5=P(1)`. In static methods `P(0)` is the first word of the first parameter. `long` and `double` occupy two words. Parameter physical registers are decided during assembly, not while writing IR.

In JSON, write `V(n)` as `n`, and `P(n)` as `-n-1`. See the constructor in `examples/dex/answer.json`. Flags are integers, for example `Public|Static=9`, `Public|Constructor=65537`. `Public|Protected` is an error: contradictory visibility combinations are disallowed.

`CodeBuilder.Emit` and `Build` copy mutable registers and references; a snapshot can be changed without affecting the builder or earlier snapshots.

## Implemented instructions

| IR `Op`                                                                                | Operands                                                               |
| -------------------------------------------------------------------------------------- | ---------------------------------------------------------------------- |
| `nop`, `return-void`                                                                   | No registers.                                                          |
| `move`, `move-object`                                                                  | `[destination, source]`; selects 12x, 22x or 32x.                      |
| `move-result`, `move-result-object`                                                    | `[destination]`, immediately after a non-void invoke.                  |
| `return`, `return-object`                                                              | `[value]`, category consistent with the prototype.                     |
| `const/4`, `const/16`, `const`                                                         | `[destination]`, int32 `Literal`; explicit limits.                     |
| `const-string`                                                                         | `[destination]`, `String`; selects jumbo above index 65535.            |
| `const-class`, `new-instance`, `check-cast`                                            | `[destination]`, `Type` descriptor.                                    |
| `invoke-direct`, `invoke-virtual`, `invoke-static`, `invoke-super`, `invoke-interface` | `Method` and register-word list, including receiver except for static. |
| `iget`, `iput`                                                                         | `[value, object]`, `Field`.                                            |
| `sget`, `sput`                                                                         | `[value]`, `Field`.                                                    |
| `if-eqz`, `if-nez`                                                                     | `[register]`, `Target` label.                                          |
| `instance-of`, `new-array`                                                             | `[destination, source/size]`, `Type`.                                  |
| `aget-object`, `aput-object`                                                           | `[value, array, index]`.                                               |
| `int-to-long`                                                                          | `[wide destination, source]`.                                          |
| `if-lt`, `if-ge`, `if-ne`                                                              | `[left, right]`, `Target`.                                             |
| `move-exception`                                                                       | `[destination]`, at the start of a handler.                            |
| `goto`                                                                                 | `Target`; emits `goto/32`.                                             |
| `label`                                                                                | `Label`; pseudo-instruction with no code units.                        |

Field access selects object, boolean, byte, char and short variants by descriptor; wide fields are rejected. `invoke` uses 35c for up to five words in four-bit registers, or 3rc for contiguous registers when that form is exceeded. Noncontiguous ranges and prototype-incompatible counts are rejected. `outs_size` is the maximum word count across invocations.

Relative jumps are measured from the opcode in 16-bit units. The parser rejects jumps into instructions, destinations outside code, jumps to `move-result`, results without invocation, and reachable flow falling past method end. It does not yet check all propagated register types.

## Identifiers and sections

Strings are sorted by UTF-16 units, not UTF-8 bytes, and serialized in MUTF-8 with a UTF-16-unit length and terminator. The emitter accepts valid UTF-8 Go strings. The decoder preserves isolated surrogate units for comparing DEX strings; Go's visible representation uses replacement characters for unpaired units. Descriptors and names are validated against DEX 035.

Types sort by string index. Prototypes sort primarily by return type and secondarily by lexicographic parameter list. Fields sort by class, name and type; methods by class, name and prototype. Repeated references are interned and repeated definitions rejected.

Classes sort by dependency, with in-file superclasses and interfaces first. Cycles are errors. Concrete methods have code; native and abstract methods do not. Constructors are direct methods and require their flag.

Headers and IDs are little-endian and fixed-size. Data sections are contiguous by type: strings, shared type_lists, code_items, class_data and map. Required structures align to four bytes. class_data field/method indices are differential ULEB128, sorted within each group.

Each map has exactly one entry per nonempty section. SHA-1 covers `data[32:]`; Adler-32 covers `data[12:]` after writing SHA-1. These are DEX format integrity hashes, not developer identity signatures.

Current limits: 65,535 types and prototypes; 65,536 field and method references; 65,535 registers per method; 255 words for invoke/range. The emitter checks 32-bit offset and size overflow. The CLI limits each input file to 64 MiB to bound hostile input cost.

## Exceptions

`CodeBuilder.Catch(start, end, handler, typeDescriptor)` declares a label-delimited protected range with an exclusive end; an empty descriptor means catch-all. One handler per range and multiple nonoverlapping ranges are supported. The emitter writes try_items and encoded_catch_handler_list, including padding and offsets. The independent parser checks bounds, destinations and move-exception. General annotations and multiple handlers for one range are not implemented in this version.
