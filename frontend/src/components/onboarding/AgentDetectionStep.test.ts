import { describe, it, expect } from "vitest"
import { mount } from "@vue/test-utils"
import AgentDetectionStep from "./AgentDetectionStep.vue"
import type { AgentStatus } from "../../types/skill"

describe("AgentDetectionStep", () => {
  const mockAgents: AgentStatus[] = [
    {
      id: "claude-code",
      displayName: "Claude Code",
      installed: true,
      skillCount: 5,
      globalSkillsDir: "",
    },
    {
      id: "cursor",
      displayName: "Cursor",
      installed: false,
      skillCount: 0,
      globalSkillsDir: "",
    },
  ]

  function mountStep(
    props: Partial<{
      agents: AgentStatus[]
      loading: boolean
      hasInstalledAgents: boolean
    }> = {},
  ) {
    return mount(AgentDetectionStep, {
      props: {
        agents: mockAgents,
        loading: false,
        hasInstalledAgents: true,
        ...props,
      },
    })
  }

  it("renders title", () => {
    const wrapper = mountStep()

    expect(wrapper.find("h1").text()).toBe("Detecting Coding Agents")
  })

  it("shows loading state", () => {
    const wrapper = mountStep({ loading: true, agents: [] })

    expect(wrapper.find(".loading").exists()).toBe(true)
    expect(wrapper.text()).toContain("Scanning for installed agents")
  })

  it("shows found agents count", () => {
    const wrapper = mountStep()

    expect(wrapper.find(".found").text()).toContain("Found 1 coding agent")
  })

  it("renders agent list", () => {
    const wrapper = mountStep()

    const items = wrapper.findAll(".agent-item")
    expect(items).toHaveLength(2)
  })

  it("marks installed agents", () => {
    const wrapper = mountStep()

    const installed = wrapper.findAll(".agent-item.installed")
    expect(installed).toHaveLength(1)
  })

  it("shows no agents detected message", () => {
    const wrapper = mountStep({ hasInstalledAgents: false })

    expect(wrapper.find(".not-found").text()).toContain(
      "No coding agents detected",
    )
  })

  it("shows blocked message when no agents", () => {
    const wrapper = mountStep({ hasInstalledAgents: false })

    expect(wrapper.find(".blocked-message").exists()).toBe(true)
  })

  it("shows continue button when agents found", () => {
    const wrapper = mountStep()

    expect(wrapper.find(".primary-button").text()).toBe("Continue")
  })

  it("emits next on continue click", async () => {
    const wrapper = mountStep()

    await wrapper.find(".primary-button").trigger("click")

    expect(wrapper.emitted("next")).toBeTruthy()
  })

  it("uses plural for multiple agents", () => {
    const agents = [
      {
        id: "a",
        displayName: "A",
        installed: true,
        skillCount: 0,
        globalSkillsDir: "",
      },
      {
        id: "b",
        displayName: "B",
        installed: true,
        skillCount: 0,
        globalSkillsDir: "",
      },
    ]
    const wrapper = mountStep({ agents, hasInstalledAgents: true })

    expect(wrapper.find(".found").text()).toContain("2 coding agents")
  })
})
