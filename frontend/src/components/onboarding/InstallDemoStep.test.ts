import { describe, it, expect } from "vitest"
import { mount } from "@vue/test-utils"
import InstallDemoStep from "./InstallDemoStep.vue"
import type { AgentStatus } from "../../types/skill"

describe("InstallDemoStep", () => {
  const mockAgents: AgentStatus[] = [
    {
      id: "claude-code",
      displayName: "Claude Code",
      installed: true,
      skillCount: 0,
      globalSkillsDir: "",
    },
    {
      id: "cursor",
      displayName: "Cursor",
      installed: true,
      skillCount: 0,
      globalSkillsDir: "",
    },
  ]

  interface StepProps {
    installing: boolean
    installedAgents: AgentStatus[]
  }

  function mountStep(props: Partial<StepProps> = {}) {
    return mount(InstallDemoStep, {
      props: {
        installing: false,
        installedAgents: mockAgents,
        ...props,
      },
    })
  }

  it("renders title", () => {
    const wrapper = mountStep()

    expect(wrapper.find("h1").text()).toBe("Install Demo Skill")
  })

  it("shows agents that will be synced to", () => {
    const wrapper = mountStep()

    const badges = wrapper.findAll(".agent-badge")
    expect(badges).toHaveLength(2)
    expect(badges[0].text()).toBe("Claude Code")
    expect(badges[1].text()).toBe("Cursor")
  })

  it("shows skill preview card", () => {
    const wrapper = mountStep()

    expect(wrapper.find(".skill-name").text()).toBe("scribe-welcome")
    expect(wrapper.find(".skill-type").text()).toBe("Demo")
  })

  it("shows install button", () => {
    const wrapper = mountStep()

    expect(wrapper.find(".primary-button").text()).toBe("Install Demo Skill")
  })

  it("shows installing state", () => {
    const wrapper = mountStep({ installing: true })

    expect(wrapper.find(".primary-button").text()).toContain("Installing...")
    expect(wrapper.find(".primary-button").attributes("disabled")).toBeDefined()
  })

  it("emits install on button click", async () => {
    const wrapper = mountStep()

    await wrapper.find(".primary-button").trigger("click")

    expect(wrapper.emitted("install")).toBeTruthy()
  })

  it("shows skip button", () => {
    const wrapper = mountStep()

    expect(wrapper.find(".skip-button").exists()).toBe(true)
    expect(wrapper.find(".skip-button").text()).toBe("Skip")
  })

  it("emits skip on skip button click", async () => {
    const wrapper = mountStep()

    await wrapper.find(".skip-button").trigger("click")

    expect(wrapper.emitted("skip")).toBeTruthy()
  })

  it("disables skip button while installing", () => {
    const wrapper = mountStep({ installing: true })

    expect(wrapper.find(".skip-button").attributes("disabled")).toBeDefined()
  })
})
