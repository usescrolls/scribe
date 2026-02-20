import { describe, it, expect, vi, beforeEach } from "vitest"
import { mockAppService } from "../test/setup"
import { useLogger, logger } from "./useLogger"

describe("useLogger", () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockAppService.Log.mockResolvedValue(undefined)
  })

  it("sends info log to backend", () => {
    const log = useLogger()
    log.info("test message")
    expect(mockAppService.Log).toHaveBeenCalledWith("info", "test message")
  })

  it("sends debug log to backend", () => {
    const log = useLogger()
    log.debug("debug message")
    expect(mockAppService.Log).toHaveBeenCalledWith("debug", "debug message")
  })

  it("sends warn log to backend", () => {
    const log = useLogger()
    log.warn("warning")
    expect(mockAppService.Log).toHaveBeenCalledWith("warn", "warning")
  })

  it("sends error log to backend", () => {
    const log = useLogger()
    log.error("error occurred")
    expect(mockAppService.Log).toHaveBeenCalledWith("error", "error occurred")
  })

  it("prepends prefix when provided", () => {
    const log = useLogger("MyComponent")
    log.info("something happened")
    expect(mockAppService.Log).toHaveBeenCalledWith(
      "info",
      "[MyComponent] something happened",
    )
  })

  it("does not throw when backend call fails", () => {
    mockAppService.Log.mockRejectedValue(new Error("connection lost"))
    const log = useLogger()
    expect(() => log.error("test")).not.toThrow()
  })

  it("singleton logger works", () => {
    logger.info("singleton test")
    expect(mockAppService.Log).toHaveBeenCalledWith("info", "singleton test")
  })
})
