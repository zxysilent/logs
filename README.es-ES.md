# logs — Registro Estructurado Simple y Rápido para Go

[文档中文](README_ZH.md)

## Características

- **Cuatro niveles de registro**: `DEBUG` `INFO` `WARN` `ERROR`
- **Instancia global predeterminada o instancias personalizadas**: `logs.New(w)` o funciones a nivel de paquete
- **Cadenas de campos estructurados**: `With().Str("k","v").Int("n",1).Info()`
- **Espacio de nombres (Trace)**: `Trace("api").Info()` → `trace=api`
- **Traza distribuida**: `TraceCtx` / `TraceId` / `Ctx`
- **Secuestro automático de stdlib** `log`: `New()` convierte stdlog → logfmt automáticamente
- **Firmas compatibles con stdlib**: `Print/Printf/Println`
- **Salida a archivo**: rotación diaria, edad máxima/tamaño configurables, espejo opcional a consola
- **Alto rendimiento**: ruta rápida sin asignación, reutilización de búfer `sync.Pool`

> **Nota de versión**: v0.12.x requiere **Go 1.24+**.  
> Los usuarios de Go 1.20 deben usar **v0.11.x** (`go get github.com/zxysilent/logs@v0.11`).

---

## Inicio Rápido

```go
package main

import (
    "context"
    stdlog "log"

    "github.com/zxysilent/logs"
)

func main() {
    // === Instancia global predeterminada ===
    logs.SetLevel(logs.LevelDebug) // LevelDebug para desarrollo, LevelInfo para producción
    logs.SetCaller(true)
    logs.Info("hello world")

    // === Espacio de nombres (Trace) ===
    apiLog := logs.Trace("api")
    apiLog.Info("server started")                  // trace=api
    ctx := logs.TraceCtx(context.Background(), "req-1")
    apiLog.Ctx(ctx).Info("handle")                 // trace=api.req-1

    // === Campos estructurados ===
    logs.With().
        Str("user", "alice").
        Int("age", 30).
        Info("user login")

    // === Traza distribuida ===
    ctx = logs.TraceCtx(context.Background())
    logs.Ctx(ctx).Str("op", "query").Debug("trace")

    // === Compatibilidad con stdlib ===
    logs.Print("stdlib", "message")                // msg=stdlibmessage
    logs.Printf("stdlib %s", "format")             // msg=stdlib format
    stdlog.Println("auto hijacked to logfmt")      // secuestrado por New()

    // === Instancia personalizada (salida a archivo) ===
    // Las instancias personalizadas tienen el salto de llamante correcto; agregue WithSkip(1) si está envuelto en un helper
    w, closeFn := logs.NewFile("./logs/app.log", logs.WithMaxAge(7), logs.WithMaxSize(64), logs.WithConsole(true))
    defer closeFn()
    applog := logs.New(w, logs.WithLevel(logs.LevelInfo))
    applog.Info("app started")
}
```

---

## Referencia de API

### Niveles de Registro

Alineados numéricamente con `log/slog` (mayor = más severo).

| Constante | Valor | Descripción |
|----------|-------|-------------|
| `logs.LevelDebug` | -4 | Depuración |
| `logs.LevelInfo` | 0 | Información |
| `logs.LevelWarn` | 4 | Advertencia |
| `logs.LevelError` | 8 | Error |
| `logs.LevelMute` | 20241020 | Desactiva toda la salida (centinela) |

> `LDEBUG` / `LINFO` / `LWARN` / `LERROR` / `LNONE` están obsoletos y se eliminarán en una futura versión principal.

`ParseLevel` convierte una cadena sin distinción de mayúsculas/minúsculas a un `Level`:
```go
logs.ParseLevel("debug")  // LevelDebug
logs.ParseLevel("WARN")   // LevelWarn
logs.ParseLevel("OFF")    // LevelMute
// Acepta: D/DBG/DEBUG/-4, I/INF/INFO/0, W/WRN/WARN/WARNING/4, E/ERR/ERROR/8, OFF/NONE/MUTE
```

### Funciones a Nivel de Paquete (operan en la instancia predeterminada)

> **Nota**: Las funciones `Set*` deben configurarse una sola vez antes de comenzar el registro.
> No se recomienda la modificación en tiempo de ejecución — las escrituras de configuración no están sincronizadas
> y pueden tener condiciones de carrera con la salida de registro concurrente. Configure durante la inicialización y
> use instancias `New()` inmutables para uso en tiempo de ejecución.

