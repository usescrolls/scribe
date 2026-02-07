import { describe, it, expect, vi, beforeEach } from "vitest"
import { mount, flushPromises } from "@vue/test-utils"
import InstallSkills from "./InstallSkills.vue"
import { mockAppService } from "../test/setup"

describe("InstallSkills", () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  function mountInstallSkills() {
    return mount(InstallSkills)
  }

  describe("rendering", () => {
    it("shows header and subtitle", () => {
      const wrapper = mountInstallSkills()

      expect(wrapper.find(".install-header h2").text()).toBe("Install Skills")
      expect(wrapper.find(".subtitle").text()).toContain("GitHub")
    })

    it("shows input and install button", () => {
      const wrapper = mountInstallSkills()

      expect(wrapper.find("input").exists()).toBe(true)
      expect(wrapper.find(".install-btn").exists()).toBe(true)
    })

    it("disables install button when input is empty", () => {
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

    it("fills with github URL example", async () => {
      const wrapper = mountInstallSkills()

      await wrapper.findAll(".example")[2].trigger("click")

      expect((wrapper.find("input").element as HTMLInputElement).value).toBe(
        "https://github.com/owner/repo",
      )
    })
  })

  describe("install flow", () => {
    it("calls InstallFromSource with trimmed input on click", async () => {
      mockAppService.InstallFromSource.mockResolvedValue({
        success: true,
        skillsCount: 1,
        skillNames: ["my-skill"],
        errorMessage: "",
      })

      const wrapper = mountInstallSkills()
      await wrapper.find("input").setValue("  owner/repo  ")
      await wrapper.find(".install-btn").trigger("click")
      await flushPromises()

      expect(mockAppService.InstallFromSource).toHaveBeenCalledWith(
        "owner/repo",
      )
    })

    it("calls InstallFromSource on Enter key", async () => {
      mockAppService.InstallFromSource.mockResolvedValue({
        success: true,
        skillsCount: 1,
        skillNames: ["my-skill"],
        errorMessage: "",
      })

      const wrapper = mountInstallSkills()
      await wrapper.find("input").setValue("owner/repo")
      await wrapper.find("input").trigger("keyup.enter")
      await flushPromises()

      expect(mockAppService.InstallFromSource).toHaveBeenCalledWith(
        "owner/repo",
      )
    })

    it("shows loading state while installing", async () => {
      // Never resolve so we stay in loading state
      mockAppService.InstallFromSource.mockReturnValue(new Promise(() => {}))

      const wrapper = mountInstallSkills()
      await wrapper.find("input").setValue("owner/repo")
      await wrapper.find(".install-btn").trigger("click")

      expect(wrapper.find(".spinner-sm").exists()).toBe(true)
      expect(wrapper.find(".install-btn").text()).toContain("Installing")
      expect((wrapper.find("input").element as HTMLInputElement).disabled).toBe(
        true,
      )
      // Examples should be hidden during install
      expect(wrapper.find(".examples").exists()).toBe(false)
    })

    it("does not call InstallFromSource with empty input", async () => {
      const wrapper = mountInstallSkills()
      await wrapper.find("input").setValue("   ")
      await wrapper.find(".install-btn").trigger("click")
      await flushPromises()

      expect(mockAppService.InstallFromSource).not.toHaveBeenCalled()
    })
  })

  describe("success result", () => {
    it("shows success message with skill count and names", async () => {
      mockAppService.InstallFromSource.mockResolvedValue({
        success: true,
        skillsCount: 2,
        skillNames: ["skill-alpha", "skill-beta"],
        errorMessage: "",
      })

      const wrapper = mountInstallSkills()
      await wrapper.find("input").setValue("owner/repo")
      await wrapper.find(".install-btn").trigger("click")
      await flushPromises()

      expect(wrapper.find(".result-success").exists()).toBe(true)
      expect(wrapper.find(".result-title").text()).toContain("2 skills")
      expect(wrapper.find(".result-names").text()).toBe(
        "skill-alpha, skill-beta",
      )
    })

    it("shows singular 'skill' for count of 1", async () => {
      mockAppService.InstallFromSource.mockResolvedValue({
        success: true,
        skillsCount: 1,
        skillNames: ["my-skill"],
        errorMessage: "",
      })

      const wrapper = mountInstallSkills()
      await wrapper.find("input").setValue("owner/repo")
      await wrapper.find(".install-btn").trigger("click")
      await flushPromises()

      expect(wrapper.find(".result-title").text()).toContain("1 skill")
      expect(wrapper.find(".result-title").text()).not.toContain("1 skills")
    })

    it("clears input on success", async () => {
      mockAppService.InstallFromSource.mockResolvedValue({
        success: true,
        skillsCount: 1,
        skillNames: ["my-skill"],
        errorMessage: "",
      })

      const wrapper = mountInstallSkills()
      await wrapper.find("input").setValue("owner/repo")
      await wrapper.find(".install-btn").trigger("click")
      await flushPromises()

      expect((wrapper.find("input").element as HTMLInputElement).value).toBe("")
    })

    it("dismisses result on close button click", async () => {
      mockAppService.InstallFromSource.mockResolvedValue({
        success: true,
        skillsCount: 1,
        skillNames: ["my-skill"],
        errorMessage: "",
      })

      const wrapper = mountInstallSkills()
      await wrapper.find("input").setValue("owner/repo")
      await wrapper.find(".install-btn").trigger("click")
      await flushPromises()

      expect(wrapper.find(".result").exists()).toBe(true)
      await wrapper.find(".dismiss-btn").trigger("click")
      expect(wrapper.find(".result").exists()).toBe(false)
    })
  })

  describe("error result", () => {
    it("shows error message from backend", async () => {
      mockAppService.InstallFromSource.mockResolvedValue({
        success: false,
        skillsCount: 0,
        skillNames: [],
        errorMessage: "no skills found in source",
      })

      const wrapper = mountInstallSkills()
      await wrapper.find("input").setValue("owner/empty-repo")
      await wrapper.find(".install-btn").trigger("click")
      await flushPromises()

      expect(wrapper.find(".result-error").exists()).toBe(true)
      expect(wrapper.text()).toContain("no skills found in source")
    })

    it("shows error message on exception", async () => {
      mockAppService.InstallFromSource.mockRejectedValue(
        new Error("Network error"),
      )

      const wrapper = mountInstallSkills()
      await wrapper.find("input").setValue("owner/repo")
      await wrapper.find(".install-btn").trigger("click")
      await flushPromises()

      expect(wrapper.find(".result-error").exists()).toBe(true)
      expect(wrapper.text()).toContain("Network error")
    })

    it("shows fallback error message for non-Error exceptions", async () => {
      mockAppService.InstallFromSource.mockRejectedValue("something broke")

      const wrapper = mountInstallSkills()
      await wrapper.find("input").setValue("owner/repo")
      await wrapper.find(".install-btn").trigger("click")
      await flushPromises()

      expect(wrapper.find(".result-error").exists()).toBe(true)
      expect(wrapper.text()).toContain("Installation failed")
    })

    it("does not clear input on error", async () => {
      mockAppService.InstallFromSource.mockResolvedValue({
        success: false,
        skillsCount: 0,
        skillNames: [],
        errorMessage: "invalid source",
      })

      const wrapper = mountInstallSkills()
      await wrapper.find("input").setValue("bad-input")
      await wrapper.find(".install-btn").trigger("click")
      await flushPromises()

      expect((wrapper.find("input").element as HTMLInputElement).value).toBe(
        "bad-input",
      )
    })
  })
})
