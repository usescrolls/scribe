import { ref, onMounted, onUnmounted } from 'vue'
import { AppService } from '../bindings/scribe'
import { Events } from '@wailsio/runtime'
import type { PluginInfo } from '../types/plugin'

export function usePlugins() {
  const plugins = ref<PluginInfo[]>([])
  const loading = ref(true)
  const error = ref<string | null>(null)
  let unsubscribe: (() => void) | null = null

  async function fetchPlugins() {
    try {
      loading.value = true
      error.value = null
      plugins.value = await AppService.GetPlugins()
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to load plugins'
    } finally {
      loading.value = false
    }
  }

  async function uninstall(name: string): Promise<boolean> {
    console.log('[usePlugins] uninstall called with:', name)
    try {
      console.log('[usePlugins] Calling AppService.UninstallPlugin...')
      await AppService.UninstallPlugin(name)
      console.log('[usePlugins] UninstallPlugin succeeded, refreshing plugins...')
      await fetchPlugins()
      return true
    } catch (e) {
      console.error('[usePlugins] UninstallPlugin failed:', e)
      error.value = e instanceof Error ? e.message : 'Failed to uninstall plugin'
      return false
    }
  }

  onMounted(() => {
    fetchPlugins()
    unsubscribe = Events.On('plugins-updated', fetchPlugins)
  })

  onUnmounted(() => {
    if (unsubscribe) {
      unsubscribe()
    }
  })

  return { plugins, loading, error, fetchPlugins, uninstall }
}
