#include "koru_plugin.h"
#include <assert.h>
static int attaches, detaches;
static jint status, attach_result;
static jint get_env(JavaVM *vm, void **env, jint version) {
  (void)vm;
  (void)version;
  *env = (void *)1;
  return status;
}
static jint attach(JavaVM *vm, JNIEnv **env, void *args) {
  (void)vm;
  (void)args;
  attaches++;
  *env = (JNIEnv *)1;
  return attach_result;
}
static jint detach(JavaVM *vm) {
  (void)vm;
  detaches++;
  return JNI_OK;
}
int main(void) {
  struct JNIInvokeInterface table = {
      .GetEnv = get_env, .AttachCurrentThread = attach, .DetachCurrentThread = detach};
  JavaVM vm = &table;
  koru_env_scope scope;
  status = JNI_OK;
  assert(koru_env_acquire(&vm, &scope) == JNI_OK && !scope.attached);
  koru_env_release(&scope);
  assert(attaches == 0 && detaches == 0 && scope.env == 0);
  status = JNI_EDETACHED;
  attach_result = JNI_OK;
  assert(koru_env_acquire(&vm, &scope) == JNI_OK && scope.attached);
  koru_env_release(&scope);
  assert(attaches == 1 && detaches == 1);
  attach_result = JNI_ERR;
  assert(koru_env_acquire(&vm, &scope) == JNI_ERR && !scope.attached);
  koru_env_release(&scope);
  assert(detaches == 1);
  status = JNI_EVERSION;
  assert(koru_env_acquire(&vm, &scope) == JNI_EVERSION);
  koru_env_release(&scope);
  assert(detaches == 1);
  return 0;
}
