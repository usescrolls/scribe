import { describe, it, expect, vi, beforeEach } from "vitest"
import { mount, flushPromises } from "@vue/test-utils"
import App from "./App.vue"
import { mockAppService } from "./test/setup"

describe("App", () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockAppService.IsOnboardingCompleted.mockResolvedValue(true)
    mockAppService.GetVersion.mockResolvedValue("1.2.3")
    mockAppService.GetSkills.mockResolvedValue([])
    mockAppService.GetWorkspaces.mockResolvedValue([
      { name: "default", description: "", skills: [], isActive: true },
    ])
    mockAppService.GetAgentStatus.mockResolvedValue([])
  })

  async function mountApp() {
    const wrapper = mount(App)
    await flushPromises()
    return wrapper
  }

  describe("onboarding", () => {
    it("shows onboarding when not completed", async () => {
      mockAppService.IsOnboardingCompleted.mockResolvedValue(false)

      const wrapper = await mountApp()

      expect(wrapper.findComponent({ name: "OnboardingWizard" }).exists()).toBe(
        true,
      )
      expect(wrapper.find(".app").exists()).toBe(false)
    })

    it("shows main app when onboarding completed", async () => {
      const wrapper = await mountApp()

      expect(wrapper.find(".app").exists()).toBe(true)
      expect(wrapper.findComponent({ name: "OnboardingWizard" }).exists()).toBe(
        false,
      )
    })

    it("shows main app on onboarding check error", async () => {
      mockAppService.IsOnboardingCompleted.mockRejectedValue(
        new Error("Failed"),
      )

      const wrapper = await mountApp()

      expect(wrapper.find(".app").exists()).toBe(true)
    })
  })

  describe("main app", () => {
    it("displays version", async () => {
      const wrapper = await mountApp()

      expect(wrapper.find(".version").text()).toBe("v1.2.3")
    })

    it("shows workspace tab by default", async () => {
      const wrapper = await mountApp()

      const tabs = wrapper.findAll(".tab")
      expect(tabs[0].classes()).toContain("active")
      expect(wrapper.findComponent({ name: "SkillList" }).exists()).toBe(true)
    })

    it("switches to browse tab on click", async () => {
      const wrapper = await mountApp()

      const browseTab = wrapper.findAll(".tab")[1]
      await browseTab.trigger("click")

      expect(browseTab.classes()).toContain("active")
      expect(wrapper.findComponent({ name: "BrowseSkills" }).exists()).toBe(
        true,
      )
    })

    it("switches back to workspace tab", async () => {
      const wrapper = await mountApp()

      // Go to browse
      await wrapper.findAll(".tab")[1].trigger("click")
      // Go back to workspace
      await wrapper.findAll(".tab")[0].trigger("click")

      expect(wrapper.findComponent({ name: "SkillList" }).exists()).toBe(true)
    })
  })
})
