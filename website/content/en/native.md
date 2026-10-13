# Optional JNI adapter

Standard plugins do not need this adapter. The C implementation demonstrates the plugin ABI through `native.info`; the precompiled x86_64 ELF is distributed at `examples/native/libkoru_native.so` and has been tested on Android 11/ART.

Its code uses only the JNI function table. It links neither libc nor a JVM; the Android process provides ART. The AOSP JNI header retains Apache-2.0, with its license in `include/`. koru's code retains MIT.

## Compile the development example

On an x86_64 host with GCC:

```sh
gcc -std=c11 -Wall -Wextra -Werror -shared -fPIC -ffreestanding -nostdlib \
  -fno-stack-protector -Wl,-soname,libkoru_native.so -Wl,--hash-style=both \
  -Wl,-z,max-page-size=16384 -I runtime/native/include \
  runtime/native/koru_plugin.c -o examples/native/libkoru_native.so
```

This is not part of `koru build`. Another ABI requires an appropriate C compiler and its precompiled ELF. Java, JDK and Gradle are unnecessary. The example is limited to JNI; it does not imply that arbitrary glibc-linked host ELF files work on Android. Other plugins need an Android-compatible C toolchain and dependencies.

## Test thread-scope ownership

```sh
gcc -std=c11 -Wall -Wextra -Werror -I runtime/native/include \
  runtime/native/koru_plugin.c runtime/native/scope_test.c -o /tmp/koru-scope-test
/tmp/koru-scope-test
```

This is an ownership unit test with a controlled JNI table, not an ART simulation. Actual Android execution is recorded in [android-smoke.json]({{base}}original/docs/results/full/android-smoke.json ":ignore").
