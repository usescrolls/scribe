import { describe, it, expect, vi, beforeEach } from "vitest"
import { mount, flushPromises } from "@vue/test-utils"
import AgentStatusPanel from "./AgentStatusPanel.vue"
import { mockAppService } from "../test/setup"
import type { AgentStatus } from "../types/skill"

describe("AgentStatusPanel", () => {
  const mockAgents: AgentStatus[] = [
    {
      id: "claude-code",
      displayName: "Claude Code",
      installed: true,
      skillCount: 5,
      globalSkillsDir: "~/.claude/skills",
    },
    {
      id: "cursor",
      displayName: "Cursor",
      installed: true,
      skillCount: 3,
      globalSkillsDir: "~/.cursor/skills",
    },
    {
      id: "github-copilot",
      displayName: "GitHub Copilot",
      installed: false,
      skillCount: 0,
      globalSkillsDir: "~/.copilot/skills",
    },
  ]

  beforeEach(() => {
    vi.clearAllMocks()
    mockAppService.GetAgentStatus.mockResolvedValue(mockAgents)
  })

  async function mountAgentStatusPanel() {
    const wrapper = mount(AgentStatusPanel)
    await flushPromises()
    return wrapper
  }

  describe("rendering", () => {
    it("renders panel header", async () => {
      const wrapper = await mountAgentStatusPanel()

      expect(wrapper.find(".panel-header h3").text()).toBe("Agents")
    })

    it("shows installed count in summary", async () => {
      const wrapper = await mountAgentStatusPanel()

      expect(wrapper.find(".summary").text()).toBe("2/3 installed")
    })

    it("renders all agents", async () => {
      const wrapper = await mountAgentStatusPanel()

      const agentItems = wrapper.findAll(".agent-item")
      expect(agentItems).toHaveLength(3)
    })

    it("shows checkmark for installed agents", async () => {
      const wrapper = await mountAgentStatusPanel()

      const checkmarks = wrapper.findAll(".checkmark")
      expect(checkmarks).toHaveLength(2)
    })

    it("shows empty circle for non-installed agents", async () => {
      const wrapper = await mountAgentStatusPanel()

      const emptyCircles = wrapper.findAll(".empty")
      expect(emptyCircles).toHaveLength(1)
    })

    it("shows skill count for installed agents", async () => {
      const wrapper = await mountAgentStatusPanel()

      const skillCounts = wrapper.findAll(".skill-count")
      expect(skillCounts[0].text()).toBe("5 skills")
      expect(skillCounts[1].text()).toBe("3 skills")
    })

    it('shows "Not installed" for non-installed agents', async () => {
      const wrapper = await mountAgentStatusPanel()

      const notInstalled = wrapper.findAll(".not-installed")
      expect(notInstalled).toHaveLength(1)
      expect(notInstalled[0].text()).toBe("Not installed")
    })

    it("sorts installed agents first", async () => {
      const wrapper = await mountAgentStatusPanel()

      const agentNames = wrapper.findAll(".agent-name")
      // Installed agents should come first (Claude Code, Cursor), then alphabetically
      expect(agentNames[0].text()).toBe("Claude Code")
      expect(agentNames[1].text()).toBe("Cursor")
      expect(agentNames[2].text()).toBe("GitHub Copilot")
    })
  })

  describe("collapsing", () => {
    it("toggles collapse on header click", async () => {
      const wrapper = await mountAgentStatusPanel()

      // Initially expanded
      expect(wrapper.find(".agent-list").exists()).toBe(true)

      // Click to collapse
      await wrapper.find(".panel-header").trigger("click")
      expect(wrapper.find(".agent-list").exists()).toBe(false)

      // Click to expand
      await wrapper.find(".panel-header").trigger("click")
      expect(wrapper.find(".agent-list").exists()).toBe(true)
    })
  })

  describe("singular/plural handling", () => {
    it('uses singular "skill" for count of 1', async () => {
      mockAppService.GetAgentStatus.mockResolvedValue([
        {
          id: "test-agent",
          displayName: "Test Agent",
          installed: true,
          skillCount: 1,
          globalSkillsDir: "~/.test/skills",
        },
      ])

      const wrapper = await mountAgentStatusPanel()

      expect(wrapper.find(".skill-count").text()).toBe("1 skill")
    })
  })
})
