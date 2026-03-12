import { describe, it, expect, vi, beforeEach, afterEach } from "vitest"
import { useUpdateChecker } from "./useUpdateChecker"
import { mockAppService, mockBrowser } from "../test/setup"

const mockUpdateAvailable = {
  currentVersion: "1.0.0",
  latestVersion: "1.1.0",
  updateAvailable: true,
  releaseURL: "https://github.com/usescrolls/scribe/releases/tag/v1.1.0",
  publishedAt: "2026-02-20T00:00:00Z",
}

const mockUpToDate = {
  currentVersion: "1.0.0",
  latestVersion: "1.0.0",
  updateAvailable: false,
  releaseURL: "",
  publishedAt: "",
}

describe("useUpdateChecker", () => {
  let checker: ReturnType<typeof useUpdateChecker>

  beforeEach(() => {
    vi.useFakeTimers()
    checker = useUpdateChecker()
    // Reset shared module-level state between tests
    checker.updateInfo.value = null
    checker.loading.value = false
    checker.notificationsDisabled.value = false
    checker.showToast.value = false
    checker.installMethod.value = null
    checker.upgrading.value = false
    checker.upgradeError.value = null
    checker.upgradeSuccess.value = false
    checker.stopPolling()

    mockAppService.CheckForAppUpdate.mockResolvedValue(mockUpdateAvailable)
    mockAppService.IsUpdateNotificationsDisabled.mockResolvedValue(false)
    mockAppService.SetUpdateNotificationsDisabled.mockResolvedValue(undefined)
    mockAppService.GetInstallMethod.mockResolvedValue("binary")
    mockAppService.UpgradeApp.mockResolvedValue(undefined)
  })

  afterEach(() => {
    checker.stopPolling()
    vi.useRealTimers()
  })

  describe("shared state", () => {
    it("returns the same refs across multiple calls", () => {
      const a = useUpdateChecker()
      const b = useUpdateChecker()

      expect(a.updateInfo).toBe(b.updateInfo)
      expect(a.loading).toBe(b.loading)
      expect(a.showToast).toBe(b.showToast)
      expect(a.notificationsDisabled).toBe(b.notificationsDisabled)
    })
  })

  describe("checkForUpdate", () => {
    it("fetches update info and sets state", async () => {
      await checker.checkForUpdate()

      expect(mockAppService.CheckForAppUpdate).toHaveBeenCalled()
      expect(checker.updateInfo.value).toEqual(mockUpdateAvailable)
    })

    it("sets loading during check", async () => {
      let resolve: (v: unknown) => void
      mockAppService.CheckForAppUpdate.mockReturnValue(
        new Promise((r) => {
          resolve = r
        }),
      )

      const promise = checker.checkForUpdate()
      expect(checker.loading.value).toBe(true)

      resolve!(mockUpToDate)
      await promise
      expect(checker.loading.value).toBe(false)
    })

    it("shows toast when update available and notifications enabled", async () => {
      await checker.checkForUpdate()
      expect(checker.showToast.value).toBe(true)
    })

    it("does not show toast when notifications disabled", async () => {
      checker.notificationsDisabled.value = true
      await checker.checkForUpdate()
      expect(checker.showToast.value).toBe(false)
    })

    it("does not show toast when up to date", async () => {
      mockAppService.CheckForAppUpdate.mockResolvedValue(mockUpToDate)
      await checker.checkForUpdate()
      expect(checker.showToast.value).toBe(false)
    })

    it("handles API errors gracefully", async () => {
      mockAppService.CheckForAppUpdate.mockRejectedValue(
        new Error("Network error"),
      )
      await checker.checkForUpdate()

      expect(checker.updateInfo.value).toBeNull()
      expect(checker.loading.value).toBe(false)
    })
  })

  describe("setNotificationsDisabled", () => {
    it("persists preference and updates state", async () => {
      await checker.setNotificationsDisabled(true)

      expect(
        mockAppService.SetUpdateNotificationsDisabled,
      ).toHaveBeenCalledWith(true)
      expect(checker.notificationsDisabled.value).toBe(true)
    })

    it("hides toast when disabling notifications", async () => {
      checker.showToast.value = true
      await checker.setNotificationsDisabled(true)
      expect(checker.showToast.value).toBe(false)
    })

    it("keeps toast visible when enabling notifications", async () => {
      checker.showToast.value = true
      checker.notificationsDisabled.value = true
      await checker.setNotificationsDisabled(false)
      expect(checker.showToast.value).toBe(true)
    })
  })

  describe("dismissToast", () => {
    it("hides the toast", () => {
      checker.showToast.value = true
      checker.dismissToast()
      expect(checker.showToast.value).toBe(false)
    })
  })

  describe("openReleasePage", () => {
    it("opens release URL in browser", async () => {
      await checker.checkForUpdate()
      checker.openReleasePage()

      expect(mockBrowser.OpenURL).toHaveBeenCalledWith(
        "https://github.com/usescrolls/scribe/releases/tag/v1.1.0",
      )
    })

    it("does nothing when no update info", () => {
      checker.openReleasePage()
      expect(mockBrowser.OpenURL).not.toHaveBeenCalled()
    })
  })

  describe("startPolling", () => {
    it("loads notification preference", async () => {
      await checker.startPolling()
      expect(mockAppService.IsUpdateNotificationsDisabled).toHaveBeenCalled()
    })

    it("checks for update after 5s delay", async () => {
      await checker.startPolling()
      expect(mockAppService.CheckForAppUpdate).not.toHaveBeenCalled()

      await vi.advanceTimersByTimeAsync(5000)
      expect(mockAppService.CheckForAppUpdate).toHaveBeenCalledTimes(1)
    })

    it("checks periodically every hour", async () => {
      await checker.startPolling()

      await vi.advanceTimersByTimeAsync(5000)
      expect(mockAppService.CheckForAppUpdate).toHaveBeenCalledTimes(1)

      await vi.advanceTimersByTimeAsync(60 * 60 * 1000)
      expect(mockAppService.CheckForAppUpdate).toHaveBeenCalledTimes(2)
    })

    it("only starts once even if called multiple times", async () => {
      await checker.startPolling()
      await checker.startPolling()
      await checker.startPolling()

      await vi.advanceTimersByTimeAsync(5000)
      expect(mockAppService.CheckForAppUpdate).toHaveBeenCalledTimes(1)
    })
  })

  describe("stopPolling", () => {
    it("stops periodic checks", async () => {
      await checker.startPolling()
      await vi.advanceTimersByTimeAsync(5000)
      expect(mockAppService.CheckForAppUpdate).toHaveBeenCalledTimes(1)

      checker.stopPolling()
      mockAppService.CheckForAppUpdate.mockClear()

      await vi.advanceTimersByTimeAsync(60 * 60 * 1000)
      expect(mockAppService.CheckForAppUpdate).not.toHaveBeenCalled()
    })

    it("allows restarting after stop", async () => {
      await checker.startPolling()
      checker.stopPolling()

      await checker.startPolling()
      await vi.advanceTimersByTimeAsync(5000)
      expect(mockAppService.CheckForAppUpdate).toHaveBeenCalled()
    })
  })

  describe("installMethod", () => {
    it("detects install method on startPolling", async () => {
      mockAppService.GetInstallMethod.mockResolvedValue("binary")
      await checker.startPolling()
      expect(checker.installMethod.value).toBe("binary")
    })

    it("handles detection errors gracefully", async () => {
      mockAppService.GetInstallMethod.mockRejectedValue(new Error("fail"))
      await checker.startPolling()
      expect(checker.installMethod.value).toBe("unknown")
    })
  })

  describe("upgradeApp", () => {
    it("performs upgrade and sets success", async () => {
      mockAppService.UpgradeApp.mockResolvedValue({
        updated: true,
        oldVersion: "1.0.0",
        newVersion: "v1.1.0",
        installMethod: "binary",
        message: "upgraded",
      })
      await checker.upgradeApp()
      expect(checker.upgradeSuccess.value).toBe(true)
      expect(checker.upgradeError.value).toBeNull()
    })

    it("handles upgrade errors", async () => {
      mockAppService.UpgradeApp.mockRejectedValue(
        new Error("permission denied"),
      )
      await checker.upgradeApp()
      expect(checker.upgradeError.value).toBe("permission denied")
      expect(checker.upgradeSuccess.value).toBe(false)
    })

    it("does not call checkForUpdate after successful upgrade", async () => {
      mockAppService.UpgradeApp.mockResolvedValue({
        updated: true,
        oldVersion: "1.0.0",
        newVersion: "v1.1.0",
        installMethod: "binary",
        message: "upgraded",
      })
      mockAppService.CheckForAppUpdate.mockClear()
      await checker.upgradeApp()
      expect(mockAppService.CheckForAppUpdate).not.toHaveBeenCalled()
    })

    it("sets upgrading flag during operation", async () => {
      let resolve: (v: unknown) => void
      mockAppService.UpgradeApp.mockReturnValue(
        new Promise((r) => {
          resolve = r
        }),
      )
      const promise = checker.upgradeApp()
      expect(checker.upgrading.value).toBe(true)

      resolve!({ updated: false })
      await promise
      expect(checker.upgrading.value).toBe(false)
    })
  })

  describe("restartApp", () => {
    it("calls RestartApp on the backend", async () => {
      mockAppService.RestartApp.mockResolvedValue(undefined)
      await checker.restartApp()
      expect(mockAppService.RestartApp).toHaveBeenCalled()
    })

    it("handles restart errors gracefully", async () => {
      mockAppService.RestartApp.mockRejectedValue(new Error("restart failed"))
      await checker.restartApp()
      // Should not throw
    })
  })
})
