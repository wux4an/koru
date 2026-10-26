// Type declarations for the koru JavaScript API (`/__koru/core.mjs`).

export type KoruErrorCode =
  | "INVALID_REQUEST"
  | "METHOD_NOT_FOUND"
  | "NATIVE_ERROR"
  | "PERMISSION_DENIED"
  | "STORAGE_ERROR"
  | "UNAVAILABLE"
  | "BUSY"
  | "TIMEOUT"
  | "CANCELLED"
  | "CHANNEL_ERROR"
  | "CHANNEL_REPLACED"
  | "INVALID_RESPONSE"
  | "PAGE_CLOSED";

/** Errors from koru calls are `Error` objects with a stable `code`. */
export interface KoruError extends Error {
  code: KoruErrorCode | string;
}

export interface CallOptions {
  /** Milliseconds from 1 to 120000. Defaults to 10000. */
  timeout?: number;
  signal?: AbortSignal;
}

export type Empty = Record<string, never>;

export interface DeviceInfo {
  model: string;
  manufacturer: string;
  osVersion: string;
  /** Android API level, or 0 in a browser. */
  sdk: number;
}

export interface AppInfo {
  appId: string;
  name: string;
  version: string;
  versionCode: number;
}

export interface PermissionStatus {
  status: "granted" | "denied";
}

export interface NotificationOptions {
  id: number;
  title: string;
  body: string;
}

/** "android" inside the APK; "web" in a browser or under `koru dev`. */
export declare const platform: "android" | "web";
export declare function isNative(): boolean;

/** Call any enabled method, including custom plugin handlers. */
export declare function invoke<T = unknown>(
  method: string,
  params?: Record<string, unknown>,
  options?: CallOptions,
): Promise<T>;

export declare const device: {
  getInfo(options?: CallOptions): Promise<DeviceInfo>;
};

export declare const app: {
  getInfo(options?: CallOptions): Promise<AppInfo>;
};

export declare const storage: {
  get(key: string, options?: CallOptions): Promise<string | null>;
  set(key: string, value: string, options?: CallOptions): Promise<Empty>;
  remove(key: string, options?: CallOptions): Promise<Empty>;
  clear(options?: CallOptions): Promise<Empty>;
};

export declare const vibration: {
  /** Duration in milliseconds, 1 to 10000. Defaults to 100. */
  vibrate(duration?: number, options?: CallOptions): Promise<Empty>;
};

export declare const share: {
  text(text: string, options?: CallOptions): Promise<Empty>;
};

export declare const notifications: {
  show(notification: NotificationOptions, options?: CallOptions): Promise<Empty>;
  cancel(id: number, options?: CallOptions): Promise<Empty>;
  permission(options?: CallOptions): Promise<PermissionStatus>;
  requestPermission(options?: CallOptions): Promise<PermissionStatus>;
};

declare global {
  interface Window {
    Koru: {
      readonly ready: Promise<void>;
      readonly platform: "android" | "web";
      invoke<T = unknown>(
        method: string,
        params?: Record<string, unknown>,
        options?: CallOptions,
      ): Promise<T>;
    };
  }
}
