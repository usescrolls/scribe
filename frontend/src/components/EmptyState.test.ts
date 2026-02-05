import { describe, it, expect } from "vitest"
import { mount } from "@vue/test-utils"
import EmptyState from "./EmptyState.vue"

describe("EmptyState", () => {
  it("renders the title", () => {
    const wrapper = mount(EmptyState)

    expect(wrapper.find("h2").text()).toBe("No skills installed")
  })

  it("shows CLI instruction", () => {
    const wrapper = mount(EmptyState)

    expect(wrapper.find("code").text()).toBe("scribe install owner/repo")
  })

  it("shows description text", () => {
    const wrapper = mount(EmptyState)

    expect(wrapper.find("p").text()).toContain(
      "Install skills from useScrolls.com",
    )
  })
})
