import { describe, it, expect, vi, beforeEach } from "vitest"
import { mount } from "@vue/test-utils"
import { nextTick } from "vue"
import WorkspaceSwitchInfoModal from "./WorkspaceSwitchInfoModal.vue"
import {
  showSwitchInfoModal,
  switchInfoWorkspaceName,
} from "../composables/useWorkspaces"

// Mock localStorage
const localStorageMock = (() => {
  let store: Record<string, string> = {}
  return {
    getItem: vi.fn((key: string) => store[key] ?? null),
    setItem: vi.fn((key: string, value: string) => {
      store[key] = value
    }),
    removeItem: vi.fn((key: string) => {
      delete store[key]
    }),
    clear: vi.fn(() => {
      store = {}
    }),
  }
})()

Object.defineProperty(globalThis, "localStorage", { value: localStorageMock })

describe("WorkspaceSwitchInfoModal", () => {
  function mountModal() {
    return mount(WorkspaceSwitchInfoModal, {
      global: {
        stubs: {
          Teleport: true,
        },
      },
    })
  }

  beforeEach(() => {
    vi.clearAllMocks()
    localStorageMock.clear()
    showSwitchInfoModal.value = false
    switchInfoWorkspaceName.value = ""
  })

  async function mountVisible() {
    const wrapper = mountModal()
    switchInfoWorkspaceName.value = "my-workspace"
    showSwitchInfoModal.value = true
    await nextTick()
    return wrapper
  }

  describe("visibility", () => {
    it("is hidden by default", () => {
      const wrapper = mountModal()

      expect(wrapper.find(".info-backdrop").exists()).toBe(false)
    })

    it("shows when showSwitchInfoModal becomes true", async () => {
      const wrapper = mountModal()

      switchInfoWorkspaceName.value = "my-workspace"
      showSwitchInfoModal.value = true
      await nextTick()

      expect(wrapper.find(".info-backdrop").exists()).toBe(true)
    })
  })

  describe("rendering", () => {
    it("renders title", async () => {
      const wrapper = await mountVisible()

      expect(wrapper.find(".info-header h3").text()).toBe("Workspace activated")
    })

    it("displays workspace name in message", async () => {
      const wrapper = await mountVisible()

      expect(wrapper.find(".info-message").text()).toContain("my-workspace")
    })

    it("mentions opening a new chat window", async () => {
      const wrapper = await mountVisible()

      expect(wrapper.find(".info-message").text()).toContain(
        "open a new coding agent chat window",
      )
    })

    it("shows Cursor-specific callout", async () => {
      const wrapper = await mountVisible()

      expect(wrapper.find(".info-callout").text()).toContain("Cursor")
      expect(wrapper.find(".info-callout").text()).toContain("new chat tab")
    })

    it("renders Got it button", async () => {
      const wrapper = await mountVisible()

      expect(wrapper.find(".btn-got-it").text()).toBe("Got it")
    })

    it("renders close button", async () => {
      const wrapper = await mountVisible()

      expect(wrapper.find(".close-btn").exists()).toBe(true)
    })

    it('renders "Don\'t show this again" checkbox', async () => {
      const wrapper = await mountVisible()

      expect(wrapper.find(".dont-show-label").exists()).toBe(true)
      expect(wrapper.find(".dont-show-label").text()).toContain(
        "Don't show this again",
      )
    })
  })

  describe("interactions", () => {
    it("hides on Got it button click", async () => {
      const wrapper = await mountVisible()

      await wrapper.find(".btn-got-it").trigger("click")
      await nextTick()

      expect(wrapper.find(".info-backdrop").exists()).toBe(false)
    })

    it("hides on close button click", async () => {
      const wrapper = await mountVisible()

      await wrapper.find(".close-btn").trigger("click")
      await nextTick()

      expect(wrapper.find(".info-backdrop").exists()).toBe(false)
    })

    it("hides on backdrop click", async () => {
      const wrapper = await mountVisible()

      await wrapper.find(".info-backdrop").trigger("click")
      await nextTick()

      expect(wrapper.find(".info-backdrop").exists()).toBe(false)
    })

    it("does not hide when clicking modal content", async () => {
      const wrapper = await mountVisible()

      await wrapper.find(".info-modal").trigger("click")
      await nextTick()

      expect(wrapper.find(".info-backdrop").exists()).toBe(true)
    })

    it("hides on Escape key", async () => {
      const wrapper = await mountVisible()

      document.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape" }))
      await nextTick()

      expect(wrapper.find(".info-backdrop").exists()).toBe(false)
      wrapper.unmount()
    })

    it("resets checkbox when shown again", async () => {
      const wrapper = mountModal()

      // Show first time
      switchInfoWorkspaceName.value = "ws-1"
      showSwitchInfoModal.value = true
      await nextTick()

      // Check the checkbox
      const checkbox = wrapper.find('.dont-show-label input[type="checkbox"]')
      await checkbox.setValue(true)
      expect((checkbox.element as HTMLInputElement).checked).toBe(true)

      // Dismiss and manually reset shared state (transition callback doesn't fire in test)
      await wrapper.find(".btn-got-it").trigger("click")
      showSwitchInfoModal.value = false
      await nextTick()

      // Show again
      showSwitchInfoModal.value = true
      await nextTick()

      const newCheckbox = wrapper.find(
        '.dont-show-label input[type="checkbox"]',
      )
      expect((newCheckbox.element as HTMLInputElement).checked).toBe(false)
    })
  })
})
