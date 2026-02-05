import { describe, it, expect } from "vitest"
import { mount } from "@vue/test-utils"
import CompleteStep from "./CompleteStep.vue"

describe("CompleteStep", () => {
  it("renders success title", () => {
    const wrapper = mount(CompleteStep)

    expect(wrapper.find("h1").text()).toBe("You're All Set!")
  })

  it("renders success icon", () => {
    const wrapper = mount(CompleteStep)

    expect(wrapper.find(".success-icon").exists()).toBe(true)
  })

  it("shows next steps section", () => {
    const wrapper = mount(CompleteStep)

    expect(wrapper.find(".next-steps").exists()).toBe(true)
    expect(wrapper.findAll("li").length).toBeGreaterThanOrEqual(3)
  })

  it("shows open button", () => {
    const wrapper = mount(CompleteStep)

    expect(wrapper.find(".primary-button").text()).toBe("Open Scribe")
  })

  it("emits finish on button click", async () => {
    const wrapper = mount(CompleteStep)

    await wrapper.find(".primary-button").trigger("click")

    expect(wrapper.emitted("finish")).toBeTruthy()
  })
})
