const storageKey = "adoptedEnvironmentConditions";

export function saveAdoptedEnvironmentConditions(value: string) {
  try {
    window.sessionStorage.setItem(storageKey, value);
  } catch {
    // 一時的なフォーム引継ぎは、Storage を使えない WebView でも画面遷移を妨げない。
  }
}

export function takeAdoptedEnvironmentConditions() {
  try {
    const value = window.sessionStorage.getItem(storageKey);
    window.sessionStorage.removeItem(storageKey);
    return value ?? undefined;
  } catch {
    return undefined;
  }
}
