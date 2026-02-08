import { describe, it, expect, vi, beforeEach } from "vitest"
import { mount, flushPromises } from "@vue/test-utils"
import InstallSkills from "./InstallSkills.vue"
import { mockAppService, mockEvents } from "../test/setup"
import type { WorkspaceInfo } from "../types/skill"

const mockWorkspaces: WorkspaceInfo[] = [
  {
    name: "default",
    description: "",
    skills: ["existing-skill"],
    isActive: true,
  },
  { name: "other", description: "", skills: [], isActive: false },
]

describe("InstallSkills", () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockAppService.GetWorkspaces.mockResolvedValue(mockWorkspaces)
  })

  function mountInstallSkills() {
    return mount(InstallSkills)
  }

  describe("step 1: source input", () => {
    it("shows header and subtitle", () => {
      const wrapper = mountInstallSkills()

      expect(wrapper.find(".install-header h2").text()).toBe("Install Skills")
      expect(wrapper.find(".subtitle").text()).toContain("GitHub")
    })

    it("shows input and fetch button", () => {
      const wrapper = mountInstallSkills()

      expect(wrapper.find("input").exists()).toBe(true)
      expect(wrapper.find(".install-btn").exists()).toBe(true)
      expect(wrapper.find(".install-btn").text()).toContain("Fetch")
    })

    it("disables fetch button when input is empty", () => {
      const wrapper = mountInstallSkills()

      const btn = wrapper.find(".install-btn")
      expect((btn.element as HTMLButtonElement).disabled).toBe(true)
    })

    it("shows examples when idle", () => {
      const wrapper = mountInstallSkills()

      expect(wrapper.find(".examples").exists()).toBe(true)
      expect(wrapper.findAll(".example")).toHaveLength(4)
    })
  })

  describe("example buttons", () => {
    it("fills input when example is clicked", async () => {
      const wrapper = mountInstallSkills()

      await wrapper.findAll(".example")[0].trigger("click")

      expect((wrapper.find("input").element as HTMLInputElement).value).toBe(
        "owner/repo",
      )
    })

    it("fills with branch example", async () => {
      const wrapper = mountInstallSkills()

      await wrapper.findAll(".example")[1].trigger("click")

      expect((wrapper.find("input").element as HTMLInputElement).value).toBe(
        "owner/repo#branch",
      )
    })

    it("fills with github URL example", async () => {
      const wrapper = mountInstallSkills()

      await wrapper.findAll(".example")[2].trigger("click")

      expect((wrapper.find("input").element as HTMLInputElement).value).toBe(
        "https://github.com/owner/repo",
      )
    })

    it("fills with zip URL example", async () => {
      const wrapper = mountInstallSkills()

      await wrapper.findAll(".example")[3].trigger("click")

      expect((wrapper.find("input").element as HTMLInputElement).value).toBe(
        "https://example.com/skills.zip",
      )
    })
  })

  describe("discover flow", () => {
    it("calls DiscoverFromSource with trimmed input on click", async () => {
      mockAppService.DiscoverFromSource.mockResolvedValue({
        skills: [{ name: "my-skill", description: "desc" }],
        source: "owner/repo",
        sourceType: "github",
      })

      const wrapper = mountInstallSkills()
      await wrapper.find("input").setValue("  owner/repo  ")
      await wrapper.find(".install-btn").trigger("click")
      await flushPromises()

      expect(mockAppService.DiscoverFromSource).toHaveBeenCalledWith(
        "owner/repo",
      )
    })

    it("calls DiscoverFromSource on Enter key", async () => {
      mockAppService.DiscoverFromSource.mockResolvedValue({
        skills: [{ name: "my-skill", description: "desc" }],
        source: "owner/repo",
        sourceType: "github",
      })

      const wrapper = mountInstallSkills()
      await wrapper.find("input").setValue("owner/repo")
      await wrapper.find("input").trigger("keyup.enter")
      await flushPromises()

      expect(mockAppService.DiscoverFromSource).toHaveBeenCalledWith(
        "owner/repo",
      )
    })

    it("shows loading state while fetching", async () => {
      mockAppService.DiscoverFromSource.mockReturnValue(new Promise(() => {}))

      const wrapper = mountInstallSkills()
      await wrapper.find("input").setValue("owner/repo")
      await wrapper.find(".install-btn").trigger("click")

      expect(wrapper.find(".spinner-sm").exists()).toBe(true)
      expect(wrapper.find(".install-btn").text()).toContain("Fetching")
      expect((wrapper.find("input").element as HTMLInputElement).disabled).toBe(
        true,
      )
      expect(wrapper.find(".examples").exists()).toBe(false)
    })

    it("does not call DiscoverFromSource with empty input", async () => {
      const wrapper = mountInstallSkills()
      await wrapper.find("input").setValue("   ")
      await wrapper.find(".install-btn").trigger("click")
      await flushPromises()

      expect(mockAppService.DiscoverFromSource).not.toHaveBeenCalled()
    })

    it("shows error on discover failure", async () => {
      mockAppService.DiscoverFromSource.mockRejectedValue(
        new Error("Network error"),
      )

      const wrapper = mountInstallSkills()
      await wrapper.find("input").setValue("owner/repo")
      await wrapper.find(".install-btn").trigger("click")
      await flushPromises()

      expect(wrapper.find(".result-error").exists()).toBe(true)
      expect(wrapper.text()).toContain("Network error")
    })

    it("shows fallback error for non-Error exceptions", async () => {
      mockAppService.DiscoverFromSource.mockRejectedValue("something broke")

      const wrapper = mountInstallSkills()
      await wrapper.find("input").setValue("owner/repo")
      await wrapper.find(".install-btn").trigger("click")
      await flushPromises()

      expect(wrapper.find(".result-error").exists()).toBe(true)
      expect(wrapper.text()).toContain("Failed to fetch skills")
    })
  })

  describe("step 2: review discovered skills", () => {
    async function mountAtReviewStep() {
      mockAppService.DiscoverFromSource.mockResolvedValue({
        skills: [
          { name: "skill-alpha", description: "Alpha desc" },
          { name: "skill-beta", description: "Beta desc" },
        ],
        source: "owner/repo",
        sourceType: "github",
      })

      const wrapper = mountInstallSkills()
      await wrapper.find("input").setValue("owner/repo")
      await wrapper.find(".install-btn").trigger("click")
      await flushPromises()
      return wrapper
    }

    it("shows review step after successful discover", async () => {
      const wrapper = await mountAtReviewStep()

      expect(wrapper.find(".step-review").exists()).toBe(true)
      expect(wrapper.find(".step-source").exists()).toBe(false)
    })

    it("shows source badge and name", async () => {
      const wrapper = await mountAtReviewStep()

      expect(wrapper.find(".source-badge").text()).toBe("github")
      expect(wrapper.find(".source-label").text()).toBe("owner/repo")
    })

    it("shows discovered skills with checkboxes", async () => {
      const wrapper = await mountAtReviewStep()

      const items = wrapper.findAll(".skill-check-item")
      expect(items).toHaveLength(2)
      expect(wrapper.text()).toContain("skill-alpha")
      expect(wrapper.text()).toContain("skill-beta")
    })

    it("all skills are checked by default", async () => {
      const wrapper = await mountAtReviewStep()

      const checkboxes = wrapper.findAll(
        '.skill-check-item input[type="checkbox"]',
      )
      expect(checkboxes).toHaveLength(2)
      expect((checkboxes[0].element as HTMLInputElement).checked).toBe(true)
      expect((checkboxes[1].element as HTMLInputElement).checked).toBe(true)
    })

    it("shows found count in description", async () => {
      const wrapper = await mountAtReviewStep()

      expect(wrapper.find(".step-description").text()).toContain(
        "Found 2 skills",
      )
    })

    it("can uncheck a skill", async () => {
      const wrapper = await mountAtReviewStep()

      const checkboxes = wrapper.findAll(
        '.skill-check-item input[type="checkbox"]',
      )
      await checkboxes[0].trigger("change")

      expect((checkboxes[0].element as HTMLInputElement).checked).toBe(false)
    })

    it("disables continue when no skills selected", async () => {
      const wrapper = await mountAtReviewStep()

      const checkboxes = wrapper.findAll(
        '.skill-check-item input[type="checkbox"]',
      )
      await checkboxes[0].trigger("change")
      await checkboxes[1].trigger("change")

      const continueBtn = wrapper.find(".btn-primary")
      expect((continueBtn.element as HTMLButtonElement).disabled).toBe(true)
    })

    it("cancel calls CancelDiscover and returns to source step", async () => {
      mockAppService.CancelDiscover.mockResolvedValue(undefined)

      const wrapper = await mountAtReviewStep()
      await wrapper.find(".btn-secondary").trigger("click")
      await flushPromises()

      expect(mockAppService.CancelDiscover).toHaveBeenCalled()
      expect(wrapper.find(".step-source").exists()).toBe(true)
      expect(wrapper.find(".step-review").exists()).toBe(false)
    })
  })

  describe("step 3: workspace selection", () => {
    async function mountAtWorkspaceStep() {
      mockAppService.DiscoverFromSource.mockResolvedValue({
        skills: [
          { name: "skill-alpha", description: "Alpha desc" },
          { name: "skill-beta", description: "Beta desc" },
        ],
        source: "owner/repo",
        sourceType: "github",
      })

      const wrapper = mountInstallSkills()
      await wrapper.find("input").setValue("owner/repo")
      await wrapper.find(".install-btn").trigger("click")
      await flushPromises()

      // Click Continue to go to workspace step
      await wrapper.find(".btn-primary").trigger("click")
      await flushPromises()
      return wrapper
    }

    it("shows workspace step after continue", async () => {
      const wrapper = await mountAtWorkspaceStep()

      expect(wrapper.find(".step-workspaces").exists()).toBe(true)
    })

    it("shows workspace list with checkboxes", async () => {
      const wrapper = await mountAtWorkspaceStep()

      const items = wrapper.findAll(".workspace-check-item")
      expect(items).toHaveLength(2)
      expect(wrapper.text()).toContain("default")
      expect(wrapper.text()).toContain("other")
    })

    it("no workspaces are checked by default", async () => {
      const wrapper = await mountAtWorkspaceStep()

      const checkboxes = wrapper.findAll(
        '.workspace-check-item input[type="checkbox"]',
      )
      for (const cb of checkboxes) {
        expect((cb.element as HTMLInputElement).checked).toBe(false)
      }
    })

    it("shows active badge on active workspace", async () => {
      const wrapper = await mountAtWorkspaceStep()

      expect(wrapper.find(".active-badge").exists()).toBe(true)
      expect(wrapper.find(".active-badge").text()).toBe("active")
    })

    it("shows optional hint", async () => {
      const wrapper = await mountAtWorkspaceStep()

      expect(wrapper.find(".optional-hint").text()).toBe("(optional)")
    })

    it("back button returns to review step", async () => {
      const wrapper = await mountAtWorkspaceStep()

      await wrapper.find(".btn-secondary").trigger("click")

      expect(wrapper.find(".step-review").exists()).toBe(true)
      expect(wrapper.find(".step-workspaces").exists()).toBe(false)
    })

    it("shows install button with skill count", async () => {
      const wrapper = await mountAtWorkspaceStep()

      expect(wrapper.find(".btn-primary").text()).toContain("Install 2 skills")
    })
  })

  describe("confirm install", () => {
    async function mountAndInstall(opts: { selectWorkspace?: boolean } = {}) {
      mockAppService.DiscoverFromSource.mockResolvedValue({
        skills: [
          { name: "skill-alpha", description: "Alpha desc" },
          { name: "skill-beta", description: "Beta desc" },
        ],
        source: "owner/repo",
        sourceType: "github",
      })
      mockAppService.ConfirmInstall.mockResolvedValue({
        success: true,
        skillsCount: 2,
        skillNames: ["skill-alpha", "skill-beta"],
        errorMessage: "",
      })

      const wrapper = mountInstallSkills()
      await wrapper.find("input").setValue("owner/repo")
      await wrapper.find(".install-btn").trigger("click")
      await flushPromises()

      // Continue to workspace step
      await wrapper.find(".btn-primary").trigger("click")
      await flushPromises()

      // Optionally select a workspace
      if (opts.selectWorkspace) {
        const checkboxes = wrapper.findAll(
          '.workspace-check-item input[type="checkbox"]',
        )
        await checkboxes[0].trigger("change")
      }

      // Click install
      await wrapper.find(".btn-primary").trigger("click")
      await flushPromises()
      return wrapper
    }

    it("calls ConfirmInstall with selected skills and no workspaces", async () => {
      await mountAndInstall()

      expect(mockAppService.ConfirmInstall).toHaveBeenCalledWith(
        expect.arrayContaining(["skill-alpha", "skill-beta"]),
        [],
      )
    })

    it("calls ConfirmInstall with selected workspaces", async () => {
      await mountAndInstall({ selectWorkspace: true })

      expect(mockAppService.ConfirmInstall).toHaveBeenCalledWith(
        expect.arrayContaining(["skill-alpha", "skill-beta"]),
        ["default"],
      )
    })

    it("shows success result with skill count and names", async () => {
      const wrapper = await mountAndInstall()

      expect(wrapper.find(".step-result").exists()).toBe(true)
      expect(wrapper.find(".result-success").exists()).toBe(true)
      expect(wrapper.find(".result-title").text()).toContain("2 skills")
      expect(wrapper.find(".result-names").text()).toBe(
        "skill-alpha, skill-beta",
      )
    })

    it("shows singular 'skill' for count of 1", async () => {
      mockAppService.DiscoverFromSource.mockResolvedValue({
        skills: [{ name: "my-skill", description: "" }],
        source: "owner/repo",
        sourceType: "github",
      })
      mockAppService.ConfirmInstall.mockResolvedValue({
        success: true,
        skillsCount: 1,
        skillNames: ["my-skill"],
        errorMessage: "",
      })

      const wrapper = mountInstallSkills()
      await wrapper.find("input").setValue("owner/repo")
      await wrapper.find(".install-btn").trigger("click")
      await flushPromises()
      await wrapper.find(".btn-primary").trigger("click")
      await flushPromises()
      await wrapper.find(".btn-primary").trigger("click")
      await flushPromises()

      expect(wrapper.find(".result-title").text()).toContain("1 skill")
      expect(wrapper.find(".result-title").text()).not.toContain("1 skills")
    })

    it("shows error result on install failure", async () => {
      mockAppService.DiscoverFromSource.mockResolvedValue({
        skills: [{ name: "my-skill", description: "" }],
        source: "owner/repo",
        sourceType: "github",
      })
      mockAppService.ConfirmInstall.mockRejectedValue(
        new Error("Install failed"),
      )

      const wrapper = mountInstallSkills()
      await wrapper.find("input").setValue("owner/repo")
      await wrapper.find(".install-btn").trigger("click")
      await flushPromises()
      await wrapper.find(".btn-primary").trigger("click")
      await flushPromises()
      await wrapper.find(".btn-primary").trigger("click")
      await flushPromises()

      // On exception, goes back to source step with error
      expect(wrapper.find(".step-source").exists()).toBe(true)
    })

    it("shows backend error in result step", async () => {
      mockAppService.DiscoverFromSource.mockResolvedValue({
        skills: [{ name: "my-skill", description: "" }],
        source: "owner/repo",
        sourceType: "github",
      })
      mockAppService.ConfirmInstall.mockResolvedValue({
        success: false,
        skillsCount: 0,
        skillNames: [],
        errorMessage: "no skills found in source",
      })

      const wrapper = mountInstallSkills()
      await wrapper.find("input").setValue("owner/repo")
      await wrapper.find(".install-btn").trigger("click")
      await flushPromises()
      await wrapper.find(".btn-primary").trigger("click")
      await flushPromises()
      await wrapper.find(".btn-primary").trigger("click")
      await flushPromises()

      expect(wrapper.find(".step-result").exists()).toBe(true)
      expect(wrapper.find(".result-error").exists()).toBe(true)
      expect(wrapper.text()).toContain("no skills found in source")
    })

    it("install more button resets to source step", async () => {
      const wrapper = await mountAndInstall()

      expect(wrapper.find(".step-result").exists()).toBe(true)
      await wrapper.find(".reset-btn").trigger("click")

      expect(wrapper.find(".step-source").exists()).toBe(true)
      expect((wrapper.find("input").element as HTMLInputElement).value).toBe("")
    })
  })

  describe("skill toggle re-add", () => {
    it("can uncheck and re-check a skill", async () => {
      mockAppService.DiscoverFromSource.mockResolvedValue({
        skills: [
          { name: "skill-alpha", description: "Alpha desc" },
          { name: "skill-beta", description: "Beta desc" },
        ],
        source: "owner/repo",
        sourceType: "github",
      })

      const wrapper = mountInstallSkills()
      await wrapper.find("input").setValue("owner/repo")
      await wrapper.find(".install-btn").trigger("click")
      await flushPromises()

      const checkboxes = wrapper.findAll(
        '.skill-check-item input[type="checkbox"]',
      )
      // Uncheck skill-alpha
      await checkboxes[0].trigger("change")
      expect((checkboxes[0].element as HTMLInputElement).checked).toBe(false)

      // Re-check skill-alpha
      await checkboxes[0].trigger("change")
      expect((checkboxes[0].element as HTMLInputElement).checked).toBe(true)
    })
  })

  describe("workspace toggle re-add", () => {
    it("can check and uncheck a workspace", async () => {
      mockAppService.DiscoverFromSource.mockResolvedValue({
        skills: [{ name: "my-skill", description: "" }],
        source: "owner/repo",
        sourceType: "github",
      })

      const wrapper = mountInstallSkills()
      await wrapper.find("input").setValue("owner/repo")
      await wrapper.find(".install-btn").trigger("click")
      await flushPromises()

      // Continue to workspace step
      await wrapper.find(".btn-primary").trigger("click")
      await flushPromises()

      const checkboxes = wrapper.findAll(
        '.workspace-check-item input[type="checkbox"]',
      )
      // Check default workspace
      await checkboxes[0].trigger("change")
      expect((checkboxes[0].element as HTMLInputElement).checked).toBe(true)

      // Uncheck default workspace
      await checkboxes[0].trigger("change")
      expect((checkboxes[0].element as HTMLInputElement).checked).toBe(false)
    })
  })

  describe("empty workspaces", () => {
    it("shows empty hint when no workspaces exist", async () => {
      mockAppService.GetWorkspaces.mockResolvedValue([])
      mockAppService.DiscoverFromSource.mockResolvedValue({
        skills: [{ name: "my-skill", description: "" }],
        source: "owner/repo",
        sourceType: "github",
      })

      const wrapper = mountInstallSkills()
      await wrapper.find("input").setValue("owner/repo")
      await wrapper.find(".install-btn").trigger("click")
      await flushPromises()

      // Continue to workspace step
      await wrapper.find(".btn-primary").trigger("click")
      await flushPromises()

      expect(wrapper.find(".empty-hint").exists()).toBe(true)
      expect(wrapper.text()).toContain("No workspaces found")
    })
  })

  describe("cleanup on unmount", () => {
    it("does not throw when unmounting", () => {
      const wrapper = mountInstallSkills()
      expect(() => wrapper.unmount()).not.toThrow()
    })
  })

  describe("dismiss error", () => {
    it("dismisses error on close button click", async () => {
      mockAppService.DiscoverFromSource.mockRejectedValue(
        new Error("Network error"),
      )

      const wrapper = mountInstallSkills()
      await wrapper.find("input").setValue("owner/repo")
      await wrapper.find(".install-btn").trigger("click")
      await flushPromises()

      expect(wrapper.find(".result-error").exists()).toBe(true)
      await wrapper.find(".dismiss-btn").trigger("click")
      expect(wrapper.find(".result-error").exists()).toBe(false)
    })
  })

  describe("workspace-changed event", () => {
    it("refreshes workspace list when workspace-changed fires", async () => {
      // Capture the callback registered for 'workspace-changed'
      type Callback = () => void
      let workspaceChangedCallback: Callback | null = null
      ;(mockEvents.On as ReturnType<typeof vi.fn>).mockImplementation(
        (event: string, cb: Callback) => {
          if (event === "workspace-changed") {
            workspaceChangedCallback = cb
          }
          return vi.fn()
        },
      )

      const workspacesAfterDelete: WorkspaceInfo[] = [
        { name: "default", description: "", skills: [], isActive: true },
      ]

      mockAppService.DiscoverFromSource.mockResolvedValue({
        skills: [{ name: "my-skill", description: "" }],
        source: "owner/repo",
        sourceType: "github",
      })

      const wrapper = mountInstallSkills()
      await flushPromises()

      // Navigate to workspace step
      await wrapper.find("input").setValue("owner/repo")
      await wrapper.find(".install-btn").trigger("click")
      await flushPromises()
      await wrapper.find(".btn-primary").trigger("click")
      await flushPromises()

      // Verify both workspaces are shown initially
      expect(wrapper.findAll(".workspace-check-item")).toHaveLength(2)

      // Simulate a workspace deletion by returning fewer workspaces
      mockAppService.GetWorkspaces.mockResolvedValue(workspacesAfterDelete)

      // Fire the workspace-changed event
      expect(workspaceChangedCallback).not.toBeNull()
      workspaceChangedCallback!()
      await flushPromises()

      // Workspace list should now reflect the deletion
      expect(wrapper.findAll(".workspace-check-item")).toHaveLength(1)
      expect(wrapper.text()).toContain("default")
      expect(wrapper.text()).not.toContain("other")
    })
  })
})
