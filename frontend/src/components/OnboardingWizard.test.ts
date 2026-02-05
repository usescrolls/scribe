import { describe, it, expect, vi, beforeEach } from "vitest"
import { mount, flushPromises } from "@vue/test-utils"
import OnboardingWizard from "./OnboardingWizard.vue"
import { mockAppService } from "../test/setup"

describe("OnboardingWizard", () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockAppService.IsOnboardingCompleted.mockResolvedValue(false)
    mockAppService.GetAgentStatus.mockResolvedValue([])
    mockAppService.DetectExistingSkills.mockResolvedValue([])
    mockAppService.DetectSkillConflicts.mockResolvedValue([])
    mockAppService.CompleteOnboarding.mockResolvedValue(undefined)
    mockAppService.InstallDemoSkill.mockResolvedValue(undefined)
  })

  async function mountWizard() {
    const wrapper = mount(OnboardingWizard)
    await flushPromises()
    return wrapper
  }

  describe("rendering", () => {
    it("shows welcome step by default", async () => {
      const wrapper = await mountWizard()

      expect(wrapper.findComponent({ name: "WelcomeStep" }).exists()).toBe(true)
    })

    it("renders progress dots", async () => {
      const wrapper = await mountWizard()

      const dots = wrapper.findAll(".dot")
      expect(dots).toHaveLength(5)
    })

    it("marks first dot as active", async () => {
      const wrapper = await mountWizard()

      const dots = wrapper.findAll(".dot")
      expect(dots[0].classes()).toContain("active")
    })
  })

  describe("navigation", () => {
    it("navigates to agents step on welcome next", async () => {
      mockAppService.GetAgentStatus.mockResolvedValue([
        {
          id: "claude-code",
          displayName: "Claude Code",
          installed: true,
          skillCount: 0,
          globalSkillsDir: "",
        },
      ])

      const wrapper = await mountWizard()
      const welcomeStep = wrapper.findComponent({ name: "WelcomeStep" })
      await welcomeStep.vm.$emit("next")
      await flushPromises()

      expect(
        wrapper.findComponent({ name: "AgentDetectionStep" }).exists(),
      ).toBe(true)
    })

    it("marks previous dots as completed", async () => {
      mockAppService.GetAgentStatus.mockResolvedValue([
        {
          id: "claude-code",
          displayName: "Claude Code",
          installed: true,
          skillCount: 0,
          globalSkillsDir: "",
        },
      ])

      const wrapper = await mountWizard()
      const welcomeStep = wrapper.findComponent({ name: "WelcomeStep" })
      await welcomeStep.vm.$emit("next")
      await flushPromises()

      const dots = wrapper.findAll(".dot")
      expect(dots[0].classes()).toContain("completed")
      expect(dots[1].classes()).toContain("active")
    })
  })
})
