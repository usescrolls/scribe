import { describe, it, expect, vi, beforeEach } from "vitest"
import { flushPromises, mount } from "@vue/test-utils"
import { defineComponent } from "vue"
import { useAgents } from "./useAgents"
import { mockAppService, mockEvents } from "../test/setup"
import type { AgentStatus } from "../types/skill"

describe("useAgents", () => {
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
    {
      id: "cline",
      displayName: "Cline",
      installed: false,
      skillCount: 0,
      globalSkillsDir: "~/.cline/skills",
    },
  ]

  beforeEach(() => {
    vi.clearAllMocks()
    mockAppService.GetAgentStatus.mockResolvedValue(mockAgents)
  })

  describe("fetchAgents", () => {
    it("fetches agent status", async () => {
      const { agents, fetchAgents } = useAgents()

      await fetchAgents()
      await flushPromises()

      expect(mockAppService.GetAgentStatus).toHaveBeenCalled()
      expect(agents.value).toEqual(mockAgents)
    })

    it("sets loading state during fetch", async () => {
      const { loading, fetchAgents } = useAgents()

      const fetchPromise = fetchAgents()
      expect(loading.value).toBe(true)

      await fetchPromise
      expect(loading.value).toBe(false)
    })

    it("handles fetch errors", async () => {
      mockAppService.GetAgentStatus.mockRejectedValue(
        new Error("Network error"),
      )

      const { error, fetchAgents } = useAgents()
      await fetchAgents()

      expect(error.value).toBe("Network error")
    })
  })

  describe("computed properties", () => {
    it("computes installedAgents correctly", async () => {
      const { installedAgents, fetchAgents } = useAgents()

      await fetchAgents()
      await flushPromises()

      expect(installedAgents.value).toHaveLength(2)
      expect(installedAgents.value.map((a) => a.id)).toEqual([
        "claude-code",
        "cursor",
      ])
    })

    it("computes installedCount correctly", async () => {
      const { installedCount, fetchAgents } = useAgents()

      await fetchAgents()
      await flushPromises()

      expect(installedCount.value).toBe(2)
    })

    it("computes totalCount correctly", async () => {
      const { totalCount, fetchAgents } = useAgents()

      await fetchAgents()
      await flushPromises()

      expect(totalCount.value).toBe(4)
    })
  })

  describe("selectAgent", () => {
    it("selects an agent", async () => {
      const { selectedAgent, selectAgent, fetchAgents } = useAgents()

      await fetchAgents()
      selectAgent("claude-code")

      expect(selectedAgent.value).toBe("claude-code")
    })

    it("clears selection with null", async () => {
      const { selectedAgent, selectAgent, fetchAgents } = useAgents()

      await fetchAgents()
      selectAgent("claude-code")
      selectAgent(null)

      expect(selectedAgent.value).toBe(null)
    })
  })

  describe("initial state", () => {
    it("starts with empty agents array", () => {
      const { agents } = useAgents()
      expect(agents.value).toEqual([])
    })

    it("starts with loading true", () => {
      const { loading } = useAgents()
      expect(loading.value).toBe(true)
    })

    it("starts with no selection", () => {
      const { selectedAgent } = useAgents()
      expect(selectedAgent.value).toBe(null)
    })
  })

  describe("event subscriptions", () => {
    it("subscribes to skills-updated and workspace-changed events on mount", async () => {
      const TestComponent = defineComponent({
        setup() {
          useAgents()
          return () => null
        },
      })

      mount(TestComponent)
      await flushPromises()

      // Should subscribe to both events
      expect(mockEvents.On).toHaveBeenCalledWith(
        "skills-updated",
        expect.any(Function),
      )
      expect(mockEvents.On).toHaveBeenCalledWith(
        "workspace-changed",
        expect.any(Function),
      )
    })

    it("unsubscribes from events on unmount", async () => {
      const mockUnsubscribeSkills = vi.fn()
      const mockUnsubscribeWorkspace = vi.fn()

      // Return different unsubscribe functions for each call
      mockEvents.On.mockReturnValueOnce(
        mockUnsubscribeSkills,
      ).mockReturnValueOnce(mockUnsubscribeWorkspace)

      const TestComponent = defineComponent({
        setup() {
          useAgents()
          return () => null
        },
      })

      const wrapper = mount(TestComponent)
      await flushPromises()

      // Unmount the component
      wrapper.unmount()

      // Both unsubscribe functions should be called
      expect(mockUnsubscribeSkills).toHaveBeenCalled()
      expect(mockUnsubscribeWorkspace).toHaveBeenCalled()
    })

    it("refreshes agents when workspace-changed event fires", async () => {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      let workspaceChangedCallback: any = null

      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      ;(mockEvents.On as any).mockImplementation(
        (event: string, callback: () => void) => {
          if (event === "workspace-changed") {
            workspaceChangedCallback = callback
          }
          return vi.fn()
        },
      )

      const TestComponent = defineComponent({
        setup() {
          useAgents()
          return () => null
        },
      })

      mount(TestComponent)
      await flushPromises()

      // Clear mock calls from initial mount
      mockAppService.GetAgentStatus.mockClear()

      // Simulate workspace-changed event
      if (workspaceChangedCallback) {
        workspaceChangedCallback()
        await flushPromises()
      }

      // Should have fetched agents again
      expect(mockAppService.GetAgentStatus).toHaveBeenCalled()
    })
  })
})
