const storageKey = "adoptedEnvironmentConditions";

export function saveAdoptedEnvironmentConditions(value: string) {
  window.sessionStorage.setItem(storageKey, value);
}

export function takeAdoptedEnvironmentConditions() {
  const value = window.sessionStorage.getItem(storageKey);
  window.sessionStorage.removeItem(storageKey);
  return value ?? undefined;
}
