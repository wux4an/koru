#ifndef KORU_PLUGIN_H
#define KORU_PLUGIN_H
#include <jni.h>
#define KORU_PLUGIN_ABI 1
/* Local references belong to the acquired thread scope. Never cache JNIEnv*. */
typedef struct {
  JavaVM *vm;
  JNIEnv *env;
  int attached;
} koru_env_scope;
int koru_env_acquire(JavaVM *vm, koru_env_scope *scope);
void koru_env_release(koru_env_scope *scope);
jboolean koru_is_ui_thread(JNIEnv *env);
jboolean koru_exception_pending(JNIEnv *env);
#endif
