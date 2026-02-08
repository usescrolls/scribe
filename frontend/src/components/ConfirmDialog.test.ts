import { describe, it, expect } from "vitest"
import { mount } from "@vue/test-utils"
import { nextTick } from "vue"
import ConfirmDialog from "./ConfirmDialog.vue"

describe("ConfirmDialog", () => {
  const defaultProps = {
    title: "Confirm Action",
    message: "Are you sure?",
    confirmLabel: "Confirm",
  }

  function mountDialog(props = {}) {
    return mount(ConfirmDialog, {
      props: { ...defaultProps, ...props },
      global: {
        stubs: {
          Teleport: true,
        },
      },
    })
  }

  describe("rendering", () => {
    it("renders title", () => {
      const wrapper = mountDialog()

      expect(wrapper.find(".confirm-header h3").text()).toBe("Confirm Action")
    })

    it("renders message", () => {
      const wrapper = mountDialog()

      expect(wrapper.find(".confirm-message").text()).toBe("Are you sure?")
    })

    it("renders confirm button with label", () => {
      const wrapper = mountDialog({ confirmLabel: "Delete" })

      expect(wrapper.find(".btn-confirm").text()).toBe("Delete")
    })

    it("renders cancel button", () => {
      const wrapper = mountDialog()

      expect(wrapper.find(".btn-cancel").text()).toBe("Cancel")
    })

    it("applies danger class when danger prop is true", () => {
      const wrapper = mountDialog({ danger: true })

      expect(wrapper.find(".btn-confirm").classes()).toContain("danger")
    })

    it("does not apply danger class by default", () => {
      const wrapper = mountDialog()

      expect(wrapper.find(".btn-confirm").classes()).not.toContain("danger")
    })

    it("renders backdrop", () => {
      const wrapper = mountDialog()

      expect(wrapper.find(".confirm-backdrop").exists()).toBe(true)
    })
  })

  describe("interactions", () => {
    it("hides dialog on confirm button click", async () => {
      const wrapper = mountDialog()

      await wrapper.find(".btn-confirm").trigger("click")
      await nextTick()

      expect(wrapper.find(".confirm-backdrop").exists()).toBe(false)
    })

    it("hides dialog on cancel button click", async () => {
      const wrapper = mountDialog()

      await wrapper.find(".btn-cancel").trigger("click")
      await nextTick()

      expect(wrapper.find(".confirm-backdrop").exists()).toBe(false)
    })

    it("hides dialog on backdrop click", async () => {
      const wrapper = mountDialog()

      await wrapper.find(".confirm-backdrop").trigger("click")
      await nextTick()

      expect(wrapper.find(".confirm-backdrop").exists()).toBe(false)
    })

    it("does not hide when clicking modal content", async () => {
      const wrapper = mountDialog()

      await wrapper.find(".confirm-modal").trigger("click")
      await nextTick()

      expect(wrapper.find(".confirm-backdrop").exists()).toBe(true)
    })

    it("hides dialog on Escape key", async () => {
      const wrapper = mountDialog()

      document.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape" }))
      await nextTick()

      expect(wrapper.find(".confirm-backdrop").exists()).toBe(false)
      wrapper.unmount()
    })
  })
})
