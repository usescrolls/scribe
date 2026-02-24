import { describe, it, expect } from "vitest"
import { mount } from "@vue/test-utils"
import SourceAvatar from "./SourceAvatar.vue"

describe("SourceAvatar", () => {
  describe("rendering", () => {
    it("renders avatar image for public github source", () => {
      const wrapper = mount(SourceAvatar, {
        props: { source: "vercel-labs/skills", sourceType: "github" },
      })

      const img = wrapper.find(".source-avatar")
      expect(img.exists()).toBe(true)
      expect(img.attributes("src")).toBe(
        "https://github.com/vercel-labs.png?size=48",
      )
    })

    it("renders placeholder for private github source", () => {
      const wrapper = mount(SourceAvatar, {
        props: {
          source: "acme-corp/private-skills",
          sourceType: "github",
          isPrivate: true,
        },
      })

      expect(wrapper.find(".source-avatar").exists()).toBe(false)
      expect(wrapper.find(".source-avatar-placeholder").exists()).toBe(true)
    })

    it("renders placeholder for non-github source type", () => {
      const wrapper = mount(SourceAvatar, {
        props: { source: "local/path", sourceType: "local" },
      })

      expect(wrapper.find(".source-avatar").exists()).toBe(false)
      // Non-github sources don't render anything (no placeholder either)
      expect(wrapper.find(".source-avatar-placeholder").exists()).toBe(false)
    })

    it("renders placeholder for github source without owner/repo format", () => {
      const wrapper = mount(SourceAvatar, {
        props: { source: "unknown", sourceType: "github" },
      })

      expect(wrapper.find(".source-avatar").exists()).toBe(false)
      expect(wrapper.find(".source-avatar-placeholder").exists()).toBe(true)
    })

    it("extracts owner from nested source paths", () => {
      const wrapper = mount(SourceAvatar, {
        props: {
          source: "gitlab-group/subgroup/repo",
          sourceType: "github",
        },
      })

      const img = wrapper.find(".source-avatar")
      expect(img.exists()).toBe(true)
      expect(img.attributes("src")).toBe(
        "https://github.com/gitlab-group.png?size=48",
      )
    })

    it("defaults isPrivate to false when not provided", () => {
      const wrapper = mount(SourceAvatar, {
        props: { source: "owner/repo", sourceType: "github" },
      })

      expect(wrapper.find(".source-avatar").exists()).toBe(true)
    })
  })

  describe("error handling", () => {
    it("shows placeholder after image load error", async () => {
      const wrapper = mount(SourceAvatar, {
        props: { source: "owner/repo", sourceType: "github" },
      })

      expect(wrapper.find(".source-avatar").exists()).toBe(true)

      await wrapper.find(".source-avatar").trigger("error")

      expect(wrapper.find(".source-avatar").exists()).toBe(false)
      expect(wrapper.find(".source-avatar-placeholder").exists()).toBe(true)
    })

    it("resets failed state when source changes", async () => {
      const wrapper = mount(SourceAvatar, {
        props: { source: "owner/repo", sourceType: "github" },
      })

      // Trigger error
      await wrapper.find(".source-avatar").trigger("error")
      expect(wrapper.find(".source-avatar").exists()).toBe(false)

      // Change source
      await wrapper.setProps({ source: "other/repo" })

      // Should attempt to load the new image
      expect(wrapper.find(".source-avatar").exists()).toBe(true)
      expect(wrapper.find(".source-avatar").attributes("src")).toBe(
        "https://github.com/other.png?size=48",
      )
    })
  })
})
