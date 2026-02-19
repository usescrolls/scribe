import { describe, it, expect, vi, beforeEach } from "vitest"
import { mount, flushPromises } from "@vue/test-utils"
import MarketplaceSkills from "./MarketplaceSkills.vue"
import { mockAppService } from "../test/setup"
import type { MarketplaceResult } from "../types/skill"

const mockResults: MarketplaceResult = {
  repos: [
    {
      owner: "alice",
      name: "skills",
      fullName: "alice/skills",
      description: "A collection of useful skills",
      url: "https://github.com/alice/skills",
      stars: 42,
      skillCount: 3,
      provider: "github",
    },
    {
      owner: "bob",
      name: "tools",
      fullName: "bob/tools",
      description: "Handy tools",
      url: "https://github.com/bob/tools",
      stars: 1500,
      skillCount: 1,
      provider: "github",
    },
  ],
  totalCount: 2,
}

describe("MarketplaceSkills", () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  function mountMarketplace() {
    return mount(MarketplaceSkills)
  }

  describe("initial state", () => {
    it("shows header and subtitle", () => {
      const wrapper = mountMarketplace()

      expect(wrapper.find(".marketplace-header h2").text()).toBe("Marketplace")
      expect(wrapper.find(".subtitle").text()).toContain("GitHub")
    })

    it("shows search input and button", () => {
      const wrapper = mountMarketplace()

      expect(wrapper.find("input").exists()).toBe(true)
      expect(wrapper.find(".search-btn").exists()).toBe(true)
      expect(wrapper.find(".search-btn").text()).toContain("Search")
    })

    it("disables search button when input is empty", () => {
      const wrapper = mountMarketplace()

      const btn = wrapper.find(".search-btn")
      expect((btn.element as HTMLButtonElement).disabled).toBe(true)
    })

    it("shows initial hint when no search has been done", () => {
      const wrapper = mountMarketplace()

      expect(wrapper.find(".initial-state").exists()).toBe(true)
      expect(wrapper.find(".initial-hint").text()).toContain("Search GitHub")
    })
  })

  describe("search flow", () => {
    it("calls SearchMarketplace with provider, query, and page on click", async () => {
      mockAppService.SearchMarketplace.mockResolvedValue(mockResults)

      const wrapper = mountMarketplace()
      await wrapper.find("input").setValue("commit")
      await wrapper.find(".search-btn").trigger("click")
      await flushPromises()

      expect(mockAppService.SearchMarketplace).toHaveBeenCalledWith(
        "github",
        "commit",
        1,
      )
    })

    it("calls SearchMarketplace on Enter key", async () => {
      mockAppService.SearchMarketplace.mockResolvedValue(mockResults)

      const wrapper = mountMarketplace()
      await wrapper.find("input").setValue("review")
      await wrapper.find("input").trigger("keyup.enter")
      await flushPromises()

      expect(mockAppService.SearchMarketplace).toHaveBeenCalledWith(
        "github",
        "review",
        1,
      )
    })

    it("does not call SearchMarketplace with empty input", async () => {
      const wrapper = mountMarketplace()
      await wrapper.find("input").setValue("   ")
      await wrapper.find(".search-btn").trigger("click")
      await flushPromises()

      expect(mockAppService.SearchMarketplace).not.toHaveBeenCalled()
    })

    it("shows loading state while searching", async () => {
      mockAppService.SearchMarketplace.mockReturnValue(new Promise(() => {}))

      const wrapper = mountMarketplace()
      await wrapper.find("input").setValue("test")
      await wrapper.find(".search-btn").trigger("click")

      expect(wrapper.find(".spinner-sm").exists()).toBe(true)
      expect(wrapper.find(".search-btn").text()).toContain("Searching")
      expect((wrapper.find("input").element as HTMLInputElement).disabled).toBe(
        true,
      )
    })

    it("shows error on search failure", async () => {
      mockAppService.SearchMarketplace.mockRejectedValue(
        new Error("GitHub API rate limit exceeded"),
      )

      const wrapper = mountMarketplace()
      await wrapper.find("input").setValue("test")
      await wrapper.find(".search-btn").trigger("click")
      await flushPromises()

      expect(wrapper.find(".result-error").exists()).toBe(true)
      expect(wrapper.text()).toContain("rate limit")
    })

    it("dismisses error on close button click", async () => {
      mockAppService.SearchMarketplace.mockRejectedValue(
        new Error("Network error"),
      )

      const wrapper = mountMarketplace()
      await wrapper.find("input").setValue("test")
      await wrapper.find(".search-btn").trigger("click")
      await flushPromises()

      expect(wrapper.find(".result-error").exists()).toBe(true)
      await wrapper.find(".dismiss-btn").trigger("click")
      expect(wrapper.find(".result-error").exists()).toBe(false)
    })
  })

  describe("results display", () => {
    async function mountWithResults() {
      mockAppService.SearchMarketplace.mockResolvedValue(mockResults)

      const wrapper = mountMarketplace()
      await wrapper.find("input").setValue("skills")
      await wrapper.find(".search-btn").trigger("click")
      await flushPromises()
      return wrapper
    }

    it("shows results count", async () => {
      const wrapper = await mountWithResults()

      expect(wrapper.find(".results-count").text()).toContain("2 results")
    })

    it("renders repo cards", async () => {
      const wrapper = await mountWithResults()

      const cards = wrapper.findAll(".repo-card")
      expect(cards).toHaveLength(2)
    })

    it("shows repo full name", async () => {
      const wrapper = await mountWithResults()

      const names = wrapper.findAll(".repo-name")
      expect(names[0].text()).toBe("alice/skills")
      expect(names[1].text()).toBe("bob/tools")
    })

    it("shows repo description", async () => {
      const wrapper = await mountWithResults()

      const descs = wrapper.findAll(".repo-desc")
      expect(descs[0].text()).toBe("A collection of useful skills")
      expect(descs[1].text()).toBe("Handy tools")
    })

    it("shows stars badge with formatting", async () => {
      const wrapper = await mountWithResults()

      const badges = wrapper.findAll(".badge-stars")
      expect(badges[0].text()).toBe("42")
      expect(badges[1].text()).toBe("1.5k")
    })

    it("shows skill count badge", async () => {
      const wrapper = await mountWithResults()

      const badges = wrapper.findAll(".badge-skills")
      expect(badges[0].text()).toBe("3 skills")
      expect(badges[1].text()).toBe("1 skill")
    })

    it("shows install button on each card", async () => {
      const wrapper = await mountWithResults()

      const buttons = wrapper.findAll(".btn-primary")
      expect(buttons).toHaveLength(2)
      expect(buttons[0].text()).toBe("Install")
    })

    it("shows singular 'result' for count of 1", async () => {
      mockAppService.SearchMarketplace.mockResolvedValue({
        repos: [mockResults.repos[0]],
        totalCount: 1,
      })

      const wrapper = mountMarketplace()
      await wrapper.find("input").setValue("test")
      await wrapper.find(".search-btn").trigger("click")
      await flushPromises()

      expect(wrapper.find(".results-count").text()).toContain("1 result")
      expect(wrapper.find(".results-count").text()).not.toContain("1 results")
    })
  })

  describe("no results", () => {
    it("shows no results message when search returns empty", async () => {
      mockAppService.SearchMarketplace.mockResolvedValue({
        repos: [],
        totalCount: 0,
      })

      const wrapper = mountMarketplace()
      await wrapper.find("input").setValue("nonexistent-xyz")
      await wrapper.find(".search-btn").trigger("click")
      await flushPromises()

      expect(wrapper.find(".empty-state").exists()).toBe(true)
      expect(wrapper.text()).toContain("nonexistent-xyz")
    })
  })

  describe("install flow", () => {
    async function mountWithResults() {
      mockAppService.SearchMarketplace.mockResolvedValue(mockResults)

      const wrapper = mountMarketplace()
      await wrapper.find("input").setValue("skills")
      await wrapper.find(".search-btn").trigger("click")
      await flushPromises()
      return wrapper
    }

    it("calls InstallFromSource with repo fullName on install click", async () => {
      mockAppService.InstallFromSource.mockResolvedValue({
        success: true,
        skillsCount: 3,
        skillNames: ["skill-a", "skill-b", "skill-c"],
        errorMessage: "",
      })

      const wrapper = await mountWithResults()
      const installBtns = wrapper.findAll(".btn-primary")
      await installBtns[0].trigger("click")
      await flushPromises()

      expect(mockAppService.InstallFromSource).toHaveBeenCalledWith(
        "alice/skills",
      )
    })

    it("shows installing state on the clicked card", async () => {
      mockAppService.InstallFromSource.mockReturnValue(new Promise(() => {}))

      const wrapper = await mountWithResults()
      const installBtns = wrapper.findAll(".btn-primary")
      await installBtns[0].trigger("click")

      expect(wrapper.find(".btn-installing").exists()).toBe(true)
      expect(wrapper.find(".btn-installing").text()).toContain("Installing")
    })

    it("shows installed state after successful install", async () => {
      mockAppService.InstallFromSource.mockResolvedValue({
        success: true,
        skillsCount: 3,
        skillNames: ["skill-a", "skill-b", "skill-c"],
        errorMessage: "",
      })

      const wrapper = await mountWithResults()
      const installBtns = wrapper.findAll(".btn-primary")
      await installBtns[0].trigger("click")
      await flushPromises()

      expect(wrapper.find(".btn-installed").exists()).toBe(true)
      expect(wrapper.find(".btn-installed").text()).toContain("Installed")
    })

    it("shows success toast after install", async () => {
      mockAppService.InstallFromSource.mockResolvedValue({
        success: true,
        skillsCount: 2,
        skillNames: ["skill-a", "skill-b"],
        errorMessage: "",
      })

      const wrapper = await mountWithResults()
      const installBtns = wrapper.findAll(".btn-primary")
      await installBtns[0].trigger("click")
      await flushPromises()

      expect(wrapper.find(".toast.success").exists()).toBe(true)
      expect(wrapper.find(".toast.success").text()).toContain("2 skills")
    })

    it("shows error toast on install failure", async () => {
      mockAppService.InstallFromSource.mockResolvedValue({
        success: false,
        skillsCount: 0,
        skillNames: [],
        errorMessage: "Failed to clone repository",
      })

      const wrapper = await mountWithResults()
      const installBtns = wrapper.findAll(".btn-primary")
      await installBtns[0].trigger("click")
      await flushPromises()

      expect(wrapper.find(".toast.error").exists()).toBe(true)
      expect(wrapper.find(".toast.error").text()).toContain(
        "Failed to clone repository",
      )
    })

    it("shows error toast on install exception", async () => {
      mockAppService.InstallFromSource.mockRejectedValue(
        new Error("Network timeout"),
      )

      const wrapper = await mountWithResults()
      const installBtns = wrapper.findAll(".btn-primary")
      await installBtns[0].trigger("click")
      await flushPromises()

      expect(wrapper.find(".toast.error").exists()).toBe(true)
      expect(wrapper.find(".toast.error").text()).toContain("Network timeout")
    })

    it("keeps installed state after install (does not revert)", async () => {
      mockAppService.InstallFromSource.mockResolvedValue({
        success: true,
        skillsCount: 1,
        skillNames: ["skill-a"],
        errorMessage: "",
      })

      const wrapper = await mountWithResults()
      const installBtns = wrapper.findAll(".btn-primary")
      await installBtns[0].trigger("click")
      await flushPromises()

      // The first card should show "Installed", second still shows "Install"
      expect(wrapper.find(".btn-installed").exists()).toBe(true)
      const remainingInstallBtns = wrapper.findAll(".btn-primary")
      expect(remainingInstallBtns).toHaveLength(1)
    })
  })

  describe("cleanup on unmount", () => {
    it("does not throw when unmounting", () => {
      const wrapper = mountMarketplace()
      expect(() => wrapper.unmount()).not.toThrow()
    })
  })
})
