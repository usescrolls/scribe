import { AppService } from "../bindings/scribe"

export type LogLevel = "debug" | "info" | "warn" | "error"

/**
 * Creates a logger that sends messages to the Go backend
 * for unified logging to ~/.scribe/scribe.log
 */
export function useLogger(prefix?: string) {
  function log(level: LogLevel, message: string) {
    try {
      const formatted = prefix ? `[${prefix}] ${message}` : message
      const result = AppService.Log(level, formatted)
      if (result && typeof result.catch === "function") {
        result.catch(() => {})
      }
    } catch {
      // Never let logging break the app
    }
  }

  return {
    debug: (message: string) => log("debug", message),
    info: (message: string) => log("info", message),
    warn: (message: string) => log("warn", message),
    error: (message: string) => log("error", message),
  }
}

export const logger = useLogger()
