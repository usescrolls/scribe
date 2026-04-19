export function sourceGroupKey(source?: string, sourceType?: string): string {
  const normalizedSource = source || "unknown"
  const normalizedType = sourceType || "local"
  return `${normalizedType}:${normalizedSource}`
}
