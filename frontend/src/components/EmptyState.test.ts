import { describe, it, expect } from "vitest"
import { mount } from "@vue/test-utils"
import EmptyState from "./EmptyState.vue"

describe("EmptyState", () => {
  it("renders the title", () => {
    const wrapper = mount(EmptyState)

    expect(wrapper.find("h2").text()).toBe("No skills in this workspace")
  })

  it("shows description text", () => {
    const wrapper = mount(EmptyState)

    expect(wrapper.find("p").text()).toContain(
      "Add skills from the Browse All tab",
    )
  })
})
