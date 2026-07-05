# File Handling

Peddle includes a first file API built on the C64 KERNAL file routines.

The V1 API is:

```peddle
n = fileload(name, buffer, device)
n = filesave(name, buffer, len, device)

f = fileopen(name, mode, device)
n = fileread(f, buffer, max)
n = filewrite(f, buffer, len)
fileclose(f)
```

`name` and `mode` can be `char[]` values or string literals. File buffers can be `byte[]` or `char[]`.

The `device` is usually `8` for the first disk device.

V1 uses one normal logical file number internally. Open one file stream at a time, and call `fileclose(f)` before opening another stream.

---

# Whole-File Helpers

```peddle
var name char[32]
var data char[128]
var n int

copy(name, "SAVEGAME")

n = fileload(name, data, 8)
```

`fileload(name, buffer, device)` opens `name` for reading, reads up to `size(buffer)` bytes, closes the file, updates `len(buffer)`, and returns the number of bytes loaded.

It returns `-1` when the KERNAL open/read path reports an obvious failure.

```peddle
var name char[32]
var data char[128]
var n int

copy(name, "SAVEGAME")
copy(data, "HELLO")

n = filesave(name, data, len(data), 8)
```

`filesave(name, buffer, len, device)` opens `name` for writing, writes up to `min(size(buffer), len)` bytes, closes the file, and returns the number of bytes written.

Write mode uses CBM DOS replace-style naming internally, so saving can replace an existing file with the same name on compatible devices.

---

# Streaming Files

```peddle
var name char[32]
var data char[128]
var f byte
var n int

copy(name, "LOGFILE")
copy(data, "HELLO")

f = fileopen(name, "w", 8)
if f != 0 {
    n = filewrite(f, data, len(data))
    fileclose(f)
}
```

`fileopen(name, mode, device)` opens a file stream and returns a handle. It returns `0` on failure.

Modes:

- `"r"` opens for reading
- `"w"` opens for writing

`filewrite(handle, buffer, len)` writes up to `min(size(buffer), len)` bytes and returns the number of bytes written.

```peddle
var name char[32]
var data char[128]
var f byte
var n int

copy(name, "LOGFILE")

f = fileopen(name, "r", 8)
if f != 0 {
    n = fileread(f, data, size(data))
    fileclose(f)
}
```

`fileread(handle, buffer, max)` clears the destination array length, reads up to `min(size(buffer), max)` bytes, updates `len(buffer)`, and returns the number of bytes read.

`fileclose(handle)` closes the stream and clears KERNAL channels.

---

# Notes

The implementation uses C64 KERNAL calls such as `SETLFS`, `SETNAM`, `OPEN`, `CHKIN`, `CHKOUT`, `CHRIN`, `CHROUT`, `READST`, `CLOSE`, and `CLRCHN`.

This is sequential file access. Random access inside a file is not part of V1.

For game assets, maps, levels, and savegames, prefer `fileload()` and `filesave()` unless you specifically need streaming.

See:

```text
examples/smoke/file_save_load.ped
```
