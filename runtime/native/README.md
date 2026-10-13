# Adaptador JNI opcional

Los plugins estándar no necesitan este adaptador. La implementación C demuestra
la ABI de plugin con `native.info`; el ELF x86_64 precompilado se distribuye en
`examples/native/libkoru_native.so` y se ha probado en Android 11/ART.

Su código usa exclusivamente la tabla de funciones JNI. No enlaza libc ni una
JVM; el proceso Android proporciona ART. El header JNI de AOSP conserva Apache-2.0,
con licencia incluida en `include/`. El código koru conserva MIT.

Compilar el ejemplo **de desarrollo**, en un host x86_64 con GCC:

```sh
gcc -std=c11 -Wall -Wextra -Werror -shared -fPIC -ffreestanding -nostdlib \
  -fno-stack-protector -Wl,-soname,libkoru_native.so -Wl,--hash-style=both \
  -Wl,-z,max-page-size=16384 -I runtime/native/include \
  runtime/native/koru_plugin.c -o examples/native/libkoru_native.so
```

Esto no forma parte de koru build. Para otra ABI se usa un compilador C adecuado
y se distribuye el ELF precompilado correspondiente. No se necesita Java/JDK/Gradle.
El ejemplo se limita a JNI y no pretende que un ELF general compilado contra glibc
funcione en Android; para otros plugins hace falta un toolchain C compatible con
Android y sus dependencias.

Prueba del scope de hilo:

```sh
gcc -std=c11 -Wall -Wextra -Werror -I runtime/native/include \
  runtime/native/koru_plugin.c runtime/native/scope_test.c -o /tmp/koru-scope-test
/tmp/koru-scope-test
```

Es una prueba unitaria de ownership con una tabla JNI controlada, no una simulación
de ART. La ejecución real Android está registrada en docs/results/full/android-smoke.json.
