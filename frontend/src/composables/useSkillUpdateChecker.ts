import { ref } from "vue"
import { AppService } from "../bindings/scribe"
import { useLogger } from "./useLogger"
import type { SourceGroupCheckResult } from "../types/skill"

const log = useLogger("SkillUpdateChecker")

const CHECK_INTERVAL_MS = 60 * 60 * 1000 // 1 hour
const INITIAL_DELAY_MS = 15_000 // 15 seconds after app start

// Shared state — all consumers see the same data
const sourceUpdates = ref<Record<string, SourceGroupCheckResult>>({})
const checking = ref(false)

let pollingStarted = false
let intervalId: ReturnType<typeof setInterval> | null = null

async function checkForSourceUpdates() {
  if (checking.value) return
  try {
    checking.value = true
    const results = await AppService.CheckSourceGroupUpdates()
    if (results) {
      sourceUpdates.value = (results as Record<string, SourceGroupCheckResult>)
    }

    const withUpdates = Object.values(sourceUpdates.value).filter(
      (r) => r.hasUpdates,
    ).length
    log.info(
      `skill update check complete: ${withUpdates} source(s) with updates`,
    )
  } catch (e) {
    log.warn(`skill update check failed: ${e instanceof Error ? e.message : e}`)
  } finally {
    checking.value = false
  }
}

function hasUpdates(source: string): boolean {
  return sourceUpdates.value[source]?.hasUpdates ?? false
}

function getUpdateInfo(source: string): SourceGroupCheckResult | null {
  return sourceUpdates.value[source] ?? null
}

function clearUpdate(source: string) {
  const current = sourceUpdates.value[source]
  if (current) {
    sourceUpdates.value = {
      ...sourceUpdates.value,
      [source]: { ...current, hasUpdates: false, updatedSkillNames: [] },
    }
  }
}

function startPolling() {
  if (pollingStarted) return
  pollingStarted = true
  // Initial check after a short delay (don't block app startup)
  setTimeout(checkForSourceUpdates, INITIAL_DELAY_MS)
  // Then check every hour
  intervalId = setInterval(checkForSourceUpdates, CHECK_INTERVAL_MS)
}

function stopPolling() {
  if (intervalId) {
    clearInterval(intervalId)
    intervalId = null
    pollingStarted = false
  }
}

export function useSkillUpdateChecker() {
  return {
    sourceUpdates,
    checking,
    checkForSourceUpdates,
    hasUpdates,
    getUpdateInfo,
    clearUpdate,
    startPolling,
    stopPolling,
  }
}
