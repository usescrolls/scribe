import { describe, it, expect } from "vitest"
import { mount } from "@vue/test-utils"
import WelcomeStep from "./WelcomeStep.vue"

describe("WelcomeStep", () => {
  it("renders welcome title", () => {
    const wrapper = mount(WelcomeStep)

    expect(wrapper.find("h1").text()).toBe("Welcome to Scribe")
  })

  it("renders tagline", () => {
    const wrapper = mount(WelcomeStep)

    expect(wrapper.find(".tagline").text()).toContain("Sync AI coding skills")
  })

  it("renders get started button", () => {
    const wrapper = mount(WelcomeStep)

    expect(wrapper.find(".primary-button").text()).toBe("Get Started")
  })

  it("emits next on button click", async () => {
    const wrapper = mount(WelcomeStep)

    await wrapper.find(".primary-button").trigger("click")

    expect(wrapper.emitted("next")).toBeTruthy()
  })
})
