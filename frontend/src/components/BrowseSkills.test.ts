import { describe, it, expect, vi, beforeEach } from "vitest"
import { mount, flushPromises } from "@vue/test-utils"
import { Dialogs } from "@wailsio/runtime"
import BrowseSkills from "./BrowseSkills.vue"
import { mockAppService, mockBrowser } from "../test/setup"
import type { SkillInfo, WorkspaceInfo } from "../types/skill"

describe("BrowseSkills", () => {
  const mockSkills: SkillInfo[] = [
    {
      name: "react-patterns",
      description: "React best practices",
      source: "vercel-labs/skills",
      sourceType: "github",
      installedAt: "2025-01-29T10:00:00Z",
      agents: ["claude-code"],
    },
    {
      name: "typescript-tips",
      description: "TypeScript tips",
      source: "vercel-labs/skills",
      sourceType: "github",
      installedAt: "2025-01-28T10:00:00Z",
      agents: ["claude-code"],
    },
    {
      name: "vue-utils",
      description: "Vue utilities",
      source: "vue/skills",
      sourceType: "github",
      installedAt: "2025-01-27T10:00:00Z",
      agents: ["cursor"],
    },
  ]

  const mockWorkspaces: WorkspaceInfo[] = [
    {
      name: "my-workspace",
      description: "My workspace",
      skills: ["react-patterns"],
      isActive: true,
    },
    {
      name: "other",
      description: "",
      skills: [],
      isActive: false,
    },
  ]

  beforeEach(() => {
    vi.clearAllMocks()
    mockAppService.GetSkills.mockResolvedValue(mockSkills)
    mockAppService.GetWorkspaces.mockResolvedValue(mockWorkspaces)
  })

  async function mountBrowseSkills() {
    const wrapper = mount(BrowseSkills)
    await flushPromises()
    return wrapper
  }

  describe("rendering", () => {
    it("shows loading state initially", () => {
      mockAppService.GetSkills.mockReturnValue(new Promise(() => {}))
      mockAppService.GetWorkspaces.mockReturnValue(new Promise(() => {}))
      const wrapper = mount(BrowseSkills)

      expect(wrapper.find(".loading").exists()).toBe(true)
    })

    it("shows error on fetch failure", async () => {
      mockAppService.GetSkills.mockRejectedValue(new Error("Failed"))

      const wrapper = await mountBrowseSkills()

      expect(wrapper.find(".error").exists()).toBe(true)
    })

    it("shows empty state when no skills", async () => {
      mockAppService.GetSkills.mockResolvedValue([])

      const wrapper = await mountBrowseSkills()

      expect(wrapper.find(".empty").exists()).toBe(true)
      expect(wrapper.text()).toContain("No skills installed yet")
    })

    it("shows total skills count", async () => {
      const wrapper = await mountBrowseSkills()

      expect(wrapper.find(".count").text()).toBe("3 skills installed")
    })

    it("groups skills by source", async () => {
      const wrapper = await mountBrowseSkills()

      const groups = wrapper.findAll(".source-group")
      expect(groups).toHaveLength(2)
    })

    it("shows source badge and name in group header", async () => {
      const wrapper = await mountBrowseSkills()

      const badges = wrapper.findAll(".group-badge")
      expect(badges.length).toBeGreaterThanOrEqual(2)
      expect(badges[0].text()).toBe("github")

      const sources = wrapper.findAll(".group-source")
      expect(sources[0].text()).toBe("vercel-labs/skills")
      expect(sources[1].text()).toBe("vue/skills")
    })

    it("renders source as link when sourceUrl is present", async () => {
      mockAppService.GetSkills.mockResolvedValue([
        {
          ...mockSkills[0],
          sourceUrl: "https://github.com/vercel-labs/skills",
        },
        {
          ...mockSkills[1],
          sourceUrl: "https://github.com/vercel-labs/skills",
        },
        mockSkills[2],
      ])

      const wrapper = await mountBrowseSkills()

      const links = wrapper.findAll(".group-source-link")
      expect(links).toHaveLength(1)
      expect(links[0].text()).toBe("vercel-labs/skills")

      // Group without sourceUrl renders as plain span
      const plainSources = wrapper.findAll(
        ".group-source:not(.group-source-link)",
      )
      expect(plainSources).toHaveLength(1)
      expect(plainSources[0].text()).toBe("vue/skills")
    })

    it("opens source URL in browser on click", async () => {
      mockAppService.GetSkills.mockResolvedValue([
        {
          ...mockSkills[0],
          sourceUrl: "https://github.com/vercel-labs/skills",
        },
      ])

      const wrapper = await mountBrowseSkills()

      const link = wrapper.find(".group-source-link")
      await link.trigger("click")

      expect(mockBrowser.OpenURL).toHaveBeenCalledWith(
        "https://github.com/vercel-labs/skills",
      )
    })

    it("shows skill count per group", async () => {
      const wrapper = await mountBrowseSkills()

      const counts = wrapper.findAll(".group-count")
      expect(counts[0].text()).toBe("2")
      expect(counts[1].text()).toBe("1")
    })

    it("renders all skills across groups", async () => {
      const wrapper = await mountBrowseSkills()

      const cards = wrapper.findAllComponents({ name: "SkillCard" })
      expect(cards).toHaveLength(3)
    })

    it("renders uninstall button on all skills", async () => {
      const wrapper = await mountBrowseSkills()

      const cards = wrapper.findAllComponents({ name: "SkillCard" })
      cards.forEach((card) => {
        expect(card.props("showUninstall")).toBe(true)
      })
    })

    it("renders workspace picker on all skills", async () => {
      const wrapper = await mountBrowseSkills()

      const cards = wrapper.findAllComponents({ name: "SkillCard" })
      cards.forEach((card) => {
        expect(card.props("showWorkspacePicker")).toBe(true)
      })
    })

    it("passes workspace membership to skill cards", async () => {
      const wrapper = await mountBrowseSkills()

      const cards = wrapper.findAllComponents({ name: "SkillCard" })
      // react-patterns is in my-workspace
      const reactCard = cards.find(
        (c) => c.props("skill").name === "react-patterns",
      )
      expect(reactCard?.props("skillWorkspaces")).toEqual(["my-workspace"])

      // vue-utils is not in any workspace
      const vueCard = cards.find((c) => c.props("skill").name === "vue-utils")
      expect(vueCard?.props("skillWorkspaces")).toEqual([])
    })
  })

  describe("add to workspace", () => {
    it("calls AddSkillToWorkspace on add-to-workspace event", async () => {
      mockAppService.AddSkillToWorkspace.mockResolvedValue(undefined)

      const wrapper = await mountBrowseSkills()
      const card = wrapper.findComponent({ name: "SkillCard" })
      await card.vm.$emit("add-to-workspace", "typescript-tips", "my-workspace")
      await flushPromises()

      expect(mockAppService.AddSkillToWorkspace).toHaveBeenCalledWith(
        "typescript-tips",
        "my-workspace",
      )
    })

    it("shows error on add failure", async () => {
      mockAppService.AddSkillToWorkspace.mockRejectedValue(
        new Error("Add failed"),
      )

      const wrapper = await mountBrowseSkills()
      const card = wrapper.findComponent({ name: "SkillCard" })
      await card.vm.$emit("add-to-workspace", "typescript-tips", "my-workspace")
      await flushPromises()

      expect(wrapper.text()).toContain("Add failed")
    })
  })

  describe("remove from workspace", () => {
    it("calls RemoveSkillFromWorkspace on remove-from-workspace event", async () => {
      mockAppService.RemoveSkillFromWorkspace.mockResolvedValue(undefined)

      const wrapper = await mountBrowseSkills()
      const card = wrapper.findComponent({ name: "SkillCard" })
      await card.vm.$emit(
        "remove-from-workspace",
        "react-patterns",
        "my-workspace",
      )
      await flushPromises()

      expect(mockAppService.RemoveSkillFromWorkspace).toHaveBeenCalledWith(
        "react-patterns",
        "my-workspace",
      )
    })
  })

  describe("uninstall", () => {
    it("calls RemoveSkill on confirm", async () => {
      vi.mocked(Dialogs.Question).mockResolvedValue("Uninstall")
      mockAppService.RemoveSkill.mockResolvedValue(undefined)

      const wrapper = await mountBrowseSkills()
      const card = wrapper.findComponent({ name: "SkillCard" })
      await card.vm.$emit("uninstall", "typescript-tips")
      await flushPromises()

      expect(Dialogs.Question).toHaveBeenCalledWith(
        expect.objectContaining({ Title: "Uninstall Skill" }),
      )
      expect(mockAppService.RemoveSkill).toHaveBeenCalledWith("typescript-tips")
    })

    it("does not uninstall on cancel", async () => {
      vi.mocked(Dialogs.Question).mockResolvedValue("Cancel")

      const wrapper = await mountBrowseSkills()
      const card = wrapper.findComponent({ name: "SkillCard" })
      await card.vm.$emit("uninstall", "typescript-tips")
      await flushPromises()

      expect(mockAppService.RemoveSkill).not.toHaveBeenCalled()
    })

    it("shows error on uninstall failure", async () => {
      vi.mocked(Dialogs.Question).mockResolvedValue("Uninstall")
      mockAppService.RemoveSkill.mockRejectedValue(
        new Error("Uninstall failed"),
      )

      const wrapper = await mountBrowseSkills()
      const card = wrapper.findComponent({ name: "SkillCard" })
      await card.vm.$emit("uninstall", "react-patterns")
      await flushPromises()

      expect(wrapper.text()).toContain("Uninstall failed")
    })
  })
})
