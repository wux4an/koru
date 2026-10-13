/* Optional freestanding JNI adapter. No JVM implementation or libc dependency.
 */
#include "koru_plugin.h"
static JavaVM *runtime_vm;
static jclass json_class;
static jmethodID json_ctor;
static jmethodID json_put;
static jmethodID json_put_int;

int koru_env_acquire(JavaVM *vm, koru_env_scope *scope) {
  if (!vm || !scope)
    return JNI_ERR;
  scope->vm = vm;
  scope->env = 0;
  scope->attached = 0;
  jint status = (*vm)->GetEnv(vm, (void **)&scope->env, JNI_VERSION_1_6);
  if (status == JNI_EDETACHED) {
    status = (*vm)->AttachCurrentThread(vm, &scope->env, 0);
    if (status == JNI_OK)
      scope->attached = 1;
  }
  return status;
}
void koru_env_release(koru_env_scope *scope) {
  if (!scope)
    return;
  if (scope->attached && scope->vm)
    (*scope->vm)->DetachCurrentThread(scope->vm);
  scope->env = 0;
  scope->vm = 0;
  scope->attached = 0;
}
jboolean koru_exception_pending(JNIEnv *env) { return (*env)->ExceptionCheck(env); }
jboolean koru_is_ui_thread(JNIEnv *env) {
  jclass looper = (*env)->FindClass(env, "android/os/Looper");
  if (!looper)
    return JNI_FALSE;
  jmethodID main_id =
      (*env)->GetStaticMethodID(env, looper, "getMainLooper", "()Landroid/os/Looper;");
  jmethodID current_id =
      (*env)->GetStaticMethodID(env, looper, "myLooper", "()Landroid/os/Looper;");
  if (!main_id || !current_id || (*env)->ExceptionCheck(env)) {
    (*env)->DeleteLocalRef(env, looper);
    return JNI_FALSE;
  }
  jobject main = (*env)->CallStaticObjectMethod(env, looper, main_id);
  if ((*env)->ExceptionCheck(env)) {
    (*env)->DeleteLocalRef(env, looper);
    return JNI_FALSE;
  }
  jobject current = (*env)->CallStaticObjectMethod(env, looper, current_id);
  jboolean result = JNI_FALSE;
  if (!(*env)->ExceptionCheck(env))
    result = (*env)->IsSameObject(env, main, current);
  if (main)
    (*env)->DeleteLocalRef(env, main);
  if (current)
    (*env)->DeleteLocalRef(env, current);
  (*env)->DeleteLocalRef(env, looper);
  return result;
}
JNIEXPORT jint JNICALL JNI_OnLoad(JavaVM *vm, void *reserved) {
  (void)reserved;
  JNIEnv *env = 0;
  if ((*vm)->GetEnv(vm, (void **)&env, JNI_VERSION_1_6) != JNI_OK)
    return JNI_ERR;
  jclass local = (*env)->FindClass(env, "org/json/JSONObject");
  if (!local)
    return JNI_ERR;
  json_class = (*env)->NewGlobalRef(env, local);
  (*env)->DeleteLocalRef(env, local);
  if (!json_class)
    return JNI_ERR;
  json_ctor = (*env)->GetMethodID(env, json_class, "<init>", "()V");
  json_put = (*env)->GetMethodID(env, json_class, "put",
                                 "(Ljava/lang/String;Ljava/lang/Object;)Lorg/json/JSONObject;");
  json_put_int =
      (*env)->GetMethodID(env, json_class, "put", "(Ljava/lang/String;I)Lorg/json/JSONObject;");
  if (!json_ctor || !json_put || !json_put_int || (*env)->ExceptionCheck(env)) {
    (*env)->DeleteGlobalRef(env, json_class);
    json_class = 0;
    return JNI_ERR;
  }
  runtime_vm = vm;
  return JNI_VERSION_1_6;
}
JNIEXPORT void JNICALL JNI_OnUnload(JavaVM *vm, void *reserved) {
  (void)reserved;
  koru_env_scope scope;
  if (koru_env_acquire(vm, &scope) == JNI_OK) {
    if (json_class)
      (*scope.env)->DeleteGlobalRef(scope.env, json_class);
    koru_env_release(&scope);
  }
  json_class = 0;
  runtime_vm = 0;
}
static int put_text(JNIEnv *env, jobject result, const char *key, const char *text) {
  jstring k = (*env)->NewStringUTF(env, key);
  if (!k)
    return 0;
  jstring v = (*env)->NewStringUTF(env, text);
  if (!v) {
    (*env)->DeleteLocalRef(env, k);
    return 0;
  }
  jobject returned = (*env)->CallObjectMethod(env, result, json_put, k, v);
  int ok = !(*env)->ExceptionCheck(env);
  if (returned)
    (*env)->DeleteLocalRef(env, returned);
  (*env)->DeleteLocalRef(env, k);
  (*env)->DeleteLocalRef(env, v);
  return ok;
}
JNIEXPORT jobject JNICALL Java_com_koru_plugins_Native_info(JNIEnv *env, jclass cls,
                                                            jobject activity, jobject params) {
  (void)cls;
  (void)activity;
  (void)params;
  if (!runtime_vm || !json_class)
    return 0;
  koru_env_scope scope;
  if (koru_env_acquire(runtime_vm, &scope) != JNI_OK)
    return 0;
  jobject result = (*env)->NewObject(env, json_class, json_ctor);
  if (!result) {
    koru_env_release(&scope);
    return 0;
  }
  int ok = put_text(env, result, "runtime", "JNI");
  if (ok) {
    jboolean ui = koru_is_ui_thread(env);
    if ((*env)->ExceptionCheck(env))
      ok = 0;
    else
      ok = put_text(env, result, "thread", ui ? "main" : "worker");
  }
  if (ok) {
    jstring key = (*env)->NewStringUTF(env, "abi");
    if (key) {
      jobject returned = (*env)->CallObjectMethod(env, result, json_put_int, key, KORU_PLUGIN_ABI);
      ok = !(*env)->ExceptionCheck(env);
      if (returned)
        (*env)->DeleteLocalRef(env, returned);
      (*env)->DeleteLocalRef(env, key);
    } else
      ok = 0;
  }
  if (!ok) {
    (*env)->DeleteLocalRef(env, result);
    result = 0;
  }
  koru_env_release(&scope);
  return result;
}
