import { ref } from "vue"
import { AppService } from "../bindings/scribe"
import { Browser } from "@wailsio/runtime"
import { useLogger } from "./useLogger"

const log = useLogger("UpdateChecker")

const CHECK_INTERVAL_MS = 60 * 60 * 1000 // 1 hour

export interface UpdateInfo {
  currentVersion: string
  latestVersion: string
  updateAvailable: boolean
  releaseURL: string
  publishedAt: string
}

// Shared state — all consumers see the same data
const updateInfo = ref<UpdateInfo | null>(null)
const loading = ref(false)
const notificationsDisabled = ref(false)
const showToast = ref(false)
const installMethod = ref<string | null>(null)
const upgrading = ref(false)
const upgradeError = ref<string | null>(null)
const upgradeSuccess = ref(false)

let pollingStarted = false
let intervalId: ReturnType<typeof setInterval> | null = null

async function loadPreference() {
  try {
    notificationsDisabled.value =
      await AppService.IsUpdateNotificationsDisabled()
  } catch (e) {
    log.error(
      `failed to load notification preference: ${
        e instanceof Error ? e.message : e
      }`,
    )
  }
}

async function checkForUpdate() {
  try {
    loading.value = true
    const info = await AppService.CheckForAppUpdate()
    if (info) {
      updateInfo.value = (info as UpdateInfo)
    }

    if (info?.updateAvailable && !notificationsDisabled.value) {
      showToast.value = true
    }
    log.info(
      `update check complete: ${
        info?.updateAvailable ? "update available" : "up to date"
      }`,
    )
  } catch (e) {
    log.warn(`update check failed: ${e instanceof Error ? e.message : e}`)
  } finally {
    loading.value = false
  }
}

async function setNotificationsDisabled(disabled: boolean) {
  try {
    await AppService.SetUpdateNotificationsDisabled(disabled)
    notificationsDisabled.value = disabled
    if (disabled) {
      showToast.value = false
    }
    log.info(`update notifications ${disabled ? "disabled" : "enabled"}`)
  } catch (e) {
    log.error(
      `failed to set notification preference: ${
        e instanceof Error ? e.message : e
      }`,
    )
  }
}

function dismissToast() {
  showToast.value = false
}

function openReleasePage() {
  if (updateInfo.value?.releaseURL) {
    Browser.OpenURL(updateInfo.value.releaseURL)
  }
}

async function detectInstallMethod() {
  try {
    installMethod.value = await AppService.GetInstallMethod()
    log.info(`install method: ${installMethod.value}`)
  } catch (e) {
    log.error(
      `failed to detect install method: ${e instanceof Error ? e.message : e}`,
    )
    installMethod.value = "unknown"
  }
}

async function upgradeApp() {
  try {
    upgrading.value = true
    upgradeError.value = null
    upgradeSuccess.value = false
    const result = await AppService.UpgradeApp()
    if (result?.updated) {
      upgradeSuccess.value = true
      await checkForUpdate()
      log.info(`upgrade complete: ${result.oldVersion} -> ${result.newVersion}`)
    } else {
      log.info("already up to date")
    }
  } catch (e) {
    upgradeError.value = e instanceof Error ? e.message : String(e)
    log.error(`upgrade failed: ${upgradeError.value}`)
  } finally {
    upgrading.value = false
  }
}

async function startPolling() {
  if (pollingStarted) return
  pollingStarted = true
  await loadPreference()
  await detectInstallMethod()
  // Initial check after a short delay (don't block app startup)
  setTimeout(checkForUpdate, 5000)
  // Then check every hour
  intervalId = setInterval(checkForUpdate, CHECK_INTERVAL_MS)
}

function stopPolling() {
  if (intervalId) {
    clearInterval(intervalId)
    intervalId = null
    pollingStarted = false
  }
}

export function useUpdateChecker() {
  return {
    updateInfo,
    loading,
    notificationsDisabled,
    showToast,
    installMethod,
    upgrading,
    upgradeError,
    upgradeSuccess,
    checkForUpdate,
    setNotificationsDisabled,
    dismissToast,
    openReleasePage,
    upgradeApp,
    startPolling,
    stopPolling,
  }
}
