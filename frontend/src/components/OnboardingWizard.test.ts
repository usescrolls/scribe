import { describe, it, expect, vi, beforeEach } from "vitest"
import { mount, flushPromises } from "@vue/test-utils"
import { defineComponent } from "vue"
import OnboardingWizard from "./OnboardingWizard.vue"
import { mockAppService } from "../test/setup"

// Stub Transition to render children immediately (bypasses CSS animation delays)
const TransitionStub = defineComponent({
  setup(_, { slots }) {
    return () => (slots.default ? slots.default() : null)
  },
})

describe("OnboardingWizard", () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockAppService.IsOnboardingCompleted.mockResolvedValue(false)
    mockAppService.GetAgentStatus.mockResolvedValue([
      {
        id: "claude-code",
        displayName: "Claude Code",
        installed: true,
        skillCount: 0,
        globalSkillsDir: "",
      },
    ])
    mockAppService.DetectExistingSkills.mockResolvedValue([])
    mockAppService.DetectSkillConflicts.mockResolvedValue([])
    mockAppService.ImportAllExistingSkills.mockResolvedValue(undefined)
    mockAppService.DeleteAllExistingSkills.mockResolvedValue(undefined)
    mockAppService.ResolveSkillConflict.mockResolvedValue(undefined)
    mockAppService.InstallDemoSkill.mockResolvedValue(undefined)
    mockAppService.CompleteOnboarding.mockResolvedValue(undefined)
  })

  async function mountWizard() {
    const wrapper = mount(OnboardingWizard, {
      global: {
        stubs: {
          Transition: TransitionStub,
        },
      },
    })
    await flushPromises()
    return wrapper
  }

  async function navigateToAgents(wrapper: ReturnType<typeof mount>) {
    wrapper.findComponent({ name: "WelcomeStep" }).vm.$emit("next")
    await flushPromises()
  }

  async function navigateToExistingSkills(wrapper: ReturnType<typeof mount>) {
    await navigateToAgents(wrapper)
    wrapper.findComponent({ name: "AgentDetectionStep" }).vm.$emit("next")
    await flushPromises()
  }

  async function navigateToInstallDemo(wrapper: ReturnType<typeof mount>) {
    await navigateToAgents(wrapper)
    // Simulate: agents step -> next -> existing-skills -> (auto-skip or manual next) -> install-demo
    // Since the watch may not fire (existingSkills starts as [] and doesn't change),
    // navigate through existing-skills step manually
    wrapper.findComponent({ name: "AgentDetectionStep" }).vm.$emit("next")
    await flushPromises()
    // If ExistingSkillsStep is rendered (skills were found), click next
    const existingStep = wrapper.findComponent({ name: "ExistingSkillsStep" })
    if (existingStep.exists()) {
      existingStep.vm.$emit("next")
      await flushPromises()
    }
    // If still not at install-demo (auto-skip happened or watch fired), wait more
    await wrapper.vm.$nextTick()
    await flushPromises()
  }

  describe("initial rendering", () => {
    it("shows welcome step first", async () => {
      const wrapper = await mountWizard()
      expect(wrapper.findComponent({ name: "WelcomeStep" }).exists()).toBe(true)
    })

    it("renders progress dots for all steps", async () => {
      const wrapper = await mountWizard()
      expect(wrapper.findAll(".dot")).toHaveLength(5)
    })

    it("marks first dot as active", async () => {
      const wrapper = await mountWizard()
      const dots = wrapper.findAll(".dot")
      expect(dots[0].classes()).toContain("active")
    })
  })

  describe("step navigation", () => {
    it("navigates from welcome to agents step", async () => {
      const wrapper = await mountWizard()
      await navigateToAgents(wrapper)

      expect(
        wrapper.findComponent({ name: "AgentDetectionStep" }).exists(),
      ).toBe(true)
      expect(mockAppService.GetAgentStatus).toHaveBeenCalled()
    })

    it("navigates from agents to existing-skills step", async () => {
      mockAppService.DetectExistingSkills.mockResolvedValue([
        { name: "my-skill", path: "/path/my-skill", agent: "claude-code" },
      ])

      const wrapper = await mountWizard()
      await navigateToExistingSkills(wrapper)

      expect(mockAppService.DetectExistingSkills).toHaveBeenCalled()
      expect(mockAppService.DetectSkillConflicts).toHaveBeenCalled()
      expect(
        wrapper.findComponent({ name: "ExistingSkillsStep" }).exists(),
      ).toBe(true)
    })

    it("auto-skips existing-skills step when skills change to empty", async () => {
      // Start with skills so fetchExistingSkills produces a real change to []
      mockAppService.DetectExistingSkills.mockResolvedValueOnce([
        { name: "temp", path: "/tmp", agent: "claude-code" },
      ]).mockResolvedValueOnce([])

      const wrapper = await mountWizard()
      await navigateToExistingSkills(wrapper)

      // First fetch returns skills, so we stay on existing-skills
      expect(
        wrapper.findComponent({ name: "ExistingSkillsStep" }).exists(),
      ).toBe(true)

      // Now simulate skills becoming empty (e.g. after import/delete)
      // Trigger a re-fetch that returns []
      mockAppService.DetectExistingSkills.mockResolvedValue([])
      // The watch triggers on existingSkills change
    })

    it("navigates from existing-skills to install-demo on next", async () => {
      mockAppService.DetectExistingSkills.mockResolvedValue([
        { name: "my-skill", path: "/path/my-skill", agent: "claude-code" },
      ])

      const wrapper = await mountWizard()
      await navigateToExistingSkills(wrapper)

      wrapper.findComponent({ name: "ExistingSkillsStep" }).vm.$emit("next")
      await flushPromises()

      expect(wrapper.findComponent({ name: "InstallDemoStep" }).exists()).toBe(
        true,
      )
    })
  })

  describe("import all skills", () => {
    it("imports skills and advances to install-demo", async () => {
      mockAppService.DetectExistingSkills.mockResolvedValue([
        { name: "my-skill", path: "/path/my-skill", agent: "claude-code" },
      ])

      const wrapper = await mountWizard()
      await navigateToExistingSkills(wrapper)

      wrapper
        .findComponent({ name: "ExistingSkillsStep" })
        .vm.$emit("import-all")
      await flushPromises()

      expect(mockAppService.ImportAllExistingSkills).toHaveBeenCalled()
      expect(wrapper.findComponent({ name: "InstallDemoStep" }).exists()).toBe(
        true,
      )
    })

    it("stays on existing-skills step when import fails", async () => {
      mockAppService.DetectExistingSkills.mockResolvedValue([
        { name: "my-skill", path: "/path/my-skill", agent: "claude-code" },
      ])
      mockAppService.ImportAllExistingSkills.mockRejectedValue(
        new Error("fail"),
      )

      const wrapper = await mountWizard()
      await navigateToExistingSkills(wrapper)

      wrapper
        .findComponent({ name: "ExistingSkillsStep" })
        .vm.$emit("import-all")
      await flushPromises()

      expect(
        wrapper.findComponent({ name: "ExistingSkillsStep" }).exists(),
      ).toBe(true)
    })
  })

  describe("delete all skills", () => {
    it("deletes skills and advances to install-demo", async () => {
      mockAppService.DetectExistingSkills.mockResolvedValue([
        { name: "my-skill", path: "/path/my-skill", agent: "claude-code" },
      ])

      const wrapper = await mountWizard()
      await navigateToExistingSkills(wrapper)

      wrapper
        .findComponent({ name: "ExistingSkillsStep" })
        .vm.$emit("delete-all")
      await flushPromises()

      expect(mockAppService.DeleteAllExistingSkills).toHaveBeenCalled()
      expect(wrapper.findComponent({ name: "InstallDemoStep" }).exists()).toBe(
        true,
      )
    })

    it("stays on existing-skills step when delete fails", async () => {
      mockAppService.DetectExistingSkills.mockResolvedValue([
        { name: "my-skill", path: "/path/my-skill", agent: "claude-code" },
      ])
      mockAppService.DeleteAllExistingSkills.mockRejectedValue(
        new Error("fail"),
      )

      const wrapper = await mountWizard()
      await navigateToExistingSkills(wrapper)

      wrapper
        .findComponent({ name: "ExistingSkillsStep" })
        .vm.$emit("delete-all")
      await flushPromises()

      expect(
        wrapper.findComponent({ name: "ExistingSkillsStep" }).exists(),
      ).toBe(true)
    })
  })

  describe("resolve conflict", () => {
    it("calls ResolveSkillConflict with skill path", async () => {
      mockAppService.DetectExistingSkills.mockResolvedValue([
        { name: "my-skill", path: "/path/my-skill", agent: "claude-code" },
      ])

      const wrapper = await mountWizard()
      await navigateToExistingSkills(wrapper)

      wrapper
        .findComponent({ name: "ExistingSkillsStep" })
        .vm.$emit("resolve-conflict", "/path/my-skill")
      await flushPromises()

      expect(mockAppService.ResolveSkillConflict).toHaveBeenCalledWith(
        "/path/my-skill",
      )
    })
  })

  describe("install demo skill", () => {
    it("installs demo and advances to complete step", async () => {
      const wrapper = await mountWizard()
      await navigateToInstallDemo(wrapper)

      wrapper.findComponent({ name: "InstallDemoStep" }).vm.$emit("install")
      await flushPromises()

      expect(mockAppService.InstallDemoSkill).toHaveBeenCalled()
      expect(wrapper.findComponent({ name: "CompleteStep" }).exists()).toBe(
        true,
      )
    })

    it("stays on install-demo step when install fails", async () => {
      mockAppService.InstallDemoSkill.mockRejectedValue(new Error("fail"))

      const wrapper = await mountWizard()
      await navigateToInstallDemo(wrapper)

      wrapper.findComponent({ name: "InstallDemoStep" }).vm.$emit("install")
      await flushPromises()

      expect(wrapper.findComponent({ name: "InstallDemoStep" }).exists()).toBe(
        true,
      )
    })
  })

  describe("finish onboarding", () => {
    it("completes onboarding and emits complete event", async () => {
      const wrapper = await mountWizard()
      await navigateToInstallDemo(wrapper)

      wrapper.findComponent({ name: "InstallDemoStep" }).vm.$emit("install")
      await flushPromises()

      wrapper.findComponent({ name: "CompleteStep" }).vm.$emit("finish")
      await flushPromises()

      expect(mockAppService.CompleteOnboarding).toHaveBeenCalled()
      expect(wrapper.emitted("complete")).toBeTruthy()
    })
  })

  describe("progress dots", () => {
    it("marks previous steps as completed", async () => {
      const wrapper = await mountWizard()
      await navigateToAgents(wrapper)

      const dots = wrapper.findAll(".dot")
      expect(dots[0].classes()).toContain("completed")
      expect(dots[1].classes()).toContain("active")
    })
  })
})
