import { describe, it, expect, vi, beforeEach, afterEach } from "vitest"
import { mount } from "@vue/test-utils"
import ToastNotification from "./ToastNotification.vue"

describe("ToastNotification", () => {
  beforeEach(() => {
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  function mountToast(props: {
    message: string
    type?: "success" | "error" | "info"
    duration?: number
  }) {
    return mount(ToastNotification, { props })
  }

  it("renders message text", async () => {
    const wrapper = mountToast({ message: "Skill updated" })

    // Wait for requestAnimationFrame to show the toast
    await vi.advanceTimersByTimeAsync(16)

    expect(wrapper.text()).toContain("Skill updated")
  })

  it("applies success class for success type", async () => {
    const wrapper = mountToast({ message: "Done", type: "success" })
    await vi.advanceTimersByTimeAsync(16)

    expect(wrapper.find(".toast-success").exists()).toBe(true)
  })

  it("applies error class for error type", async () => {
    const wrapper = mountToast({ message: "Failed", type: "error" })
    await vi.advanceTimersByTimeAsync(16)

    expect(wrapper.find(".toast-error").exists()).toBe(true)
  })

  it("applies info class by default", async () => {
    const wrapper = mountToast({ message: "Info" })
    await vi.advanceTimersByTimeAsync(16)

    expect(wrapper.find(".toast-info").exists()).toBe(true)
  })

  it("emits close after duration", async () => {
    const wrapper = mountToast({ message: "Auto close", duration: 1000 })
    await vi.advanceTimersByTimeAsync(16)

    expect(wrapper.emitted("close")).toBeFalsy()

    // Advance past duration
    await vi.advanceTimersByTimeAsync(1000)
    // Plus the 200ms transition delay in close()
    await vi.advanceTimersByTimeAsync(200)

    expect(wrapper.emitted("close")).toBeTruthy()
  })

  it("emits close on click", async () => {
    const wrapper = mountToast({ message: "Click me" })
    await vi.advanceTimersByTimeAsync(16)

    await wrapper.find(".toast").trigger("click")
    // Wait for the 200ms transition delay
    await vi.advanceTimersByTimeAsync(200)

    expect(wrapper.emitted("close")).toBeTruthy()
  })

  it("uses default 4000ms duration", async () => {
    const wrapper = mountToast({ message: "Default duration" })
    await vi.advanceTimersByTimeAsync(16)

    // Not yet closed at 3s
    await vi.advanceTimersByTimeAsync(3000)
    expect(wrapper.emitted("close")).toBeFalsy()

    // Closed after 4s + 200ms transition
    await vi.advanceTimersByTimeAsync(1000)
    await vi.advanceTimersByTimeAsync(200)
    expect(wrapper.emitted("close")).toBeTruthy()
  })
})
