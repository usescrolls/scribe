import { describe, it, expect, vi, beforeEach } from "vitest"
import { mount, flushPromises } from "@vue/test-utils"
import MarketplaceSkills from "./MarketplaceSkills.vue"
import { mockAppService, mockBrowser } from "../test/setup"
import type { MarketplaceResult } from "../types/skill"

const mockResults: MarketplaceResult = {
  repos: [
    {
      owner: "alice",
      name: "skills",
      fullName: "alice/skills",
      description: "A collection of useful skills",
      url: "https://github.com/alice/skills",
      avatarUrl: "https://avatars.githubusercontent.com/u/alice",
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
      avatarUrl: "https://avatars.githubusercontent.com/u/bob",
      stars: 1500,
      skillCount: 1,
      provider: "github",
    },
  ],
  totalCount: 2,
}

const emptyResults: MarketplaceResult = { repos: [], totalCount: 0 }

describe("MarketplaceSkills", () => {
  beforeEach(() => {
    vi.clearAllMocks()
    // Default: auto-search on mount returns empty
    mockAppService.SearchMarketplace.mockResolvedValue(emptyResults)
  })

  function mountMarketplace() {
    return mount(MarketplaceSkills)
  }

  describe("initial state", () => {
    it("shows header and subtitle", async () => {
      const wrapper = mountMarketplace()
      await flushPromises()

      expect(wrapper.find(".marketplace-header h2").text()).toBe("Marketplace")
      expect(wrapper.find(".subtitle").text()).toContain("GitHub")
    })

    it("shows search input and button", async () => {
      const wrapper = mountMarketplace()
      await flushPromises()

      expect(wrapper.find("input").exists()).toBe(true)
      expect(wrapper.find(".search-btn").exists()).toBe(true)
      expect(wrapper.find(".search-btn").text()).toContain("Search")
    })

    it("auto-searches on mount with empty query", async () => {
      mockAppService.SearchMarketplace.mockResolvedValue(mockResults)
      mountMarketplace()
      await flushPromises()

      expect(mockAppService.SearchMarketplace).toHaveBeenCalledWith(
        "github",
        "",
        1,
      )
    })

    it("shows results from auto-search", async () => {
      mockAppService.SearchMarketplace.mockResolvedValue(mockResults)
      const wrapper = mountMarketplace()
      await flushPromises()

      expect(wrapper.findAll(".repo-card")).toHaveLength(2)
    })
  })

  describe("search flow", () => {
    it("calls SearchMarketplace with provider, query, and page on click", async () => {
      const wrapper = mountMarketplace()
      await flushPromises()
      mockAppService.SearchMarketplace.mockClear()

      mockAppService.SearchMarketplace.mockResolvedValue(mockResults)
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
      const wrapper = mountMarketplace()
      await flushPromises()
      mockAppService.SearchMarketplace.mockClear()

      mockAppService.SearchMarketplace.mockResolvedValue(mockResults)
      await wrapper.find("input").setValue("review")
      await wrapper.find("input").trigger("keyup.enter")
      await flushPromises()

      expect(mockAppService.SearchMarketplace).toHaveBeenCalledWith(
        "github",
        "review",
        1,
      )
    })

    it("shows loading state while searching", async () => {
      const wrapper = mountMarketplace()
      await flushPromises()

      mockAppService.SearchMarketplace.mockReturnValue(new Promise(() => {}))
      await wrapper.find("input").setValue("test")
      await wrapper.find(".search-btn").trigger("click")

      expect(wrapper.find(".spinner-sm").exists()).toBe(true)
      expect(wrapper.find(".search-btn").text()).toContain("Searching")
      expect((wrapper.find("input").element as HTMLInputElement).disabled).toBe(
        true,
      )
    })

    it("shows error on search failure", async () => {
      const wrapper = mountMarketplace()
      await flushPromises()

      mockAppService.SearchMarketplace.mockRejectedValue(
        new Error("GitHub API rate limit exceeded"),
      )
      await wrapper.find("input").setValue("test")
      await wrapper.find(".search-btn").trigger("click")
      await flushPromises()

      expect(wrapper.find(".result-error").exists()).toBe(true)
      expect(wrapper.text()).toContain("rate limit")
    })

    it("dismisses error on close button click", async () => {
      const wrapper = mountMarketplace()
      await flushPromises()

      mockAppService.SearchMarketplace.mockRejectedValue(
        new Error("Network error"),
      )
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

    it("shows avatar image", async () => {
      const wrapper = await mountWithResults()

      const avatars = wrapper.findAll(".repo-avatar")
      expect(avatars).toHaveLength(2)
      expect(avatars[0].attributes("src")).toBe(
        "https://avatars.githubusercontent.com/u/alice",
      )
    })

    it("shows stars badge with star icon and formatting", async () => {
      const wrapper = await mountWithResults()

      const badges = wrapper.findAll(".badge-stars")
      expect(badges[0].text()).toBe("42")
      expect(badges[1].text()).toBe("1.5k")
      // Star icon SVG should be present
      expect(badges[0].find("svg").exists()).toBe(true)
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

    it("shows link button on each card", async () => {
      const wrapper = await mountWithResults()

      const linkBtns = wrapper.findAll(".btn-icon")
      expect(linkBtns).toHaveLength(2)
    })

    it("shows singular 'result' for count of 1", async () => {
      mockAppService.SearchMarketplace.mockResolvedValue({
        repos: [mockResults.repos[0]],
        totalCount: 1,
      })

      const wrapper = mountMarketplace()
      await flushPromises()

      expect(wrapper.find(".results-count").text()).toContain("1 result")
      expect(wrapper.find(".results-count").text()).not.toContain("1 results")
    })
  })

  describe("no results", () => {
    it("shows no results message when search returns empty", async () => {
      const wrapper = mountMarketplace()
      await flushPromises()
      mockAppService.SearchMarketplace.mockClear()

      mockAppService.SearchMarketplace.mockResolvedValue(emptyResults)
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
      await flushPromises()
      return wrapper
    }

    it("emits install-from-source with repo fullName on install click", async () => {
      const wrapper = await mountWithResults()
      const installBtns = wrapper.findAll(".btn-primary")
      await installBtns[0].trigger("click")

      expect(wrapper.emitted("install-from-source")).toEqual([["alice/skills"]])
    })

    it("emits correct fullName for different repos", async () => {
      const wrapper = await mountWithResults()
      const installBtns = wrapper.findAll(".btn-primary")
      await installBtns[1].trigger("click")

      expect(wrapper.emitted("install-from-source")).toEqual([["bob/tools"]])
    })

    it("does not call InstallFromSource directly", async () => {
      const wrapper = await mountWithResults()
      const installBtns = wrapper.findAll(".btn-primary")
      await installBtns[0].trigger("click")
      await flushPromises()

      expect(mockAppService.InstallFromSource).not.toHaveBeenCalled()
    })
  })

  describe("link button", () => {
    it("opens repo URL in browser on link button click", async () => {
      mockAppService.SearchMarketplace.mockResolvedValue(mockResults)
      const wrapper = mountMarketplace()
      await flushPromises()

      const linkBtns = wrapper.findAll(".btn-icon")
      await linkBtns[0].trigger("click")

      expect(mockBrowser.OpenURL).toHaveBeenCalledWith(
        "https://github.com/alice/skills",
      )
    })
  })

  describe("card click opens detail", () => {
    it("shows RepoReadmeModal on card click", async () => {
      mockAppService.SearchMarketplace.mockResolvedValue(mockResults)
      mockAppService.GetRepoReadme.mockResolvedValue("# README content")
      const wrapper = mountMarketplace()
      await flushPromises()

      const cards = wrapper.findAll(".repo-card")
      await cards[0].trigger("click")
      await flushPromises()

      expect(mockAppService.GetRepoReadme).toHaveBeenCalledWith(
        "alice",
        "skills",
      )
    })
  })

  describe("cleanup on unmount", () => {
    it("does not throw when unmounting", async () => {
      const wrapper = mountMarketplace()
      await flushPromises()
      expect(() => wrapper.unmount()).not.toThrow()
    })
  })
})
