import { describe, it, expect, vi, beforeEach, afterEach } from "vitest"
import { mount, flushPromises } from "@vue/test-utils"
import BrowseSkills from "./BrowseSkills.vue"
import ConfirmDialog from "./ConfirmDialog.vue"
import SkillDetailModal from "./SkillDetailModal.vue"
import { mockAppService, mockBrowser, mockEvents } from "../test/setup"
import { useSkillUpdateChecker } from "../composables/useSkillUpdateChecker"
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

    it("shows source avatar in group header for github sources", async () => {
      const wrapper = await mountBrowseSkills()

      const avatars = wrapper.findAllComponents({ name: "SourceAvatar" })
      expect(avatars).toHaveLength(2)
      expect(avatars[0].props("source")).toBe("vercel-labs/skills")
      expect(avatars[0].props("sourceType")).toBe("github")
    })

    it("passes isPrivate to source avatar for private groups", async () => {
      mockAppService.GetSkills.mockResolvedValue([
        { ...mockSkills[0], isPrivate: true },
        { ...mockSkills[1], isPrivate: true },
      ])

      const wrapper = await mountBrowseSkills()

      const avatar = wrapper.findComponent({ name: "SourceAvatar" })
      expect(avatar.props("isPrivate")).toBe(true)
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
    it("shows confirm dialog on uninstall", async () => {
      const wrapper = await mountBrowseSkills()
      const card = wrapper.findComponent({ name: "SkillCard" })
      await card.vm.$emit("uninstall", "typescript-tips")
      await flushPromises()

      const dialog = wrapper.findComponent(ConfirmDialog)
      expect(dialog.exists()).toBe(true)
      expect(dialog.props("title")).toBe("Uninstall Skill")
      expect(dialog.props("danger")).toBe(true)
    })

    it("calls RemoveSkill on confirm", async () => {
      mockAppService.RemoveSkill.mockResolvedValue(undefined)

      const wrapper = await mountBrowseSkills()
      const card = wrapper.findComponent({ name: "SkillCard" })
      await card.vm.$emit("uninstall", "typescript-tips")
      await flushPromises()

      const dialog = wrapper.findComponent(ConfirmDialog)
      await dialog.vm.$emit("confirm")
      await flushPromises()

      expect(mockAppService.RemoveSkill).toHaveBeenCalledWith("typescript-tips")
    })

    it("does not uninstall on cancel", async () => {
      const wrapper = await mountBrowseSkills()
      const card = wrapper.findComponent({ name: "SkillCard" })
      await card.vm.$emit("uninstall", "typescript-tips")
      await flushPromises()

      const dialog = wrapper.findComponent(ConfirmDialog)
      await dialog.vm.$emit("cancel")
      await flushPromises()

      expect(mockAppService.RemoveSkill).not.toHaveBeenCalled()
    })

    it("shows error on uninstall failure", async () => {
      mockAppService.RemoveSkill.mockRejectedValue(
        new Error("Uninstall failed"),
      )

      const wrapper = await mountBrowseSkills()
      const card = wrapper.findComponent({ name: "SkillCard" })
      await card.vm.$emit("uninstall", "react-patterns")
      await flushPromises()

      const dialog = wrapper.findComponent(ConfirmDialog)
      await dialog.vm.$emit("confirm")
      await flushPromises()

      expect(wrapper.text()).toContain("Uninstall failed")
    })
  })

  describe("uninstall all from source", () => {
    it("shows uninstall all button for groups with multiple skills", async () => {
      const wrapper = await mountBrowseSkills()

      const buttons = wrapper.findAll(".group-uninstall-btn")
      // vercel-labs/skills has 2 skills -> shows button
      // vue/skills has 1 skill -> no button
      expect(buttons).toHaveLength(1)
      expect(buttons[0].text()).toBe("Uninstall all")
    })

    it("does not show uninstall all button for single-skill groups", async () => {
      mockAppService.GetSkills.mockResolvedValue([
        mockSkills[2], // vue/skills group with 1 skill
      ])

      const wrapper = await mountBrowseSkills()

      expect(wrapper.find(".group-uninstall-btn").exists()).toBe(false)
    })

    it("shows confirm dialog when uninstall all is clicked", async () => {
      const wrapper = await mountBrowseSkills()

      await wrapper.find(".group-uninstall-btn").trigger("click")
      await flushPromises()

      const dialogs = wrapper.findAllComponents(ConfirmDialog)
      const groupDialog = dialogs.find(
        (d) => d.props("title") === "Uninstall All Skills",
      )
      expect(groupDialog).toBeDefined()
      expect(groupDialog!.props("message")).toContain("2 skills")
      expect(groupDialog!.props("message")).toContain("vercel-labs/skills")
      expect(groupDialog!.props("danger")).toBe(true)
    })

    it("calls RemoveSkill for each skill in group on confirm", async () => {
      mockAppService.RemoveSkill.mockResolvedValue(undefined)

      const wrapper = await mountBrowseSkills()

      await wrapper.find(".group-uninstall-btn").trigger("click")
      await flushPromises()

      const dialogs = wrapper.findAllComponents(ConfirmDialog)
      const groupDialog = dialogs.find(
        (d) => d.props("title") === "Uninstall All Skills",
      )
      await groupDialog!.vm.$emit("confirm")
      await flushPromises()

      expect(mockAppService.RemoveSkill).toHaveBeenCalledWith("react-patterns")
      expect(mockAppService.RemoveSkill).toHaveBeenCalledWith("typescript-tips")
      expect(mockAppService.RemoveSkill).toHaveBeenCalledTimes(2)
    })

    it("does not uninstall on cancel", async () => {
      const wrapper = await mountBrowseSkills()

      await wrapper.find(".group-uninstall-btn").trigger("click")
      await flushPromises()

      const dialogs = wrapper.findAllComponents(ConfirmDialog)
      const groupDialog = dialogs.find(
        (d) => d.props("title") === "Uninstall All Skills",
      )
      await groupDialog!.vm.$emit("cancel")
      await flushPromises()

      expect(mockAppService.RemoveSkill).not.toHaveBeenCalled()
    })

    it("shows error on bulk uninstall failure", async () => {
      mockAppService.RemoveSkill.mockRejectedValue(
        new Error("Bulk uninstall failed"),
      )

      const wrapper = await mountBrowseSkills()

      await wrapper.find(".group-uninstall-btn").trigger("click")
      await flushPromises()

      const dialogs = wrapper.findAllComponents(ConfirmDialog)
      const groupDialog = dialogs.find(
        (d) => d.props("title") === "Uninstall All Skills",
      )
      await groupDialog!.vm.$emit("confirm")
      await flushPromises()

      expect(wrapper.text()).toContain("Bulk uninstall failed")
    })
  })

  describe("update group", () => {
    it("shows update button for github source groups", async () => {
      const wrapper = await mountBrowseSkills()

      const buttons = wrapper.findAll(".group-update-btn")
      // Both groups are github -> both get update button
      expect(buttons).toHaveLength(2)
      expect(buttons[0].text()).toBe("Update")
    })

    it("does not show update button for local source groups", async () => {
      mockAppService.GetSkills.mockResolvedValue([
        {
          name: "local-skill",
          description: "A local skill",
          source: "/path/to/skill",
          sourceType: "local",
          installedAt: "2025-01-29T10:00:00Z",
          agents: ["claude-code"],
        },
      ])

      const wrapper = await mountBrowseSkills()

      expect(wrapper.find(".group-update-btn").exists()).toBe(false)
    })

    it("does not show update button for builtin source groups", async () => {
      mockAppService.GetSkills.mockResolvedValue([
        {
          name: "builtin-skill",
          description: "A builtin skill",
          source: "builtin",
          sourceType: "builtin",
          installedAt: "2025-01-29T10:00:00Z",
          agents: ["claude-code"],
        },
      ])

      const wrapper = await mountBrowseSkills()

      expect(wrapper.find(".group-update-btn").exists()).toBe(false)
    })

    it("calls UpdateSkill for each skill in the group", async () => {
      mockAppService.UpdateSkill.mockResolvedValue(undefined)

      const wrapper = await mountBrowseSkills()

      const updateButtons = wrapper.findAll(".group-update-btn")
      // Click update on the first group (vercel-labs/skills with 2 skills)
      await updateButtons[0].trigger("click")
      await flushPromises()

      expect(mockAppService.UpdateSkill).toHaveBeenCalledWith("react-patterns")
      expect(mockAppService.UpdateSkill).toHaveBeenCalledWith("typescript-tips")
      expect(mockAppService.UpdateSkill).toHaveBeenCalledTimes(2)
    })

    it("shows 'Updating...' text while updating", async () => {
      // eslint-disable-next-line @typescript-eslint/no-empty-function
      let resolveUpdate: () => void = () => {}
      mockAppService.UpdateSkill.mockImplementation(
        () =>
          new Promise<void>((r) => {
            resolveUpdate = r
          }),
      )

      const wrapper = await mountBrowseSkills()

      const updateButtons = wrapper.findAll(".group-update-btn")
      await updateButtons[0].trigger("click")
      await wrapper.vm.$nextTick()

      // Button should show "Updating..." and be disabled
      const btn = wrapper.findAll(".group-update-btn")[0]
      expect(btn.text()).toBe("Updating...")
      expect(btn.attributes("disabled")).toBeDefined()

      // Resolve to clean up
      resolveUpdate()
      await flushPromises()
    })

    it("refreshes skills after update completes", async () => {
      mockAppService.UpdateSkill.mockResolvedValue(undefined)

      const wrapper = await mountBrowseSkills()

      // Clear to track calls after update
      mockAppService.GetSkills.mockClear()
      mockAppService.GetWorkspaces.mockClear()
      mockAppService.GetSkills.mockResolvedValue(mockSkills)
      mockAppService.GetWorkspaces.mockResolvedValue(mockWorkspaces)

      const updateButtons = wrapper.findAll(".group-update-btn")
      await updateButtons[0].trigger("click")
      await flushPromises()

      expect(mockAppService.GetSkills).toHaveBeenCalled()
      expect(mockAppService.GetWorkspaces).toHaveBeenCalled()
    })

    it("shows error on update failure for all skills in group", async () => {
      mockAppService.UpdateSkill.mockRejectedValue(new Error("Update failed"))

      const wrapper = await mountBrowseSkills()

      const updateButtons = wrapper.findAll(".group-update-btn")
      await updateButtons[0].trigger("click")
      await flushPromises()

      const toast = wrapper.findComponent({ name: "ToastNotification" })
      expect(toast.exists()).toBe(true)
      expect(toast.props("message")).toContain("Update failed")
      expect(toast.props("type")).toBe("error")
    })

    it("shows success toast when skills are updated", async () => {
      mockAppService.UpdateSkill.mockResolvedValue({
        skillName: "react-patterns",
        updated: true,
        oldHash: "abc1234",
        newHash: "def5678",
      })

      const wrapper = await mountBrowseSkills()

      const updateButtons = wrapper.findAll(".group-update-btn")
      await updateButtons[0].trigger("click")
      await flushPromises()

      const toast = wrapper.findComponent({ name: "ToastNotification" })
      expect(toast.exists()).toBe(true)
      expect(toast.props("message")).toContain("Updated")
      expect(toast.props("message")).toContain("abc1234")
      expect(toast.props("message")).toContain("def5678")
      expect(toast.props("type")).toBe("success")
    })

    it("shows info toast when all skills already up to date", async () => {
      mockAppService.UpdateSkill.mockResolvedValue({
        skillName: "react-patterns",
        updated: false,
      })

      const wrapper = await mountBrowseSkills()

      const updateButtons = wrapper.findAll(".group-update-btn")
      await updateButtons[0].trigger("click")
      await flushPromises()

      const toast = wrapper.findComponent({ name: "ToastNotification" })
      expect(toast.exists()).toBe(true)
      expect(toast.props("message")).toContain("already up to date")
      expect(toast.props("type")).toBe("info")
    })

    it("shows error toast on update failure", async () => {
      mockAppService.UpdateSkill.mockRejectedValue(new Error("Network error"))

      const wrapper = await mountBrowseSkills()

      const updateButtons = wrapper.findAll(".group-update-btn")
      await updateButtons[0].trigger("click")
      await flushPromises()

      const toast = wrapper.findComponent({ name: "ToastNotification" })
      expect(toast.exists()).toBe(true)
      expect(toast.props("message")).toContain("Network error")
      expect(toast.props("type")).toBe("error")
    })

    it("continues updating other skills when one fails", async () => {
      mockAppService.UpdateSkill.mockRejectedValueOnce(
        new Error("Skill removed"),
      ).mockResolvedValueOnce({ skillName: "typescript-tips", updated: false })

      const wrapper = await mountBrowseSkills()

      const updateButtons = wrapper.findAll(".group-update-btn")
      await updateButtons[0].trigger("click")
      await flushPromises()

      // Both skills should have been attempted
      expect(mockAppService.UpdateSkill).toHaveBeenCalledTimes(2)
      // Toast shows partial failure
      const toast = wrapper.findComponent({ name: "ToastNotification" })
      expect(toast.exists()).toBe(true)
      expect(toast.props("message")).toContain("1 failed")
      expect(toast.props("type")).toBe("error")
    })

    it("shows removed skills in toast when skill no longer in source", async () => {
      mockAppService.UpdateSkill.mockResolvedValue({
        skillName: "react-patterns",
        updated: false,
        removed: true,
      })

      const wrapper = await mountBrowseSkills()

      const updateButtons = wrapper.findAll(".group-update-btn")
      await updateButtons[0].trigger("click")
      await flushPromises()

      const toast = wrapper.findComponent({ name: "ToastNotification" })
      expect(toast.exists()).toBe(true)
      expect(toast.props("message")).toContain("Removed")
      expect(toast.props("message")).toContain("no longer in source")
      expect(toast.props("type")).toBe("info")
    })

    it("shows combined toast for updates and removals", async () => {
      mockAppService.UpdateSkill.mockResolvedValueOnce({
        skillName: "react-patterns",
        updated: true,
        oldHash: "aaa",
        newHash: "bbb",
      }).mockResolvedValueOnce({
        skillName: "typescript-tips",
        updated: false,
        removed: true,
      })

      const wrapper = await mountBrowseSkills()

      const updateButtons = wrapper.findAll(".group-update-btn")
      await updateButtons[0].trigger("click")
      await flushPromises()

      const toast = wrapper.findComponent({ name: "ToastNotification" })
      expect(toast.exists()).toBe(true)
      expect(toast.props("message")).toContain("Updated 1 skill")
      expect(toast.props("message")).toContain("Removed 1 skill")
      expect(toast.props("type")).toBe("info")
    })
  })

  describe("version display", () => {
    it("shows commit hash with relative time in group header", async () => {
      mockAppService.GetSkills.mockResolvedValue([
        {
          ...mockSkills[0],
          commitHash: "abc1234",
          commitDate: "2025-06-15T10:30:00Z",
        },
        {
          ...mockSkills[1],
          commitHash: "abc1234",
          commitDate: "2025-06-15T10:30:00Z",
        },
      ])

      const wrapper = await mountBrowseSkills()

      const version = wrapper.find(".group-version")
      expect(version.exists()).toBe(true)
      const text = version.text()
      expect(text).toContain("abc1234")
      expect(text).toContain("·")
      expect(text).toMatch(/ago|just now/)
    })

    it("does not show version when no commit hash", async () => {
      const wrapper = await mountBrowseSkills()

      // Default mockSkills have no commitHash
      const versions = wrapper.findAll(".group-version")
      expect(versions).toHaveLength(0)
    })

    it("shows tooltip with commit details", async () => {
      mockAppService.GetSkills.mockResolvedValue([
        {
          ...mockSkills[0],
          commitHash: "abc1234",
          commitDate: "2025-06-15T10:30:00Z",
          updatedAt: "2025-06-15T14:30:00Z",
        },
      ])

      const wrapper = await mountBrowseSkills()

      const version = wrapper.find(".group-version")
      expect(version.exists()).toBe(true)
      const title = version.attributes("title")
      expect(title).toContain("Commit: abc1234")
    })
  })

  describe("skill detail modal", () => {
    it("opens detail modal on card detail event", async () => {
      mockAppService.GetSkillContent.mockResolvedValue("# Content")

      const wrapper = await mountBrowseSkills()
      const card = wrapper.findComponent({ name: "SkillCard" })
      await card.vm.$emit("detail", mockSkills[0])
      await flushPromises()

      const modal = wrapper.findComponent(SkillDetailModal)
      expect(modal.exists()).toBe(true)
      expect(modal.props("skill")).toEqual(mockSkills[0])
    })

    it("closes detail modal on close event", async () => {
      mockAppService.GetSkillContent.mockResolvedValue("# Content")

      const wrapper = await mountBrowseSkills()
      const card = wrapper.findComponent({ name: "SkillCard" })
      await card.vm.$emit("detail", mockSkills[0])
      await flushPromises()

      const modal = wrapper.findComponent(SkillDetailModal)
      modal.vm.$emit("close")
      await flushPromises()

      expect(wrapper.findComponent(SkillDetailModal).exists()).toBe(false)
    })
  })

  describe("selection mode", () => {
    it("shows Select button in header", async () => {
      const wrapper = await mountBrowseSkills()

      const selectBtn = wrapper.find(".header-actions button")
      expect(selectBtn.exists()).toBe(true)
      expect(selectBtn.text()).toBe("Select")
    })

    it("enters selection mode on Select click", async () => {
      const wrapper = await mountBrowseSkills()

      await wrapper.find(".header-actions button").trigger("click")

      // Should show Select All and Cancel buttons
      const buttons = wrapper.findAll(".header-actions button")
      expect(buttons.some((b) => b.text() === "Select All")).toBe(true)
      expect(buttons.some((b) => b.text() === "Cancel")).toBe(true)
    })

    it("passes selectable prop to skill cards in selection mode", async () => {
      const wrapper = await mountBrowseSkills()

      await wrapper.find(".header-actions button").trigger("click")

      const cards = wrapper.findAllComponents({ name: "SkillCard" })
      cards.forEach((card) => {
        expect(card.props("selectable")).toBe(true)
      })
    })

    it("hides uninstall and workspace picker in selection mode", async () => {
      const wrapper = await mountBrowseSkills()

      await wrapper.find(".header-actions button").trigger("click")

      const cards = wrapper.findAllComponents({ name: "SkillCard" })
      cards.forEach((card) => {
        expect(card.props("showUninstall")).toBe(false)
        expect(card.props("showWorkspacePicker")).toBe(false)
      })
    })

    it("exits selection mode on Cancel click", async () => {
      const wrapper = await mountBrowseSkills()

      // Enter selection mode
      await wrapper.find(".header-actions button").trigger("click")

      // Click Cancel
      const cancelBtn = wrapper
        .findAll(".header-actions button")
        .find((b) => b.text() === "Cancel")
      await cancelBtn!.trigger("click")

      // Should show Select button again
      const selectBtn = wrapper.find(".header-actions button")
      expect(selectBtn.text()).toBe("Select")
    })

    it("toggles individual skill selection via toggle-select", async () => {
      const wrapper = await mountBrowseSkills()

      // Enter selection mode
      await wrapper.find(".header-actions button").trigger("click")

      // Toggle a skill
      const card = wrapper.findComponent({ name: "SkillCard" })
      await card.vm.$emit("toggle-select", "react-patterns")
      await wrapper.vm.$nextTick()

      // Card should now be selected
      const reactCard = wrapper
        .findAllComponents({ name: "SkillCard" })
        .find((c) => c.props("skill").name === "react-patterns")
      expect(reactCard?.props("selected")).toBe(true)
    })

    it("deselects a skill on second toggle", async () => {
      const wrapper = await mountBrowseSkills()

      // Enter selection mode
      await wrapper.find(".header-actions button").trigger("click")

      const card = wrapper.findComponent({ name: "SkillCard" })
      // Select
      await card.vm.$emit("toggle-select", "react-patterns")
      await wrapper.vm.$nextTick()
      // Deselect
      await card.vm.$emit("toggle-select", "react-patterns")
      await wrapper.vm.$nextTick()

      const reactCard = wrapper
        .findAllComponents({ name: "SkillCard" })
        .find((c) => c.props("skill").name === "react-patterns")
      expect(reactCard?.props("selected")).toBe(false)
    })

    it("Select All selects all skills", async () => {
      const wrapper = await mountBrowseSkills()

      // Enter selection mode
      await wrapper.find(".header-actions button").trigger("click")

      // Click Select All
      const selectAllBtn = wrapper
        .findAll(".header-actions button")
        .find((b) => b.text() === "Select All")
      await selectAllBtn!.trigger("click")

      const cards = wrapper.findAllComponents({ name: "SkillCard" })
      cards.forEach((card) => {
        expect(card.props("selected")).toBe(true)
      })
    })

    it("Deselect All deselects all skills after selecting all", async () => {
      const wrapper = await mountBrowseSkills()

      // Enter selection mode
      await wrapper.find(".header-actions button").trigger("click")

      // Click Select All
      const selectAllBtn = wrapper
        .findAll(".header-actions button")
        .find((b) => b.text() === "Select All")
      await selectAllBtn!.trigger("click")

      // Now button should say "Deselect All"
      const deselectAllBtn = wrapper
        .findAll(".header-actions button")
        .find((b) => b.text() === "Deselect All")
      expect(deselectAllBtn).toBeDefined()
      await deselectAllBtn!.trigger("click")

      const cards = wrapper.findAllComponents({ name: "SkillCard" })
      cards.forEach((card) => {
        expect(card.props("selected")).toBe(false)
      })
    })

    it("clears selection when exiting selection mode", async () => {
      const wrapper = await mountBrowseSkills()

      // Enter selection mode and select a skill
      await wrapper.find(".header-actions button").trigger("click")
      const card = wrapper.findComponent({ name: "SkillCard" })
      await card.vm.$emit("toggle-select", "react-patterns")
      await wrapper.vm.$nextTick()

      // Exit via Cancel
      const cancelBtn = wrapper
        .findAll(".header-actions button")
        .find((b) => b.text() === "Cancel")
      await cancelBtn!.trigger("click")

      // Re-enter selection mode - should have no selections
      await wrapper.find(".header-actions button").trigger("click")
      const cards = wrapper.findAllComponents({ name: "SkillCard" })
      cards.forEach((card) => {
        expect(card.props("selected")).toBe(false)
      })
    })
  })

  async function enterSelectionMode(wrapper: ReturnType<typeof mount>) {
    await wrapper.find(".header-actions button").trigger("click")
  }

  async function enterSelectionAndSelect(
    wrapper: ReturnType<typeof mount>,
    skillNames: string[],
  ) {
    await enterSelectionMode(wrapper)
    for (const name of skillNames) {
      const card = wrapper
        .findAllComponents({ name: "SkillCard" })
        .find((c) => c.props("skill").name === name)
      await card!.vm.$emit("toggle-select", name)
    }
    await wrapper.vm.$nextTick()
  }

  describe("bulk action bar", () => {
    it("does not show bulk action bar when no skills selected", async () => {
      const wrapper = await mountBrowseSkills()

      // Enter selection mode but don't select anything
      await wrapper.find(".header-actions button").trigger("click")

      expect(wrapper.find(".bulk-action-bar").exists()).toBe(false)
    })

    it("shows bulk action bar when skills are selected", async () => {
      const wrapper = await mountBrowseSkills()

      await enterSelectionAndSelect(wrapper, ["react-patterns"])

      expect(wrapper.find(".bulk-action-bar").exists()).toBe(true)
    })

    it("shows correct selection count", async () => {
      const wrapper = await mountBrowseSkills()

      await enterSelectionAndSelect(wrapper, [
        "react-patterns",
        "typescript-tips",
      ])

      expect(wrapper.find(".bulk-count").text()).toBe("2 skills selected")
    })

    it("shows singular count for one skill", async () => {
      const wrapper = await mountBrowseSkills()

      await enterSelectionAndSelect(wrapper, ["react-patterns"])

      expect(wrapper.find(".bulk-count").text()).toBe("1 skill selected")
    })

    it("shows workspace dropdown with all workspaces", async () => {
      const wrapper = await mountBrowseSkills()

      await enterSelectionAndSelect(wrapper, ["react-patterns"])

      const select = wrapper.find(".bulk-workspace-select")
      expect(select.exists()).toBe(true)

      const options = select.findAll("option")
      // First option is the disabled placeholder
      expect(options).toHaveLength(3)
      expect(options[1].text()).toBe("my-workspace")
      expect(options[2].text()).toBe("other")
    })

    it("disables Add button when no workspace is chosen", async () => {
      const wrapper = await mountBrowseSkills()

      await enterSelectionAndSelect(wrapper, ["react-patterns"])

      const addBtn = wrapper.find(".bulk-action-bar .btn-primary")
      expect(addBtn.attributes("disabled")).toBeDefined()
    })

    it("calls AddSkillToWorkspace for each selected skill", async () => {
      mockAppService.AddSkillToWorkspace.mockResolvedValue(undefined)

      const wrapper = await mountBrowseSkills()

      await enterSelectionAndSelect(wrapper, [
        "react-patterns",
        "typescript-tips",
      ])

      // Select a workspace
      const select = wrapper.find(".bulk-workspace-select")
      await select.setValue("other")

      // Click Add to workspace
      const addBtn = wrapper.find(".bulk-action-bar .btn-primary")
      await addBtn.trigger("click")
      await flushPromises()

      expect(mockAppService.AddSkillToWorkspace).toHaveBeenCalledWith(
        "react-patterns",
        "other",
      )
      expect(mockAppService.AddSkillToWorkspace).toHaveBeenCalledWith(
        "typescript-tips",
        "other",
      )
      expect(mockAppService.AddSkillToWorkspace).toHaveBeenCalledTimes(2)
    })

    it("shows success toast after bulk add", async () => {
      mockAppService.AddSkillToWorkspace.mockResolvedValue(undefined)

      const wrapper = await mountBrowseSkills()

      await enterSelectionAndSelect(wrapper, [
        "react-patterns",
        "typescript-tips",
      ])

      const select = wrapper.find(".bulk-workspace-select")
      await select.setValue("other")

      const addBtn = wrapper.find(".bulk-action-bar .btn-primary")
      await addBtn.trigger("click")
      await flushPromises()

      const toast = wrapper.findComponent({ name: "ToastNotification" })
      expect(toast.exists()).toBe(true)
      expect(toast.props("message")).toContain("2 skills")
      expect(toast.props("message")).toContain("other")
      expect(toast.props("type")).toBe("success")
    })

    it("exits selection mode after successful bulk add", async () => {
      mockAppService.AddSkillToWorkspace.mockResolvedValue(undefined)

      const wrapper = await mountBrowseSkills()

      await enterSelectionAndSelect(wrapper, ["react-patterns"])

      const select = wrapper.find(".bulk-workspace-select")
      await select.setValue("other")

      const addBtn = wrapper.find(".bulk-action-bar .btn-primary")
      await addBtn.trigger("click")
      await flushPromises()

      // Should be back to browse mode
      const selectBtn = wrapper.find(".header-actions button")
      expect(selectBtn.text()).toBe("Select")
      expect(wrapper.find(".bulk-action-bar").exists()).toBe(false)
    })

    it("shows error toast on bulk add failure", async () => {
      mockAppService.AddSkillToWorkspace.mockRejectedValue(
        new Error("Workspace full"),
      )

      const wrapper = await mountBrowseSkills()

      await enterSelectionAndSelect(wrapper, ["react-patterns"])

      const select = wrapper.find(".bulk-workspace-select")
      await select.setValue("other")

      const addBtn = wrapper.find(".bulk-action-bar .btn-primary")
      await addBtn.trigger("click")
      await flushPromises()

      const toast = wrapper.findComponent({ name: "ToastNotification" })
      expect(toast.exists()).toBe(true)
      expect(toast.props("message")).toContain("Workspace full")
      expect(toast.props("type")).toBe("error")
    })

    it("refreshes data after bulk add", async () => {
      mockAppService.AddSkillToWorkspace.mockResolvedValue(undefined)

      const wrapper = await mountBrowseSkills()

      mockAppService.GetSkills.mockClear()
      mockAppService.GetWorkspaces.mockClear()
      mockAppService.GetSkills.mockResolvedValue(mockSkills)
      mockAppService.GetWorkspaces.mockResolvedValue(mockWorkspaces)

      await enterSelectionAndSelect(wrapper, ["react-patterns"])

      const select = wrapper.find(".bulk-workspace-select")
      await select.setValue("other")

      const addBtn = wrapper.find(".bulk-action-bar .btn-primary")
      await addBtn.trigger("click")
      await flushPromises()

      expect(mockAppService.GetSkills).toHaveBeenCalled()
      expect(mockAppService.GetWorkspaces).toHaveBeenCalled()
    })

    it("hides bulk action bar on Cancel", async () => {
      const wrapper = await mountBrowseSkills()

      await enterSelectionAndSelect(wrapper, ["react-patterns"])
      expect(wrapper.find(".bulk-action-bar").exists()).toBe(true)

      // Click Cancel in the bulk action bar
      const cancelBtn = wrapper
        .findAll(".bulk-action-bar button")
        .find((b) => b.text() === "Cancel")
      await cancelBtn!.trigger("click")

      expect(wrapper.find(".bulk-action-bar").exists()).toBe(false)
    })
  })

  describe("group selection", () => {
    it("shows group checkbox in selection mode", async () => {
      const wrapper = await mountBrowseSkills()

      await enterSelectionMode(wrapper)

      const checkboxes = wrapper.findAll(".group-checkbox")
      expect(checkboxes.length).toBeGreaterThanOrEqual(1)
    })

    it("does not show group checkbox outside selection mode", async () => {
      const wrapper = await mountBrowseSkills()

      expect(wrapper.find(".group-checkbox").exists()).toBe(false)
    })

    it("selects all skills in a group when group checkbox is clicked", async () => {
      const wrapper = await mountBrowseSkills()

      await enterSelectionMode(wrapper)

      // Click the first group checkbox (vercel-labs/skills with 2 skills)
      const checkbox = wrapper.find(".group-checkbox")
      await checkbox.trigger("click")

      // Both skills in the group should be selected
      const cards = wrapper.findAllComponents({ name: "SkillCard" })
      const vercelCards = cards.filter(
        (c) => c.props("skill").source === "vercel-labs/skills",
      )
      vercelCards.forEach((card) => {
        expect(card.props("selected")).toBe(true)
      })
    })

    it("deselects all skills in a group when all are selected", async () => {
      const wrapper = await mountBrowseSkills()

      await enterSelectionMode(wrapper)

      // Select all in group
      const checkbox = wrapper.find(".group-checkbox")
      await checkbox.trigger("click")

      // Deselect all in group
      await checkbox.trigger("click")

      const cards = wrapper.findAllComponents({ name: "SkillCard" })
      const vercelCards = cards.filter(
        (c) => c.props("skill").source === "vercel-labs/skills",
      )
      vercelCards.forEach((card) => {
        expect(card.props("selected")).toBe(false)
      })
    })

    it("does not select skills from other groups", async () => {
      const wrapper = await mountBrowseSkills()

      await enterSelectionMode(wrapper)

      // Click the first group checkbox
      const checkbox = wrapper.find(".group-checkbox")
      await checkbox.trigger("click")

      // vue/skills group should not be selected
      const cards = wrapper.findAllComponents({ name: "SkillCard" })
      const vueCard = cards.find(
        (c) => c.props("skill").source === "vue/skills",
      )
      expect(vueCard?.props("selected")).toBe(false)
    })

    it("does not show group checkbox for system-only groups", async () => {
      mockAppService.GetSkills.mockResolvedValue([
        {
          name: "system-skill",
          description: "System",
          source: "system",
          sourceType: "system",
          installedAt: "2025-01-29T10:00:00Z",
          agents: ["claude-code"],
          isSystem: true,
        },
      ])

      const wrapper = await mountBrowseSkills()

      await enterSelectionMode(wrapper)

      expect(wrapper.find(".group-checkbox").exists()).toBe(false)
    })
  })

  describe("bulk remove from workspace", () => {
    it("shows remove from workspace button in bulk action bar", async () => {
      const wrapper = await mountBrowseSkills()

      await enterSelectionAndSelect(wrapper, ["react-patterns"])

      const removeBtn = wrapper.find(".bulk-action-bar .btn-danger")
      expect(removeBtn.exists()).toBe(true)
      expect(removeBtn.text()).toBe("Remove from workspace")
    })

    it("disables remove button when no workspace is chosen", async () => {
      const wrapper = await mountBrowseSkills()

      await enterSelectionAndSelect(wrapper, ["react-patterns"])

      const removeBtn = wrapper.find(".bulk-action-bar .btn-danger")
      expect(removeBtn.attributes("disabled")).toBeDefined()
    })

    it("calls RemoveSkillFromWorkspace for each selected skill", async () => {
      mockAppService.RemoveSkillFromWorkspace.mockResolvedValue(undefined)

      const wrapper = await mountBrowseSkills()

      await enterSelectionAndSelect(wrapper, [
        "react-patterns",
        "typescript-tips",
      ])

      const select = wrapper.find(".bulk-workspace-select")
      await select.setValue("my-workspace")

      const removeBtn = wrapper.find(".bulk-action-bar .btn-danger")
      await removeBtn.trigger("click")
      await flushPromises()

      // react-patterns is in my-workspace, typescript-tips is not
      expect(mockAppService.RemoveSkillFromWorkspace).toHaveBeenCalledWith(
        "react-patterns",
        "my-workspace",
      )
      expect(mockAppService.RemoveSkillFromWorkspace).toHaveBeenCalledTimes(1)
    })

    it("shows success toast after bulk remove", async () => {
      mockAppService.RemoveSkillFromWorkspace.mockResolvedValue(undefined)

      const wrapper = await mountBrowseSkills()

      await enterSelectionAndSelect(wrapper, ["react-patterns"])

      const select = wrapper.find(".bulk-workspace-select")
      await select.setValue("my-workspace")

      const removeBtn = wrapper.find(".bulk-action-bar .btn-danger")
      await removeBtn.trigger("click")
      await flushPromises()

      const toast = wrapper.findComponent({ name: "ToastNotification" })
      expect(toast.exists()).toBe(true)
      expect(toast.props("message")).toContain("1 skill")
      expect(toast.props("message")).toContain("my-workspace")
      expect(toast.props("type")).toBe("success")
    })

    it("shows info toast when no selected skills are in the workspace", async () => {
      const wrapper = await mountBrowseSkills()

      await enterSelectionAndSelect(wrapper, ["vue-utils"])

      const select = wrapper.find(".bulk-workspace-select")
      await select.setValue("my-workspace")

      const removeBtn = wrapper.find(".bulk-action-bar .btn-danger")
      await removeBtn.trigger("click")
      await flushPromises()

      const toast = wrapper.findComponent({ name: "ToastNotification" })
      expect(toast.exists()).toBe(true)
      expect(toast.props("message")).toContain("No selected skills")
      expect(toast.props("type")).toBe("info")
    })

    it("exits selection mode after bulk remove", async () => {
      mockAppService.RemoveSkillFromWorkspace.mockResolvedValue(undefined)

      const wrapper = await mountBrowseSkills()

      await enterSelectionAndSelect(wrapper, ["react-patterns"])

      const select = wrapper.find(".bulk-workspace-select")
      await select.setValue("my-workspace")

      const removeBtn = wrapper.find(".bulk-action-bar .btn-danger")
      await removeBtn.trigger("click")
      await flushPromises()

      const selectBtn = wrapper.find(".header-actions button")
      expect(selectBtn.text()).toBe("Select")
    })

    it("shows error toast on bulk remove failure", async () => {
      mockAppService.RemoveSkillFromWorkspace.mockRejectedValue(
        new Error("Remove failed"),
      )

      const wrapper = await mountBrowseSkills()

      await enterSelectionAndSelect(wrapper, ["react-patterns"])

      const select = wrapper.find(".bulk-workspace-select")
      await select.setValue("my-workspace")

      const removeBtn = wrapper.find(".bulk-action-bar .btn-danger")
      await removeBtn.trigger("click")
      await flushPromises()

      const toast = wrapper.findComponent({ name: "ToastNotification" })
      expect(toast.exists()).toBe(true)
      expect(toast.props("message")).toContain("Remove failed")
      expect(toast.props("type")).toBe("error")
    })
  })

  describe("refetch without loading flicker", () => {
    it("does not show loading spinner when refetching after initial load", async () => {
      const wrapper = await mountBrowseSkills()

      // Should not be in loading state after initial load
      expect(wrapper.find(".loading").exists()).toBe(false)
      expect(wrapper.findAll(".source-group").length).toBeGreaterThan(0)

      // Trigger a refetch (e.g. after an update)
      mockAppService.GetSkills.mockResolvedValue(mockSkills)
      mockAppService.GetWorkspaces.mockResolvedValue(mockWorkspaces)

      const updateButtons = wrapper.findAll(".group-update-btn")
      mockAppService.UpdateSkill.mockResolvedValue({
        skillName: "react-patterns",
        updated: false,
      })
      await updateButtons[0].trigger("click")
      await flushPromises()

      // Loading spinner should never have appeared — skill list should remain visible
      expect(wrapper.find(".loading").exists()).toBe(false)
      expect(wrapper.findAll(".source-group").length).toBeGreaterThan(0)
    })

    it("shows loading spinner only on initial load", () => {
      mockAppService.GetSkills.mockReturnValue(new Promise(() => {}))
      mockAppService.GetWorkspaces.mockReturnValue(new Promise(() => {}))
      const wrapper = mount(BrowseSkills)

      expect(wrapper.find(".loading").exists()).toBe(true)
    })
  })

  describe("debounced event handling", () => {
    it("registers event listeners with debounced handler on mount", async () => {
      await mountBrowseSkills()

      // Events.On should have been called for both events
      expect(mockEvents.On).toHaveBeenCalledWith(
        "skills-updated",
        expect.any(Function),
      )
      expect(mockEvents.On).toHaveBeenCalledWith(
        "workspace-changed",
        expect.any(Function),
      )
    })

    it("coalesces rapid event emissions into a single fetch", async () => {
      vi.useFakeTimers()

      // Capture event handlers registered via Events.On
      const eventHandlers: Record<string, Function> = {}
      ;(mockEvents.On as ReturnType<typeof vi.fn>).mockImplementation(
        (event: string, handler: Function) => {
          eventHandlers[event] = handler
          return vi.fn()
        },
      )

      mockAppService.GetSkills.mockResolvedValue(mockSkills)
      mockAppService.GetWorkspaces.mockResolvedValue(mockWorkspaces)

      mount(BrowseSkills)
      // Flush the initial fetchAll
      await flushPromises()

      // Clear to track subsequent calls
      mockAppService.GetSkills.mockClear()
      mockAppService.GetWorkspaces.mockClear()
      mockAppService.GetSkills.mockResolvedValue(mockSkills)
      mockAppService.GetWorkspaces.mockResolvedValue(mockWorkspaces)

      // Simulate 5 rapid "skills-updated" events
      for (let i = 0; i < 5; i++) {
        eventHandlers["skills-updated"]?.()
      }

      // Advance past debounce period
      vi.advanceTimersByTime(350)
      await flushPromises()

      // Should have fetched only once despite 5 events
      expect(mockAppService.GetSkills).toHaveBeenCalledTimes(1)

      vi.useRealTimers()
    })

    it("cleans up debounce timer on unmount", async () => {
      vi.useFakeTimers()

      const eventHandlers: Record<string, Function> = {}
      const unsubscribeFns = { skills: vi.fn(), workspace: vi.fn() }
      ;(mockEvents.On as ReturnType<typeof vi.fn>).mockImplementation(
        (event: string, handler: Function) => {
          eventHandlers[event] = handler
          return event === "skills-updated"
            ? unsubscribeFns.skills
            : unsubscribeFns.workspace
        },
      )

      mockAppService.GetSkills.mockResolvedValue(mockSkills)
      mockAppService.GetWorkspaces.mockResolvedValue(mockWorkspaces)

      const wrapper = mount(BrowseSkills)
      await flushPromises()

      // Fire an event to start a debounce timer
      eventHandlers["skills-updated"]?.()

      // Unmount before debounce fires
      wrapper.unmount()

      // Clear and check that nothing fires after unmount
      mockAppService.GetSkills.mockClear()
      vi.advanceTimersByTime(350)
      await flushPromises()

      expect(mockAppService.GetSkills).not.toHaveBeenCalled()
      expect(unsubscribeFns.skills).toHaveBeenCalled()
      expect(unsubscribeFns.workspace).toHaveBeenCalled()

      vi.useRealTimers()
    })
  })

  describe("source URL validation", () => {
    it("renders source as link for https URLs", async () => {
      mockAppService.GetSkills.mockResolvedValue([
        {
          ...mockSkills[0],
          sourceUrl: "https://github.com/vercel-labs/skills",
        },
        {
          ...mockSkills[1],
          sourceUrl: "https://github.com/vercel-labs/skills",
        },
      ])

      const wrapper = await mountBrowseSkills()

      expect(wrapper.find(".group-source-link").exists()).toBe(true)
    })

    it("does not render source as link for SSH URLs", async () => {
      mockAppService.GetSkills.mockResolvedValue([
        {
          ...mockSkills[0],
          sourceUrl: "git@github.com:vercel-labs/skills.git",
        },
        {
          ...mockSkills[1],
          sourceUrl: "git@github.com:vercel-labs/skills.git",
        },
      ])

      const wrapper = await mountBrowseSkills()

      // SSH URL group should not get a clickable link
      const links = wrapper.findAll(".group-source-link")
      expect(links).toHaveLength(0)

      // But should still show source as plain text
      const plainSources = wrapper.findAll(
        ".group-source:not(.group-source-link)",
      )
      expect(plainSources.length).toBeGreaterThanOrEqual(1)
    })

    it("does not call Browser.OpenURL for SSH URLs", async () => {
      mockAppService.GetSkills.mockResolvedValue([
        {
          ...mockSkills[0],
          sourceUrl: "git@github.com:vercel-labs/skills.git",
        },
      ])

      const wrapper = await mountBrowseSkills()

      // No link should exist to click
      expect(wrapper.find(".group-source-link").exists()).toBe(false)
      expect(mockBrowser.OpenURL).not.toHaveBeenCalled()
    })
  })

  describe("new available skills CTA", () => {
    const { sourceUpdates } = useSkillUpdateChecker()

    afterEach(() => {
      sourceUpdates.value = {}
    })

    it("does not show CTA when no new skills available", async () => {
      const wrapper = await mountBrowseSkills()

      expect(wrapper.find(".new-skills-btn").exists()).toBe(false)
    })

    it("shows CTA with correct count when new skills available", async () => {
      sourceUpdates.value = {
        "vercel-labs/skills": {
          source: "vercel-labs/skills",
          hasUpdates: false,
          updatedSkillNames: [],
          newAvailableSkills: [
            {
              name: "new-skill-1",
              description: "New 1",
              alreadyInstalled: false,
            },
            {
              name: "new-skill-2",
              description: "New 2",
              alreadyInstalled: false,
            },
          ],
          checkedAt: "2025-01-30T10:00:00Z",
        },
      }

      const wrapper = await mountBrowseSkills()

      const btn = wrapper.find(".new-skills-btn")
      expect(btn.exists()).toBe(true)
      expect(btn.text()).toBe("2 other skills available")
    })

    it("uses singular 'skill' for count of 1", async () => {
      sourceUpdates.value = {
        "vercel-labs/skills": {
          source: "vercel-labs/skills",
          hasUpdates: false,
          updatedSkillNames: [],
          newAvailableSkills: [
            {
              name: "new-skill-1",
              description: "New 1",
              alreadyInstalled: false,
            },
          ],
          checkedAt: "2025-01-30T10:00:00Z",
        },
      }

      const wrapper = await mountBrowseSkills()

      const btn = wrapper.find(".new-skills-btn")
      expect(btn.exists()).toBe(true)
      expect(btn.text()).toBe("1 other skill available")
    })

    it("emits install-from-source when CTA clicked", async () => {
      sourceUpdates.value = {
        "vercel-labs/skills": {
          source: "vercel-labs/skills",
          hasUpdates: false,
          updatedSkillNames: [],
          newAvailableSkills: [
            {
              name: "new-skill-1",
              description: "New 1",
              alreadyInstalled: false,
            },
          ],
          checkedAt: "2025-01-30T10:00:00Z",
        },
      }

      const wrapper = await mountBrowseSkills()

      await wrapper.find(".new-skills-btn").trigger("click")
      await flushPromises()

      expect(wrapper.emitted("install-from-source")).toEqual([
        ["vercel-labs/skills"],
      ])
    })
  })
})
