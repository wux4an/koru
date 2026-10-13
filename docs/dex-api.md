# API DEX 035

Esta API es experimental (`0.2.0-dev`); se estabilizará antes de v1. Cada error se
propaga explícitamente. No hay resolución silenciosa de referencias ausentes.

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

`dex.Build` valida las definiciones, genera las tablas y secciones y pasa el archivo
por el parser independiente. `dex.Inspect` devuelve la representación decodificada
(`dex.Inspection`) solo cuando la estructura y el bytecode del subconjunto pasan
la validación. Las comprobaciones no garantizan que ART resuelva y verifique las
clases; ese contrato requiere ejecución en Android. El runtime incluido sí se ha
ejecutado en ART/API 30; un IR arbitrario requiere su propia comprobación.

## Registros y parámetros

`Code.Locals` reserva slots locales contiguos. `V(n)` refiere un slot local;
`P(n)` refiere una palabra de parámetro. No se implementa aún asignación automática
de registros a variables de un lenguaje de alto nivel.

Para `onCreate(Bundle)` con cuatro locales hay seis registros: `v0..v3`, `this`
en `v4=P(0)` y `Bundle` en `v5=P(1)`. En métodos estáticos `P(0)` es la primera
palabra del primer parámetro. `long` y `double` ocupan dos palabras. El registro
físico de un parámetro se decide al ensamblar, no al escribir el IR.

En JSON `V(n)` se escribe como `n`, y `P(n)` como `-n-1`. Véase el constructor en
`examples/dex/answer.json`. Los flags se escriben como enteros; por ejemplo
`Public|Static=9`, `Public|Constructor=65537`. `Public|Protected` es un error:
no se permite combinar visibilidades contradictorias.

`CodeBuilder.Emit` y `Build` copian registros y referencias mutables; cada
snapshot puede modificarse sin cambiar el builder ni snapshots anteriores.

## Instrucciones implementadas

| `Op` del IR | Operandos |
| --- | --- |
| `nop`, `return-void` | Sin registros |
| `move`, `move-object` | `[destino, fuente]`; elige 12x, 22x o 32x |
| `move-result`, `move-result-object` | `[destino]`, inmediatamente después de invoke no void |
| `return`, `return-object` | `[valor]`, categoría coherente con el prototipo |
| `const/4`, `const/16`, `const` | `[destino]`, `Literal` int32; límites explícitos |
| `const-string` | `[destino]`, `String`; selecciona jumbo si el índice supera 65535 |
| `const-class`, `new-instance`, `check-cast` | `[destino]`, `Type` descriptor |
| `invoke-direct`, `invoke-virtual`, `invoke-static`, `invoke-super`, `invoke-interface` | `Method` y lista de palabras de registros, incluido receptor salvo static |
| `iget`, `iput` | `[valor, objeto]`, `Field` |
| `sget`, `sput` | `[valor]`, `Field` |
| `if-eqz`, `if-nez` | `[registro]`, `Target` etiqueta |
| `instance-of`, `new-array` | `[destino, fuente/tamaño]`, `Type` |
| `aget-object`, `aput-object` | `[valor, array, índice]` |
| `int-to-long` | `[destino wide, fuente]` |
| `if-lt`, `if-ge`, `if-ne` | `[izquierda, derecha]`, `Target` |
| `move-exception` | `[destino]`, al inicio de un handler |
| `goto` | `Target`; se emite `goto/32` |
| `label` | `Label`; pseudo instrucción sin code units |

Los accesos a campos eligen las variantes object, boolean, byte, char y short a
partir del descriptor. Se rechazan campos wide. `invoke` usa 35c hasta cinco
palabras en registros de cuatro bits; usa 3rc cuando los registros son contiguos
si supera esa forma. Se rechazan rangos no contiguos y conteos incompatibles con
el prototipo. `outs_size` es el máximo de palabras de todas las invocaciones.

Los saltos relativos se miden desde el opcode en unidades de 16 bits. El parser
rechaza saltos dentro de instrucciones, destinos fuera del código, saltos a
`move-result`, resultados sin invocación y flujo alcanzable que caiga fuera del
método. No comprueba todavía todos los tipos propagados por cada registro.

## Identificadores y secciones

Strings: orden por unidades UTF-16, no por bytes UTF-8. Se serializan en MUTF-8
con longitud en unidades UTF-16 y terminador. El emisor acepta strings Go UTF-8
válidos; el decoder conserva unidades surrogate aisladas para comparar strings
DEX. Su representación visible en Go usa el carácter de reemplazo para unidades
que no forman pares. Los descriptores y nombres se validan contra DEX 035.

Tipos: orden por string index. Prototipos: return type mayor y lista de parámetros
lexicográfica menor. Campos: clase, nombre, tipo. Métodos: clase, nombre, prototipo.
Se internan referencias repetidas y se rechazan definiciones repetidas.

Las clases se ordenan por dependencias: primero superclases e interfaces definidas
en el archivo. Los ciclos son errores. Métodos concretos tienen código; native y
abstract no lo tienen. Constructores son direct methods y requieren su flag.

Header e IDs son little endian y de tamaño fijo. Las secciones de datos se emiten
contiguas por tipo: strings, type_lists compartidas, code_items, class_data, mapa.
Las estructuras que lo requieren se alinean a cuatro bytes. Los campos/métodos
en class_data usan índices diferenciales ULEB128 ordenados por grupo.

Cada mapa contiene exactamente una entrada por sección no vacía. SHA-1 cubre
`data[32:]`; Adler-32 cubre `data[12:]`, después de escribir SHA-1. Son hashes de
integridad del formato DEX, no una firma de identidad del desarrollador.

Los límites actuales son 65535 tipos y prototipos, 65536 referencias a campos y
métodos, 65535 registros por método y 255 palabras para la forma invoke/range.
El emisor comprueba desbordamientos de offsets/tamaños de 32 bits; la CLI limita
cada archivo de entrada a 64 MiB para contener el coste de entradas hostiles.

## Excepciones

`CodeBuilder.Catch(start, end, handler, typeDescriptor)` declara un rango protegido
por etiquetas, con final exclusivo; descriptor vacío indica catch-all. Se admiten
un handler por rango y varios rangos no solapados. El emisor escribe try_items y
encoded_catch_handler_list, incluyendo padding y offsets; el parser independiente
comprueba límites, destinos y move-exception. No se implementan anotaciones generales
ni listas de múltiples handlers para un mismo rango en esta versión.