```go
logs.SetLevel(lv Level)                             // establecer nivel de registro
logs.SetCaller(b bool)                              // habilitar/deshabilitar línea de llamante
logs.SetSep(sep ...string)                          // separadores de ruta, por defecto "/" (la coincidencia más a la derecha gana)
logs.SetSkip(skip int)                              // marcos de salto adicionales del llamante
logs.SetOutput(out io.Writer)                       // establecer escritor de salida
logs.SetFile(path string)                           // establecer salida a archivo
logs.SetMaxAge(ma int)                              // días máximos de retención, por defecto 64
logs.SetMaxSize(ms int64)                           // tamaño máximo del archivo (MiB), por defecto 64
logs.SetConsole(b bool)                             // también imprimir en stderr (recomendado)
logs.SetTrace(trace string)                          // establecer espacio de nombres en la instancia predeterminada
logs.Close() error                                    // cerrar

// Salida
logs.Debug(args ...any)
logs.Debugf(format string, args ...any)
logs.Info(args ...any)
logs.Infof(format string, args ...any)
logs.Warn(args ...any)
logs.Warnf(format string, args ...any)
logs.Error(args ...any)
logs.Errorf(format string, args ...any)

// Compatibilidad con stdlib
logs.Print(args ...any)
logs.Println(args ...any)
logs.Printf(format string, args ...any)

// Cadena de campos / trazado
logs.With(trace ...string) *fielder
logs.Ctx(ctx context.Context) *fielder

// Espacio de nombres / sub-Logger
logs.Trace(trace string) *Logger     // reemplazar espacio de nombres
logs.Clone(trace ...string) *Logger  // copiar (sin args) o agregar trace
```

### Logger (instancia personalizada)

**Prefiera la instancia predeterminada a nivel de paquete.** No requiere inicialización y el salto de llamante ya es correcto.

Un logger `New` se configura una sola vez mediante opciones funcionales y es **inmutable** después
(no tiene métodos `Set*`). Para configuración mutable en tiempo de ejecución, use la instancia predeterminada a nivel de paquete.

```go
// Construir con opciones (out=nil significa Descartar)
l := logs.New(w,
    logs.WithLevel(logs.LevelDebug),
    logs.WithCaller(true),
    logs.WithSep("/internal", "/"),
    logs.WithSkip(0),
    logs.WithHijack(true),  // por defecto true; false para deshabilitar el secuestro de stdlib
)

// Si su instancia personalizada está envuelta en un helper, agregue WithSkip(1) para que el llamante
// apunte al sitio de llamada real:
helper := func(msg string) {
    l.Info(msg)
}
_ = logs.New(w, logs.WithCaller(true), logs.WithSkip(1))
_ = helper // caller(file:line) apunta al llamante de helper("msg")

// Salida a archivo: NewFile devuelve el Writer + un control de cierre; WithMaxAge/WithMaxSize/WithConsole opcionales
w, closeFn := logs.NewFile("app.log", logs.WithMaxAge(7), logs.WithMaxSize(64), logs.WithConsole(true))
defer closeFn()
fl := logs.New(w)
l.Debug(...)  l.Debugf(...)  l.Info(...)  l.Infof(...)
l.Warn(...)   l.Warnf(...)   l.Error(...) l.Errorf(...)
l.Print(...)  l.Println(...) l.Printf(...)
l.With(trace ...string) *fielder   l.Ctx(ctx) *fielder
l.Trace(trace string) *Logger      // sub-logger con espacio de nombres (comparte la configuración raíz)
l.Clone(trace ...string) *Logger  // copiar (sin args) o agregar trace
```

### Espacio de Nombres / sub-Logger

`Trace`/`Clone` derivan un sub-`Logger` que comparte la `Config` raíz del padre.

```go
api := logs.Trace("api")         // *Logger, trace=api (reemplazar)
pay := api.Clone("pay")          // *Logger, trace=api.pay (agregar)
api.Debug(...)  api.Info(...)  api.Warn(...)  api.Error(...)  api.Print(...)
api.With() *fielder              // derivar un fielder de un solo uso, hereda attr+trace
api.Ctx(ctx context.Context) *fielder
// trace = ns (sin ctx) o ns.trace (con ctx)

// Congelar una cadena de campos en un *Logger persistente, reutilizable y seguro para concurrencia:
base := logs.With().Str("svc", "api").Int("pid", 1).Group() // *Logger
base.Info("started")             // svc=api pid=1, no liberado, reutilizable
base.With().Int("uid", 9).Info("login")
```

### fielder (campos estructurados + control de salida)

```go
// Campos
fl.Str(key, val string)          fl.Stringer(key string, val fmt.Stringer)
fl.Bytes(key string, val []byte) fl.Err(err error)    fl.IfErr(err error)
fl.Bool(key string, b bool)
fl.Int(key string, i int)        fl.Int8(key, i int8)   fl.Int16(key, i int16)
fl.Int32(key, i int32)           fl.Int64(key, i int64)
fl.Uint(key, i uint)             fl.Uint8(key, i uint8) fl.Uint16(key, i uint16)
fl.Uint32(key, i uint32)         fl.Uint64(key, i uint64)
fl.Float32(key string, f float32) fl.Float64(key string, f float64)
fl.Time(key string, t time.Time) fl.Dur(key string, d time.Duration)
fl.Any(key string, i any)        fl.Raw(key string, b []byte)

// Control
fl.If(b bool)                    // salida condicional
fl.Caller(b bool)                // control de llamante por entrada

// Congelar en un *Logger reutilizable
fl.Group() *Logger                    // persistir cadena de campos (sin liberación manual)

// Métodos terminales (fielder se recicla después de la llamada)
fl.Debug(args ...any)   fl.Debugf(format string, args ...any)
fl.Info(args ...any)    fl.Infof(format string, args ...any)
fl.Warn(args ...any)    fl.Warnf(format string, args ...any)
fl.Error(args ...any)   fl.Errorf(format string, args ...any)
```

