# Peddle Smoke Programs

Small runnable programs used to exercise compiler and runtime behavior.

These are good candidates for automated compile checks. Programs with stable
text or screen output can also be checked with VICE screenshots.

Use `capture_screenshots.sh` to compile smoke entries and write VICE screenshots
beside the programs:

```sh
examples/smoke/capture_screenshots.sh
examples/smoke/capture_screenshots.sh players import_demo/main
examples/smoke/capture_screenshots.sh --help
```