### Traza Distribuida

```go
ctx := logs.TraceCtx(context.Background())          // generar nueva traza
ctx := logs.TraceCtx(context.Background(), "myid")  // usar id especificado
ctx = logs.TraceCtx(ctx, "child")                   // agregar → myid.child
ctx = logs.TraceCtx(ctx)                            // reutilizar traza existente
traceId := logs.TraceOf(ctx)                        // leer traza
id := logs.TraceId()                                  // generar id independiente
```

### Integración con stdlib

```go
// Auto-secuestro — New() llama a hijackstd(), convirtiendo stdlog → logfmt
// prefix se captura como espacio de nombres de log
stdlog.SetPrefix("myprefix")
_ = logs.New(nil)             // hijack lee prefix → trace=myprefix
stdlog.Println("hello")       // salida: trace=myprefix level=INF msg=hello

// Compatibilidad Print — Print/Printf/Println a nivel de paquete → logfmt
logs.Print("a", "b")          // msg=ab
logs.Printf("%s:%d", "k", 1)  // msg=k:1
```

---

## Formato de Salida (logfmt)

```
time=2026-01-01T12:00:00.000 level=INF msg="hello world"
time=2026-01-01T12:00:00.000 level=INF trace=api.req-1 caller=/main.go:42 user=alice msg=login
time=2026-01-01T12:00:00.000 level=ERR trace=api error="something failed" msg="request failed"
```

- `time` / `level` siempre presentes
- `trace` — presente cuando se usa trazado/espacio de nombres
- `caller` — presente cuando `SetCaller(true)` está establecido (`archivo:línea`)
- `error` — presente cuando se llama a `Err/IfErr`

---

## Integración con xorm

```go
db.AddHook(&repoHook{showSql: true})

type repoHook struct { showSql bool }

func (rh *repoHook) BeforeProcess(ctx *contexts.ContextHook) (context.Context, error) {
    return ctx.Ctx, nil
}

func (rh *repoHook) AfterProcess(ctx *contexts.ContextHook) error {
    if ctx.Err != nil {
        logs.Ctx(ctx.Ctx).Err(ctx.Err).Str("SQL", ctx.SQL).
            Any("args", ctx.Args).Dur("dur", ctx.ExecuteTime).Error()
    } else if ctx.ExecuteTime > 200*time.Millisecond {
        logs.Ctx(ctx.Ctx).Str("SlowSQL", ctx.SQL).
            Any("args", ctx.Args).Dur("dur", ctx.ExecuteTime).Warn()
    } else if rh.showSql {
        logs.Ctx(ctx.Ctx).Str("SQL", ctx.SQL).
            Any("args", ctx.Args).Dur("dur", ctx.ExecuteTime).Debug()
    }
    return ctx.Err
}
```

---

## Rendimiento

```
pkg: github.com/zxysilent/logs
cpu: 12th Gen Intel(R) Core(TM) i5-12500H
count: average of 3 runs

BenchmarkDisabled         1.0 ns/op,   0 B/op, 0 allocs   // ruta rápida de filtro de nivel
BenchmarkParallelSimple    12 ns/op,   0 B/op, 0 allocs   // salida paralela simple
BenchmarkParallelSpan      62 ns/op,   0 B/op, 0 allocs   // paralelo Trace + salida
BenchmarkParallel          60 ns/op,   0 B/op, 0 allocs   // paralelo With 7 campos
BenchmarkSimple            50 ns/op,   0 B/op, 0 allocs   // Info() básico
BenchmarkError            106 ns/op,   0 B/op, 0 allocs   // Log de error
BenchmarkInfof            100 ns/op,  16 B/op, 1 allocs   // salida formateada
BenchmarkWith5Fields      187 ns/op,   0 B/op, 0 allocs   // 5 campos estructurados
BenchmarkWith10Fields     289 ns/op,   0 B/op, 0 allocs   // 10 campos estructurados
BenchmarkSimpleCaller     284 ns/op,   0 B/op, 0 allocs   // Info + llamante
BenchmarkParallelFile     309 ns/op,   0 B/op, 0 allocs   // escritura de archivo paralelo
```


## Inspirado por

[zerolog](https://github.com/rs/zerolog/)
